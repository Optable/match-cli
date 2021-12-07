package util

import (
	"bufio"
	"context"
	"io"
	"strings"

	v1 "github.com/optable/match-api/match/v1"
)

//GetInputChannel reads identifiers from a file to a channel
func GenInputChannel(ctx context.Context, f io.Reader) (int64, *v1.Insights, <-chan []byte, error) {
	n, elementsInFile, insight, err := count(f)
	if err != nil {
		return n, nil, nil, err
	}

	// make the output channel
	identifiers := make(chan []byte)

	// wrap f in a bufio reader
	go func() {
		defer close(identifiers)
		for identifier := range elementsInFile {
			// push to channel
			identifiers <- []byte(identifier)
		}
	}()

	return n, insight, identifiers, nil
}

// count returns number of lines in file, as well as the
// number of each id type
func count(r io.Reader) (int64, map[string]bool, *v1.Insights, error) {
	var nUnique int64
	var insight v1.Insights
	visited := make(map[string]bool)
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		s := string(scanner.Bytes())
		if _, found := visited[s]; found {
			continue
		}
		// ignoring the invalid types
		switch {
		case strings.HasPrefix(s, "e:"):
			insight.Emails++
		case strings.HasPrefix(s, "p:"):
			insight.PhoneNumbers++
		case strings.HasPrefix(s, "i4:"):
			insight.Ipv4S++
		case strings.HasPrefix(s, "i6:"):
			insight.Ipv6S++
		case strings.HasPrefix(s, "a:"):
			insight.AppleIdfas++
		case strings.HasPrefix(s, "g:"):
			insight.GoogleGaids++
		case strings.HasPrefix(s, "r:"):
			insight.RokuRidas++
		case strings.HasPrefix(s, "s:"):
			insight.SamsungTifas++
		case strings.HasPrefix(s, "f:"):
			insight.AmazonAfais++
		default:
			continue
		}
		visited[s] = true
		nUnique++
	}

	if err := scanner.Err(); err != nil {
		return nUnique, nil, nil, err
	}

	return nUnique, visited, &insight, nil
}

// clamp changes the received numbers from the partner which can have differential privacy noise in them,
// meaning the numbers could be negative or exceed the total number of IDs.
// We want to normalize the number to be 0 <= candidate <= maxValue
func clamp(max, n int64) int64 {
	if n < 0 {
		return 0
	}
	if n > max {
		return max
	}
	return n
}

func threshold(n int64, threshold int32) int64 {
	// no thresholding
	if int64(threshold) == 0 {
		return n
	}
	if n < int64(threshold) {
		return 0
	}
	return n
}

// ThresholdAndClampMatchResult modifies the received match result insight numbers
// by applying a threshold on the received value first, if the value is less than the threshold,
// it will be set to 0. Afterwards, we clamp the thresholded value.
func ThresholdAndClampMatchResult(result *v1.ExternalMatchResult, srcInsight *v1.Insights) {
	for idkind := range v1.IdKind_name {
		switch v1.IdKind(idkind) {
		case v1.IdKind_ID_KIND_EMAIL_HASH:
			result.Insights.Emails = clamp(srcInsight.Emails, threshold(result.Insights.Emails, result.Insights.DifferentialPrivacyThreshold))
		case v1.IdKind_ID_KIND_PHONE_NUMBER:
			result.Insights.PhoneNumbers = clamp(srcInsight.PhoneNumbers, threshold(result.Insights.PhoneNumbers, result.Insights.DifferentialPrivacyThreshold))
		case v1.IdKind_ID_KIND_IPV4:
			result.Insights.Ipv4S = clamp(srcInsight.Ipv4S, threshold(result.Insights.Ipv4S, result.Insights.DifferentialPrivacyThreshold))
		case v1.IdKind_ID_KIND_IPV6:
			result.Insights.Ipv6S = clamp(srcInsight.Ipv6S, threshold(result.Insights.Ipv6S, result.Insights.DifferentialPrivacyThreshold))
		case v1.IdKind_ID_KIND_APPLE_IDFA:
			result.Insights.AppleIdfas = clamp(srcInsight.AppleIdfas, threshold(result.Insights.AppleIdfas, result.Insights.DifferentialPrivacyThreshold))
		case v1.IdKind_ID_KIND_GOOGLE_GAID:
			result.Insights.GoogleGaids = clamp(srcInsight.GoogleGaids, threshold(result.Insights.GoogleGaids, result.Insights.DifferentialPrivacyThreshold))
		case v1.IdKind_ID_KIND_ROKU_RIDA:
			result.Insights.RokuRidas = clamp(srcInsight.RokuRidas, threshold(result.Insights.RokuRidas, result.Insights.DifferentialPrivacyThreshold))
		case v1.IdKind_ID_KIND_SAMSUNG_TIFA:
			result.Insights.SamsungTifas = clamp(srcInsight.SamsungTifas, threshold(result.Insights.SamsungTifas, result.Insights.DifferentialPrivacyThreshold))
		case v1.IdKind_ID_KIND_AMAZON_AFAI:
			result.Insights.AmazonAfais = clamp(srcInsight.AmazonAfais, threshold(result.Insights.AmazonAfais, result.Insights.DifferentialPrivacyThreshold))
		}
	}
}
