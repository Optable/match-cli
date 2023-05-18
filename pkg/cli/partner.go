package cli

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"regexp"

	v1 "github.com/optable/match-api/match/v1"

	"google.golang.org/protobuf/encoding/protojson"
)

type (
	PartnerGetCmd struct {
		Name string `arg:"" required:"" help:"Name of the partner."`
	}

	PartnerListCmd struct {
	}

	PartnerConnectCmd struct {
		Name  string `arg:"" required:"" help:"Name of the partner."`
		Token string `arg:"" required:"" help:"The invite token from the partner."`
		OrganizationName string `hidden:"" help:"Organization Name must be between 2 and 64 characters."`
		UniqueNodeId string `hidden:"" help:"Unique Node ID must start with an alphabet, contains only alphanumerical(lowercase) and dashes, and be between 3 to 64 characters."`
	}

	PartnerCmd struct {
		Connect PartnerConnectCmd `cmd:"" help:"Connect to a partner sandbox with an invite token."`
		List    PartnerListCmd    `cmd:"" help:"List partners."`
		Get     PartnerGetCmd     `cmd:"" help:"Get partner."`
	}
)

func (p *PartnerGetCmd) Run(cli *CliContext) error {
	partner := cli.config.findPartner(p.Name)
	if partner == nil {
		return fmt.Errorf("partner %s does not exist", p.Name)
	}

	return printJson(partner)
}

func (p *PartnerListCmd) Run(cli *CliContext) error {
	for _, partner := range cli.config.Partners {
		if err := printJson(partner); err != nil {
			return err
		}
	}
	return nil
}

func (p *PartnerConnectCmd) Run(cli *CliContext) error {
	existingPartner := cli.config.findPartner(p.Name)
	if existingPartner != nil {
		return fmt.Errorf("a partner with name %s already exists", p.Name)
	}

	if len(p.OrganizationName) < 2 || 64 < len(p.OrganizationName) {
		return fmt.Errorf("organization name must be between 2 and 64 characters")
	}

	if err := validateNodeId(p.UniqueNodeId); err != nil {
		return err
	}

	token, err := decodeToken(p.Token)
	if err != nil {
		return fmt.Errorf("failed to decode token: %w", err)
	}

	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return fmt.Errorf("failed to generate key pair : %w", err)
	}
	marshaledPrivateKey, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		return fmt.Errorf("failed to marshal private key : %w", err)
	}

	publicKey := privateKey.Public()
	marshaledPublicKey, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return fmt.Errorf("failed to marshal public key: %w", err)
	}

	conf := PartnerConfig{
		Name:       p.Name,
		URL:        token.SandboxInfo,
		PublicKey:  base64.StdEncoding.EncodeToString(marshaledPublicKey),
		PrivateKey: base64.StdEncoding.EncodeToString(marshaledPrivateKey),
	}

	client, err := conf.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	headlessPartner := 	&v1.HeadlessPartner{
		PublicKey: conf.PublicKey,
		Token:    p.Token,
	}

	if p.OrganizationName != "" {
		headlessPartner.OrganizationName = p.OrganizationName
	}

	if p.UniqueNodeId != "" {
		headlessPartner.NodeId = p.UniqueNodeId
	}

	err = client.RegisterPartner(cli.ctx, &v1.RegisterPartnerReq{
		PartnerInfo: &v1.RegisterPartnerReq_HeadlessPartner{
			HeadlessPartner: headlessPartner,
		},
	})

	if err != nil {
		return fmt.Errorf("failed to register with partner: %w", err)
	}

	cli.config.Partners = append(cli.config.Partners, conf)
	err = cli.SaveConfig()
	if err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	return printJson(conf)
}

func decodeToken(token string) (*v1.PartnerInitToken, error) {
	json, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return nil, err
	}

	var message v1.PartnerInitToken
	if err := protojson.Unmarshal(json, &message); err != nil {
		return nil, err
	}
	return &message, nil
}

func validateNodeId(nodeId string) error {
	isAlphabet := regexp.MustCompile(`^[a-z]`).MatchString
	if !isAlphabet(nodeId[0:1]) {
		return fmt.Errorf("node id must start with an alphabet")
	}

	isAplhanumericOrDash := regexp.MustCompile(`^[a-z0-9-]+$`).MatchString
	if !isAplhanumericOrDash(nodeId) {
		return fmt.Errorf("node id must contain only alphanumerical(lowercase) and dashes")
	}

	if len(nodeId) < 3 || 64 < len(nodeId) {
		return fmt.Errorf("node id must be between 2 and 64 characters")
	}
	
	return nil
}
