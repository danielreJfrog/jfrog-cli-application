package version

import (
	"testing"

	"github.com/jfrog/jfrog-cli-application/apptrust/commands"
	"github.com/jfrog/jfrog-cli-application/apptrust/model"
	"github.com/jfrog/jfrog-cli-core/v2/plugins/components"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseOverwriteStrategy(t *testing.T) {
	tests := []struct {
		name          string
		flagValue     string
		expectError   bool
		expectedValue string
	}{
		{
			name:          "valid value - disabled",
			flagValue:     "disabled",
			expectError:   false,
			expectedValue: "DISABLED",
		},
		{
			name:          "valid value - latest",
			flagValue:     "latest",
			expectError:   false,
			expectedValue: "LATEST",
		},
		{
			name:          "valid value - all",
			flagValue:     "all",
			expectError:   false,
			expectedValue: "ALL",
		},
		{
			name:          "empty value",
			flagValue:     "",
			expectError:   false,
			expectedValue: "",
		},
		{
			name:        "invalid value",
			flagValue:   "INVALID",
			expectError: true,
		},
		{
			name:        "uppercase value",
			flagValue:   "DISABLED",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &components.Context{}
			if tt.flagValue != "" {
				ctx.AddStringFlag(commands.OverwriteStrategyFlag, tt.flagValue)
			}

			result, err := ParseOverwriteStrategy(ctx)

			if tt.expectError {
				assert.Error(t, err, "ParseOverwriteStrategy(%q) expected error, got nil", tt.flagValue)
				return
			}

			assert.NoError(t, err, "ParseOverwriteStrategy(%q) unexpected error: %v", tt.flagValue, err)
			assert.Equal(t, tt.expectedValue, result, "ParseOverwriteStrategy(%q) = %v, want %v", tt.flagValue, result, tt.expectedValue)
		})
	}
}

func TestBuildPromotionParams(t *testing.T) {
	tests := []struct {
		name                  string
		promotionType         string
		dryRun                bool
		includeRepos          string
		excludeRepos          string
		expectedPromotionType string
		expectedIncludeRepos  []string
		expectedExcludeRepos  []string
		expectError           bool
	}{
		{
			name:                  "default promotion type (copy)",
			promotionType:         "",
			dryRun:                false,
			expectedPromotionType: model.PromotionTypeCopy,
			expectedIncludeRepos:  []string(nil),
			expectedExcludeRepos:  []string(nil),
			expectError:           false,
		},
		{
			name:                  "promotion type move",
			promotionType:         "move",
			dryRun:                false,
			expectedPromotionType: model.PromotionTypeMove,
			expectedIncludeRepos:  []string(nil),
			expectedExcludeRepos:  []string(nil),
			expectError:           false,
		},
		{
			name:                  "dry run overrides promotion type",
			promotionType:         "copy",
			dryRun:                true,
			expectedPromotionType: model.PromotionTypeDryRun,
			expectedIncludeRepos:  []string(nil),
			expectedExcludeRepos:  []string(nil),
			expectError:           false,
		},
		{
			name:                  "with include and exclude repos",
			promotionType:         "copy",
			dryRun:                false,
			includeRepos:          "repo1;repo2",
			excludeRepos:          "repo3;repo4",
			expectedPromotionType: model.PromotionTypeCopy,
			expectedIncludeRepos:  []string{"repo1", "repo2"},
			expectedExcludeRepos:  []string{"repo3", "repo4"},
			expectError:           false,
		},
		{
			name:          "invalid promotion type",
			promotionType: "invalid",
			dryRun:        false,
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &components.Context{}

			// Set flag values using AddStringFlag
			if tt.promotionType != "" {
				ctx.AddStringFlag(commands.PromotionTypeFlag, tt.promotionType)
			}
			if tt.dryRun {
				ctx.AddBoolFlag(commands.DryRunFlag, tt.dryRun)
			}
			if tt.includeRepos != "" {
				ctx.AddStringFlag(commands.IncludeReposFlag, tt.includeRepos)
			}
			if tt.excludeRepos != "" {
				ctx.AddStringFlag(commands.ExcludeReposFlag, tt.excludeRepos)
			}

			promotionType, includeRepos, excludeRepos, err := BuildPromotionParams(ctx)

			if tt.expectError {
				assert.Error(t, err, "BuildPromotionParams expected error, got nil")
				return
			}

			assert.NoError(t, err, "BuildPromotionParams unexpected error: %v", err)
			assert.Equal(t, tt.expectedPromotionType, promotionType, "promotion type mismatch")
			assert.Equal(t, tt.expectedIncludeRepos, includeRepos, "include repos mismatch")
			assert.Equal(t, tt.expectedExcludeRepos, excludeRepos, "exclude repos mismatch")
		})
	}
}

func TestParsePathMappings(t *testing.T) {
	tests := []struct {
		name        string
		mapIn       string
		mapOut      string
		mapType     string
		expected    *model.PromotionModifications
		expectError bool
		errContains string
	}{
		{
			name:     "no flags - returns nil",
			expected: nil,
		},
		{
			name:   "single mapping without package type",
			mapIn:  "(.*)",
			mapOut: "stable-release/$1",
			expected: &model.PromotionModifications{
				Mappings: []model.PromotionPathMapping{
					{Input: "(.*)", Output: "stable-release/$1"},
				},
			},
		},
		{
			name:    "single mapping with package type",
			mapIn:   "(.*)",
			mapOut:  "stable-release/$1",
			mapType: ".*",
			expected: &model.PromotionModifications{
				Mappings: []model.PromotionPathMapping{
					{PackageType: ".*", Input: "(.*)", Output: "stable-release/$1"},
				},
			},
		},
		{
			name:    "multiple mappings",
			mapIn:   "(.*);(.*\\.jar)",
			mapOut:  "release/$1;jars/$1",
			mapType: ".*;maven",
			expected: &model.PromotionModifications{
				Mappings: []model.PromotionPathMapping{
					{PackageType: ".*", Input: "(.*)", Output: "release/$1"},
					{PackageType: "maven", Input: "(.*\\.jar)", Output: "jars/$1"},
				},
			},
		},
		{
			name:    "fewer package types than inputs - partial assignment",
			mapIn:   "(.*);(.*\\.jar)",
			mapOut:  "release/$1;jars/$1",
			mapType: "maven",
			expected: &model.PromotionModifications{
				Mappings: []model.PromotionPathMapping{
					{PackageType: "maven", Input: "(.*)", Output: "release/$1"},
					{Input: "(.*\\.jar)", Output: "jars/$1"},
				},
			},
		},
		{
			name:        "map-in without map-out - error",
			mapIn:       "(.*)",
			expectError: true,
			errContains: "must be provided together",
		},
		{
			name:        "map-out without map-in - error",
			mapOut:      "target/$1",
			expectError: true,
			errContains: "must be provided together",
		},
		{
			name:        "mismatched count - error",
			mapIn:       "(.*);(.*\\.jar)",
			mapOut:      "release/$1",
			expectError: true,
			errContains: "same number of entries",
		},
		{
			name:        "more package types than inputs - error",
			mapIn:       "(.*)",
			mapOut:      "release/$1",
			mapType:     "maven;npm",
			expectError: true,
			errContains: "more entries",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &components.Context{}
			if tt.mapIn != "" {
				ctx.AddStringFlag(commands.MapInFlag, tt.mapIn)
			}
			if tt.mapOut != "" {
				ctx.AddStringFlag(commands.MapOutFlag, tt.mapOut)
			}
			if tt.mapType != "" {
				ctx.AddStringFlag(commands.MapTypeFlag, tt.mapType)
			}

			result, err := ParsePathMappings(ctx)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}
