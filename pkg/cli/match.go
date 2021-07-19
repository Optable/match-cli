package cli

import (
	"context"
	"crypto/tls"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	v1 "github.com/optable/match-cli/api/v1"
	"github.com/optable/match-cli/internal/auth"
	matchclient "github.com/optable/match-cli/internal/client"
	"github.com/optable/match-cli/internal/util"

	"github.com/segmentio/ksuid"
	"google.golang.org/protobuf/proto"
)

type (
	MatchCreateCmd struct {
		Partner string `arg:"" required:"" help:"Name of the partner"`
		Name    string `arg:"" required:"" help:"Name of the match"`
	}

	MatchListCmd struct {
		Partner string `arg:"" required:"" help:"Name of the partner"`
	}

	MatchGetResultsCmd struct {
		Partner string `arg:"" required:"" help:"Name of the partner"`
		MatchId string `arg:"" required:"" help:"ID of the match"`
	}

	MatchRunCmd struct {
		Partner     string        `arg:"" required:"" help:"Name of the partner"`
		InitTimeout time.Duration `default:"1m" help:"Timeout for the initialization of the match"`
		RunTimeout  time.Duration `default:"1h" help:"Timeout for the match operation"`
		MatchID     string        `arg:"" required:"" help:"ID of the match"`
		File        *os.File      `arg:"" required:"" help:"File to match"`
	}

	MatchCmd struct {
		Create     MatchCreateCmd     `cmd:"" help:"Create a match"`
		List       MatchListCmd       `cmd:"" help:"List matches"`
		GetResults MatchGetResultsCmd `cmd:"" help:"Get a match results"`
		Run        MatchRunCmd        `cmd:"" help:"Run a match"`
	}
)

type matchResult struct {
	Time     time.Time    `json:"time"`
	Id       string       `json:"id"`
	State    string       `json:"state"`
	ErrorMsg string       `json:"error_msg,omitempty"`
	Results  *v1.Insights `json:"results,omitempty"`
}

func matchResultStateFromProto(state v1.ExternalMatchResultState) string {
	switch state {
	case v1.ExternalMatchResultState_EXTERNAL_MATCH_RESULT_STATE_UNKNOWN:
		return "unknown"
	case v1.ExternalMatchResultState_EXTERNAL_MATCH_RESULT_STATE_PENDING:
		return "pending"
	case v1.ExternalMatchResultState_EXTERNAL_MATCH_RESULT_STATE_COMPLETED:
		return "completed"
	case v1.ExternalMatchResultState_EXTERNAL_MATCH_RESULT_STATE_ERRORED:
		return "errored"
	default:
		return "unknown"
	}
}

func matchResultFromProto(resultpb *v1.ExternalMatchResult) *matchResult {
	result := &matchResult{
		Time:     resultpb.UpdatedAt.AsTime(),
		Id:       resultpb.Uuid,
		State:    matchResultStateFromProto(resultpb.State),
		ErrorMsg: resultpb.ErrorMsg,
	}

	if resultpb.Insights != nil {
		result.Results = proto.Clone(resultpb.Insights).(*v1.Insights)
		result.Results.ComputedAt = nil
	}

	return result
}

func (m *MatchCreateCmd) Run(cli *CliContext) error {
	partner := cli.config.findPartner(m.Partner)
	if partner == nil {
		return fmt.Errorf("partner %s does not exist", m.Partner)
	}

	client, err := partner.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	req := &v1.CreateExternalMatchReq{
		MatchUuid: ksuid.New().String(),
		Name:      m.Name,
		RefreshFrequency: &v1.CreateExternalMatchReq_Adhoc{
			Adhoc: &v1.ExternalMatchRefreshAdhoc{},
		},
	}

	res, err := client.CreateMatch(cli.ctx, req)
	if err != nil {
		return err
	}

	return printJson(res)
}

func (m *MatchGetResultsCmd) Run(cli *CliContext) error {
	partner := cli.config.findPartner(m.Partner)
	if partner == nil {
		return fmt.Errorf("partner %s does not exist", m.Partner)
	}

	client, err := partner.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	req := &v1.GetExternalMatchResultsReq{
		MatchUuid: m.MatchId,
	}

	res, err := client.GetMatchResults(cli.ctx, req)
	if err != nil {
		return err
	}

	sort.SliceStable(res.Results, func(i, j int) bool {
		return res.Results[i].UpdatedAt.AsTime().After(res.Results[j].UpdatedAt.AsTime())
	})

	for _, result := range res.Results {
		if err := printJson(matchResultFromProto(result)); err != nil {
			return err
		}
	}
	return nil
}

func (m *MatchListCmd) Run(cli *CliContext) error {
	partner := cli.config.findPartner(m.Partner)
	if partner == nil {
		return fmt.Errorf("partner %s does not exist", m.Partner)
	}

	client, err := partner.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	req := &v1.ListExternalMatchReq{}
	res, err := client.ListMatches(cli.ctx, req)
	if err != nil {
		return err
	}

	for _, match := range res.Matches {
		if err := printJson(match); err != nil {
			return err
		}
	}
	return nil
}

func getTLSConfig(cert *auth.EphemerealCertificate, peerCertPem, hostport string) (*tls.Config, error) {
	tlsCertificate, err := cert.GetTLSCertificate()
	if err != nil {
		return nil, fmt.Errorf("failed to get TLS certificate from ephemereal certificate: %w", err)
	}

	pinnedCert, err := auth.ParseCertificatePEM(peerCertPem)
	if err != nil {
		return nil, fmt.Errorf("failed to parse peer pinned certificate: %w", err)
	}

	return &tls.Config{
		Certificates: []tls.Certificate{tlsCertificate},
		// We skip verification and validate that the received certificate
		// is stricly equal to the expected one with VerifyPeerCertificate
		InsecureSkipVerify:    true,
		ServerName:            strings.Split(hostport, ":")[0],
		ClientAuth:            tls.RequireAnyClientCert,
		VerifyPeerCertificate: auth.MakeVerifyPinnedCertificate(pinnedCert),
	}, nil
}

func pollRunMatch(ctx context.Context, partner *PartnerConfig, matchUUID string, cert *auth.EphemerealCertificate) (*v1.RunExternalMatchRes, error) {
	matchResultUUID := ksuid.New().String()
	info(ctx).Msgf("generated match result id %s", matchResultUUID)

	client, err := partner.NewClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}
	for {
		info(ctx).Msgf("still polling /match/run to get match endpoint")
		res, err := client.RunMatch(ctx, &v1.RunExternalMatchReq{
			MatchUuid:            matchUUID,
			MatchResultUuid:      matchResultUUID,
			ClientCertificatePem: string(cert.CertificatePem),
		})
		if err != nil {
			return nil, err
		}
		if err == nil && res.Endpoint != "" {
			info(ctx).Msgf("got match endpoint %s", res.Endpoint)
			return res, nil
		}
		debug(ctx).Msg("match endpoint not ready, sleeping for 5 seconds")
		time.Sleep(5 * time.Second)
	}
}

func pollGetMatchResult(ctx context.Context, partner *PartnerConfig, matchResultUUID string) (*v1.ExternalMatchResult, error) {
	// Need to create new client because of token expiry
	client, err := partner.NewClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	for {
		info(ctx).Msgf("still polling /match/get-result for results")
		res, err := client.GetResult(ctx, &v1.GetExternalMatchResultReq{MatchResultUuid: matchResultUUID})
		if err != nil {
			return nil, err
		}
		if res.GetMatchResult().GetState() != v1.ExternalMatchResultState_EXTERNAL_MATCH_RESULT_STATE_PENDING {
			return res.MatchResult, nil
		}
		debug(ctx).Msg("results not ready, sleeping for 5 seconds")
		time.Sleep(10 * time.Second)
	}
}

func (m *MatchRunCmd) Run(cli *CliContext) error {
	defer m.File.Close()
	ctx := withInfoLogger(cli.ctx)

	ctx, cancel := context.WithTimeout(ctx, m.RunTimeout)
	defer cancel()
	info(ctx).Msgf("running match %s with a timeout of %v", m.MatchID, m.RunTimeout)

	n, records, err := util.GenInputChannel(ctx, m.File)
	if err != nil {
		return fmt.Errorf("failed to load record file %s : %w", m.File.Name(), err)
	}
	info(ctx).Msgf("loaded %d records from %s", n, m.File.Name())

	partner := cli.config.findPartner(m.Partner)
	if partner == nil {
		return fmt.Errorf("partner %s does not exist", m.Partner)
	}

	key, err := partner.ParsedPrivateKey()
	if err != nil {
		return fmt.Errorf("failed to parse private key for partner %s: %w", m.Partner, err)
	}

	ephemerealCertificate, err := auth.NewEphemerealCertificate(key)
	if err != nil {
		return fmt.Errorf("failed to create ephemereal certificate: %w", err)
	}
	debug(ctx).Msg("Generated ephemereal certificate for tls authentication")

	info(ctx).Msgf("polling /match/run with a timeout of %v to get match endpoint", m.InitTimeout)
	runMatchCtx, runMatchCancel := context.WithTimeout(ctx, m.InitTimeout)
	runMatchRes, err := pollRunMatch(runMatchCtx, partner, m.MatchID, ephemerealCertificate)
	runMatchCancel()
	if err != nil {
		return fmt.Errorf("failed while polling run/match: %w", err)
	}

	info(ctx).Msgf("running dhpsi protocol on %s", runMatchRes.Endpoint)
	tlsConfig, err := getTLSConfig(ephemerealCertificate, runMatchRes.ServerCertificatePem, runMatchRes.Endpoint)
	if err != nil {
		return fmt.Errorf("failed to create TLS config for dhpsi protocol: %w", err)
	}

	if err = matchclient.Send(ctx, runMatchRes.Endpoint, tlsConfig, n, records); err != nil {
		return fmt.Errorf("failed to run DHPSI: %w", err)
	}
	info(ctx).Msg("successfully completed dhpsi protocol")

	info(ctx).Msgf("polling /match/get-result for results")
	result, err := pollGetMatchResult(ctx, partner, runMatchRes.MatchResultUuid)
	if err != nil {
		return fmt.Errorf("failed to poll /match/get-result: %w", err)
	}

	if result.State == v1.ExternalMatchResultState_EXTERNAL_MATCH_RESULT_STATE_ERRORED {
		return fmt.Errorf("got an errored state from /match/get-result: %s", result.ErrorMsg)
	}

	info(ctx).Msg("got results from /match/get-result")

	return printJson(matchResultFromProto(result))
}
