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

func SupportedVersion(sentinelVersion, minimumVersion string) bool {
	return sentinelVersion == LatestSentinelVersion ||
		semver.Compare(minimumVersion, sentinelVersion) <= 0
}

func UnsupportedVersion(sentinelVersion, minimumVersion string) bool {
	return !SupportedVersion(sentinelVersion, minimumVersion)
}

func ValidateSentinelVersion(sentinelVersion string) (bool, string) {
	if sentinelVersion == LatestSentinelVersion || sentinelVersion == "" {
		return true, LatestSentinelVersion
	}

	return semver.IsValid(sentinelVersion), sentinelVersion
}
