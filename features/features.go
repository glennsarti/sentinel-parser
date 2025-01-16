package features

import "golang.org/x/mod/semver"

const (
	LatestSentinelVersion = "latest"
)

type SentinelFeatures struct {
	// Language features
	FuncDeclarations   bool
	IsDefinedFunctions bool
	// Configuration file features
	BadStdlibImportName      bool
	ConfigurationOverride    bool
	OptionalEnforcementLevel bool
	ParamsInPolicy           bool
	SentinelBlock            bool
	V2ImportBlock            bool
}

// Sentinel Language feature versions
const (
	FuncDeclarationsMinimumVerions   = "v0.20.0"
	IsDefinedFunctionsMinimumVerions = "v0.21.0"
)

// Sentinel Configuration Language feature versions
const (
	BadStdlibImportNameMinimumVersion      = "v0.19.0"
	ConfigurationOverrideMinimumVersion    = "v0.19.0"
	OptionalEnforcementLevelMinimumVersion = "v0.18.11"
	ParamsInPolicyMinimumVersion           = "v0.21.0"
	SentinelBlockMinimumVersion            = "v0.22.0"
	V2ImportBlockMinimumVersion            = "v0.19.0"
)

func SupportedFeatures(sentinelversion string) SentinelFeatures {
	return SentinelFeatures{
		// Language features
		IsDefinedFunctions: SupportedVersion(sentinelversion, IsDefinedFunctionsMinimumVerions),
		FuncDeclarations:   SupportedVersion(sentinelversion, FuncDeclarationsMinimumVerions),
		// Configuration file features
		BadStdlibImportName:      SupportedVersion(sentinelversion, BadStdlibImportNameMinimumVersion),
		ConfigurationOverride:    SupportedVersion(sentinelversion, ConfigurationOverrideMinimumVersion),
		OptionalEnforcementLevel: SupportedVersion(sentinelversion, OptionalEnforcementLevelMinimumVersion),
		ParamsInPolicy:           SupportedVersion(sentinelversion, ParamsInPolicyMinimumVersion),
		SentinelBlock:            SupportedVersion(sentinelversion, SentinelBlockMinimumVersion),
		V2ImportBlock:            SupportedVersion(sentinelversion, V2ImportBlockMinimumVersion),
	}
}

// Whether the Sentinel version is the same or later than another version, typically
// the version when a feature was introduced.
func SupportedVersion(sentinelVersion, minimumVersion string) bool {
	actual := sentinelVersion
	if actual == LatestSentinelVersion || actual == "" {
		actual = SentinelVersions[0]
	}

	if !semver.IsValid(minimumVersion) || !semver.IsValid(actual) {
		return false
	}

	return semver.Compare(minimumVersion, actual) <= 0
}

// Whether the Sentinel version is earlier than another version, typically
// the version when a feature was introduced.
func UnsupportedVersion(sentinelVersion, minimumVersion string) bool {
	actual := sentinelVersion
	if actual == LatestSentinelVersion || actual == "" {
		actual = SentinelVersions[0]
	}

	if !semver.IsValid(minimumVersion) || !semver.IsValid(actual) {
		return false
	}

	return semver.Compare(minimumVersion, actual) > 0
}

// Validates a Semver string is valid, and resolves to an existing Sentinel version
func ValidateSentinelVersion(sentinelVersion string) (bool, string) {
	if sentinelVersion == LatestSentinelVersion || sentinelVersion == "" {
		return true, SentinelVersions[0]
	}

	if !semver.IsValid(sentinelVersion) {
		return false, sentinelVersion
	}

	for _, ver := range SentinelVersions {
		if semver.Compare(ver, sentinelVersion) == 0 {
			return true, ver
		}
	}

	return false, sentinelVersion
}
