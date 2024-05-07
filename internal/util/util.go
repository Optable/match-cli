package util

import (
	"bufio"
	"context"
	"io"
	"strings"

	v1 "github.com/optable/match-api/match/v1"
)

var validIdentifiersPrefix = map[string]string{
	"emails":       "e:",
	"phoneNumbers": "p:",
	"ipv4S":        "i4:",
	"ipv6S":        "i6:",
	"appleIdfas":   "a:",
	"googleGaids":  "g:",
	"rokuRidas":    "r:",
	"samsungTifas": "s:",
	"amazonAfais":  "f:",
	"netids":       "n:",
	"postalCodes":  "z:",
	"id5s":         "id5:",
	"utiqs":        "utiq:",
}

// GetInputChannel reads identifiers from a file to a channel
func GetInputChannel(ctx context.Context, uniqueIdentifiersInFile map[string]bool) (<-chan []byte, error) {
	// make the output channel
	identifiers := make(chan []byte)
	go func() {
		defer close(identifiers)
		for identifier := range uniqueIdentifiersInFile {
			// push to channel
			identifiers <- []byte(identifier)
		}
	}()

	return identifiers, nil
}

// returns insights for the identifiers in the file
func GetInsights(uniqueIdentifiersInFile map[string]bool) *v1.Insights {
	var insight v1.Insights
	for identifier := range uniqueIdentifiersInFile {
		switch {
		case strings.HasPrefix(identifier, validIdentifiersPrefix["emails"]):
			insight.Emails++
		case strings.HasPrefix(identifier, validIdentifiersPrefix["phoneNumbers"]):
			insight.PhoneNumbers++
		case strings.HasPrefix(identifier, validIdentifiersPrefix["ipv4S"]):
			insight.Ipv4S++
		case strings.HasPrefix(identifier, validIdentifiersPrefix["ipv6S"]):
			insight.Ipv6S++
		case strings.HasPrefix(identifier, validIdentifiersPrefix["appleIdfas"]):
			insight.AppleIdfas++
		case strings.HasPrefix(identifier, validIdentifiersPrefix["googleGaids"]):
			insight.GoogleGaids++
		case strings.HasPrefix(identifier, validIdentifiersPrefix["rokuRidas"]):
			insight.RokuRidas++
		case strings.HasPrefix(identifier, validIdentifiersPrefix["samsungTifas"]):
			insight.SamsungTifas++
		case strings.HasPrefix(identifier, validIdentifiersPrefix["amazonAfais"]):
			insight.AmazonAfais++
		case strings.HasPrefix(identifier, validIdentifiersPrefix["netids"]):
			insight.Netids++
		case strings.HasPrefix(identifier, validIdentifiersPrefix["postalCodes"]):
			insight.PostalCodes++
		case strings.HasPrefix(identifier, validIdentifiersPrefix["id5s"]):
			insight.Id5S++
		case strings.HasPrefix(identifier, validIdentifiersPrefix["utiqs"]):
			insight.Utiqs++
		}
	}
	return &insight
}

// returns unique identifiers in the file
func GetUniqueIdentifiersInFile(r io.Reader) (map[string]bool, error) {
	uniqueIdentifiersInFile := make(map[string]bool)
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		element := string(scanner.Bytes())
		for _, validPrefix := range validIdentifiersPrefix {
			if strings.HasPrefix(element, validPrefix) {
				uniqueIdentifiersInFile[element] = true
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return uniqueIdentifiersInFile, nil
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
		case v1.IdKind_ID_KIND_NETID:
			result.Insights.Netids = clamp(srcInsight.Netids, threshold(result.Insights.Netids, result.Insights.DifferentialPrivacyThreshold))
		case v1.IdKind_ID_KIND_POSTAL_CODE:
			result.Insights.PostalCodes = clamp(srcInsight.PostalCodes, threshold(result.Insights.PostalCodes, result.Insights.DifferentialPrivacyThreshold))
		case v1.IdKind_ID_KIND_ID5:
			result.Insights.Id5S = clamp(srcInsight.Id5S, threshold(result.Insights.Id5S, result.Insights.DifferentialPrivacyThreshold))
		case v1.IdKind_ID_KIND_UTIQ:
			result.Insights.Utiqs = clamp(srcInsight.Utiqs, threshold(result.Insights.Utiqs, result.Insights.DifferentialPrivacyThreshold))
		}
	}
}
