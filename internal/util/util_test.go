package util

import (
	"context"
	"strings"
	"testing"

	v1 "github.com/optable/match-api/match/v1"
)

var input = `i4:8.8.8.8
p:18055554321
i4:1.1.1.1
i6:1.1.1.1.1.1
p:12125551122
p:12125551122
e:920d0b248f5eea3b9c4838867d8dc8392e8522f2f89f7dc67a3f0e3d52ba2c14
e:920d0b248f5eea3b9c4838867d8dc8392e8522f2f89f7dc67a3f0e3d52ba2c14
e:920d1212465e48d839b47102826b8c574959e5fcc6bf0fe4f888811a6d14c8de
a:214as2d4asasdasd
e:920d43ae6aebac63291f0476a63f9dc3d3cd7d3b071673c7f145f58e893740f4
r:4as6d4a3s4dasdad
g:a2354ds35as4d3asd
g:5a4d35a4d35as4d3a
f:21312230udklsjfaklhjda
s:alhjklashsjklfahs23e0923ur420`

func TestCount(t *testing.T) {
	n, elementsInFile, insight, err := count(strings.NewReader("e:920d43ae6aebac63291f0476a63f9dc3d3cd7d3b071673c7f145f58e893740f4\ns:alhjklashsjklfahs23e0923ur420"))
	if err != nil {
		t.Fatal(err)
	}

	identifier1 := "e:920d43ae6aebac63291f0476a63f9dc3d3cd7d3b071673c7f145f58e893740f4"
	if val, found := elementsInFile[string(identifier1)]; !found || !val {
		t.Fatal("count failed")
	}

	identifier2 := "s:alhjklashsjklfahs23e0923ur420"
	if val, found := elementsInFile[string(identifier2)]; !found || !val {
		t.Fatal("count failed")
	}

	if n != 2 || insight.Emails != 1 || insight.SamsungTifas != 1 {
		t.Fatalf("count failed")
	}

	n, _, insight, err = count(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if n != 14 || insight.Emails != 3 || insight.Ipv4S != 2 || insight.Ipv6S != 1 ||
		insight.PhoneNumbers != 2 || insight.AppleIdfas != 1 || insight.SamsungTifas != 1 ||
		insight.GoogleGaids != 2 || insight.RokuRidas != 1 || insight.AmazonAfais != 1 {
		t.Fatalf("count failed")
	}
}

func TestClamp(t *testing.T) {
	max := int64(10)
	clamped := clamp(max, -1)
	if clamped != 0 {
		t.Fatalf("clamp failed, want %d, got %d", 0, clamped)
	}

	clamped = clamp(max, 11)
	if clamped != max {
		t.Fatalf("clamp failed, want %d, got %d", max, clamped)
	}

	clamped = clamp(max, 1)
	if clamped != 1 {
		t.Fatalf("clamp failed, want %d, got %d", 1, clamped)
	}
}

func TestClampMatchResult(t *testing.T) {
	_, _, srcInsight, err := count(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}

	received := v1.ExternalMatchResult{Insights: &v1.Insights{}}

	received.Insights.Emails = 5
	received.Insights.PhoneNumbers = -2
	received.Insights.SamsungTifas = 2
	received.Insights.Ipv4S = 0
	received.Insights.DifferentialPrivacyThreshold = 0

	ThresholdAndClampMatchResult(&received, srcInsight)
	if received.Insights.Emails != srcInsight.Emails || received.Insights.PhoneNumbers != 0 ||
		received.Insights.SamsungTifas != srcInsight.SamsungTifas || received.Insights.Ipv4S != 0 {
		t.Fatal("clamp result failed")
	}
}

func TestClampAndThresholdMatchResult(t *testing.T) {
	_, _, srcInsight, err := count(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}

	received := v1.ExternalMatchResult{Insights: &v1.Insights{}}

	received.Insights.Emails = 5
	received.Insights.PhoneNumbers = -2
	received.Insights.SamsungTifas = 2
	received.Insights.Ipv4S = 0
	received.Insights.DifferentialPrivacyThreshold = 2

	ThresholdAndClampMatchResult(&received, srcInsight)
	if received.Insights.Emails != srcInsight.Emails || received.Insights.Ipv4S != 0 ||
		received.Insights.Ipv6S != 0 || received.Insights.PhoneNumbers != 0 ||
		received.Insights.AppleIdfas != 0 || received.Insights.SamsungTifas != srcInsight.SamsungTifas ||
		received.Insights.GoogleGaids != 0 || received.Insights.RokuRidas != 0 || received.Insights.AmazonAfais != 0 {
		t.Fatal("clamp result failed")
	}

	src := v1.Insights{}
	src.Emails = 1001
	src.PhoneNumbers = 570
	src.GoogleGaids = 2

	received.Insights.Emails = 600
	received.Insights.PhoneNumbers = 500
	received.Insights.GoogleGaids = -2
	received.Insights.DifferentialPrivacyThreshold = 600
	t.Log(received.Insights.PhoneNumbers)

	ThresholdAndClampMatchResult(&received, &src)
	t.Log(received.Insights.PhoneNumbers)
	if received.Insights.Emails != 600 || received.Insights.PhoneNumbers != 0 ||
		received.Insights.GoogleGaids != 0 {
		t.Fatal("clamp result failed")
	}
}

func TestGetInputChannel(t *testing.T) {
	inputData := strings.NewReader(input)
	n, insight, records, err := GenInputChannel(context.Background(), inputData)
	if err != nil {
		t.Fatalf("failed creating input channel from temporary input Data for testing: %s", err)
	}

	expectedOutput := map[string]bool{
		"i4:8.8.8.8":     true,
		"p:18055554321":  true,
		"i4:1.1.1.1":     true,
		"i6:1.1.1.1.1.1": true,
		"p:12125551122":  true,
		"e:920d0b248f5eea3b9c4838867d8dc8392e8522f2f89f7dc67a3f0e3d52ba2c14": true,
		"e:920d1212465e48d839b47102826b8c574959e5fcc6bf0fe4f888811a6d14c8de": true,
		"a:214as2d4asasdasd": true,
		"e:920d43ae6aebac63291f0476a63f9dc3d3cd7d3b071673c7f145f58e893740f4": true,
		"r:4as6d4a3s4dasdad":              true,
		"g:a2354ds35as4d3asd":             true,
		"g:5a4d35a4d35as4d3a":             true,
		"f:21312230udklsjfaklhjda":        true,
		"s:alhjklashsjklfahs23e0923ur420": true,
	}

	chanLength := int64(0)

	for identifier := range records {
		if _, ok := expectedOutput[string(identifier)]; !ok {
			t.Fatal("unexpected output")
		}
		chanLength++
	}

	if n != 14 || insight.Emails != 3 || insight.Ipv4S != 2 || insight.Ipv6S != 1 ||
		insight.PhoneNumbers != 2 || insight.AppleIdfas != 1 || insight.SamsungTifas != 1 ||
		insight.GoogleGaids != 2 || insight.RokuRidas != 1 || insight.AmazonAfais != 1 || n != chanLength {
		t.Fatal("get input channel failed")
	}

}
