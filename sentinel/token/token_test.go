package token

import (
	"strings"
	"testing"

	"github.com/glennsarti/sentinel-parser/features"
)

func TestSentinelToken(t *testing.T) {
	minVersion := "0.0.0.0"

	for name, vtt := range allKeywords {
		t.Run(name, func(t *testing.T) {
			// Just checking that the static data is accurate compared with the allKeywords map
			t.Run("LookupIdent", func(t *testing.T) {
				// Latest version
				assertLookupIdent(name, features.LatestSentinelVersion, vtt.TokenType, t)

				if vtt.MinimumVersion == "" {
					// There's no minimum version so it should always return vtt.TokenType
					assertLookupIdent(name, minVersion, vtt.TokenType, t)
				} else {
					assertLookupIdent(name, minVersion, IDENT, t)
				}
			})

			// Just checking that the static data is accurate compared with the allKeywords map
			t.Run("SingleWordKeyword", func(t *testing.T) {
				noSpace := !strings.Contains(name, " ")

				// Latest version
				assertSingleWordKeyword(vtt.TokenType, features.LatestSentinelVersion, noSpace, t)

				if vtt.MinimumVersion == "" {
					// There's no minimum version so it should always return whether it's a single word
					assertSingleWordKeyword(vtt.TokenType, minVersion, noSpace, t)
				} else {
					assertSingleWordKeyword(vtt.TokenType, minVersion, false, t)
				}
			})
		})
	}
}

func assertLookupIdent(ident, sentinelVersion string, expected TokenType, t *testing.T) {
	actual := LookupIdent(sentinelVersion, ident)
	if actual != expected {
		t.Fatalf("Expected LookupIdent to return %q for version %q, but got %q", expected, sentinelVersion, actual)
	}
}

func assertSingleWordKeyword(tt TokenType, sentinelVersion string, expected bool, t *testing.T) {
	actual := SingleWordKeyword(sentinelVersion, tt)
	if actual != expected {
		t.Fatalf("Expected SingleWordKeyword to return %t for version %q, but got %t", expected, sentinelVersion, actual)
	}
}
