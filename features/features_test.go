package features

import (
	"testing"
)

func TestFeatureValid(t *testing.T) {
	// All of the Sentinel versions should be valid
	for _, ver := range SentinelVersions {
		ok, actual := ValidateSentinelVersion(ver)
		expected := ver
		if !ok {
			t.Fatalf("Expected sentinel version %q would validate but got false", ver)
		}
		if expected != actual {
			t.Fatalf("Expected sentinel version %q would validate as %q but got %q", ver, expected, actual)
		}
	}

	// Special keyword version should be valid
	latestVersions := []string{"", LatestSentinelVersion}
	for _, ver := range latestVersions {
		ok, actual := ValidateSentinelVersion(ver)
		expected := SentinelVersions[0]
		if !ok {
			t.Fatalf("Expected sentinel version %q would validate but got false", ver)
		}
		if actual != expected {
			t.Fatalf("Expected sentinel version %q would validate as %q but got %q", ver, expected, actual)
		}
	}

	// Bad version numbers should fail
	badVersions := []string{
		// Malformed
		"a.b.c",
		"0.0.1",
		"v0.2.0abc",
		// Valid but was never a sentinel version
		"v99.99.99",
	}
	for _, ver := range badVersions {
		ok, actual := ValidateSentinelVersion(ver)
		expected := ver
		if ok {
			t.Fatalf("Expected sentinel version %q would not validate but got true", ver)
		}
		if actual != expected {
			t.Fatalf("Expected sentinel version %q would validate as %q but got %q", ver, expected, actual)
		}
	}
}

func TestFeatureSupported(t *testing.T) {
	trueTestcases := []struct {
		ver, minVer string
	}{
		{ver: "", minVer: "v0.0.1"},
		{ver: LatestSentinelVersion, minVer: "v0.0.1"},
		{ver: "v0.0.2", minVer: "v0.0.1"},
	}
	for _, testcase := range trueTestcases {
		actual := SupportedVersion(testcase.ver, testcase.minVer)
		if !actual {
			t.Fatalf("Expected sentinel version %q with minimum version %q would return true but got false",
				testcase.ver, testcase.minVer)
		}
	}

	falseTestcases := []struct {
		ver, minVer string
	}{
		// Minimum versions must be a valid semver
		{ver: "v0.0.1", minVer: ""},
		{ver: "v0.0.1", minVer: "latest"},
		// Sentinel version must be a valid semver
		{ver: "0.0.2", minVer: "v0.0.1"},
		// Latest doesn't always resolve as true and still needs to be greater than the minimum version
		{ver: LatestSentinelVersion, minVer: "v999999.99.99"},
	}
	for _, testcase := range falseTestcases {
		actual := SupportedVersion(testcase.ver, testcase.minVer)
		if actual {
			t.Fatalf("Expected sentinel version %q with minimum version %q would return false but got true",
				testcase.ver, testcase.minVer)
		}
	}
}

func TestFeatureUnsupported(t *testing.T) {
	trueTestcases := []struct {
		ver, minVer string
	}{
		{ver: "v0.0.1", minVer: "v0.0.2"},
		// Latest doesn't always resolve as false and still needs to be less than the minimum version
		{ver: LatestSentinelVersion, minVer: "v999999.99.99"},
	}
	for _, testcase := range trueTestcases {
		actual := UnsupportedVersion(testcase.ver, testcase.minVer)
		if !actual {
			t.Fatalf("Expected sentinel version %q with minimum version %q would return true but got false",
				testcase.ver, testcase.minVer)
		}
	}

	falseTestcases := []struct {
		ver, minVer string
	}{
		// Minimum versions must be a valid semver
		{ver: "v0.0.1", minVer: ""},
		{ver: "v0.0.1", minVer: "latest"},
		// Sentinel version must be a valid semver
		{ver: "0.0.2", minVer: "v0.0.1"},
		// Actual testcases
		{ver: "v0.0.1", minVer: "v0.0.1"},
		{ver: "v0.0.2", minVer: "v0.0.1"},
	}
	for _, testcase := range falseTestcases {
		actual := UnsupportedVersion(testcase.ver, testcase.minVer)
		if actual {
			t.Fatalf("Expected sentinel version %q with minimum version %q would return false but got true",
				testcase.ver, testcase.minVer)
		}
	}
}
