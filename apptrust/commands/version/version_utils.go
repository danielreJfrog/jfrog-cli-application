package version

import (
	"strings"

	"github.com/jfrog/jfrog-cli-application/apptrust/commands"
	"github.com/jfrog/jfrog-cli-application/apptrust/commands/utils"
	"github.com/jfrog/jfrog-cli-application/apptrust/model"
	"github.com/jfrog/jfrog-cli-core/v2/plugins/components"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
)

// BuildPromotionParams extracts common promotion parameters from command context
// Used by both promote and release commands
func BuildPromotionParams(ctx *components.Context) (string, []string, []string, error) {
	var includedRepos []string
	var excludedRepos []string

	if includeReposStr := ctx.GetStringFlagValue(commands.IncludeReposFlag); includeReposStr != "" {
		includedRepos = utils.ParseSliceFlag(includeReposStr)
	}

	if excludeReposStr := ctx.GetStringFlagValue(commands.ExcludeReposFlag); excludeReposStr != "" {
		excludedRepos = utils.ParseSliceFlag(excludeReposStr)
	}

	promotionType := ctx.GetStringFlagValue(commands.PromotionTypeFlag)

	validatedPromotionType, err := utils.ValidateEnumFlag(commands.PromotionTypeFlag, promotionType, model.PromotionTypeCopy, model.PromotionTypeValues)
	if err != nil {
		return "", nil, nil, err
	}

	// If dry-run is true, override with dry_run
	dryRun := ctx.GetBoolFlagValue(commands.DryRunFlag)
	if dryRun {
		validatedPromotionType = model.PromotionTypeDryRun
	}

	return validatedPromotionType, includedRepos, excludedRepos, nil
}

// ParseArtifactProps extracts artifact properties from command context
func ParseArtifactProps(ctx *components.Context) ([]model.ArtifactProperty, error) {
	if propsStr := ctx.GetStringFlagValue(commands.PropsFlag); propsStr != "" {
		props, err := utils.ParseListPropertiesFlag(propsStr)
		if err != nil {
			return nil, errorutils.CheckErrorf("failed to parse properties: %s", err.Error())
		}

		var artifactProps []model.ArtifactProperty
		for key, values := range props {
			artifactProps = append(artifactProps, model.ArtifactProperty{
				Key:    key,
				Values: values,
			})
		}
		return artifactProps, nil
	}
	return nil, nil
}

// ParseOverwriteStrategy extracts and validates the overwrite strategy from command context
func ParseOverwriteStrategy(ctx *components.Context) (string, error) {
	overwriteStrategy := ctx.GetStringFlagValue(commands.OverwriteStrategyFlag)
	if overwriteStrategy == "" {
		return "", nil
	}

	validatedStrategy, err := utils.ValidateEnumFlag(commands.OverwriteStrategyFlag, overwriteStrategy, "", model.OverwriteStrategyValues)
	if err != nil {
		return "", err
	}

	// Convert to uppercase for API request
	return strings.ToUpper(validatedStrategy), nil
}

// ParsePathMappings extracts path mapping rules from the --path-mapping flag.
// Format: "input=(.*), output=stable-release/$1[, package-type=.*]; input=(...), output=..."
// Returns nil if flag is not provided.
func ParsePathMappings(ctx *components.Context) (*model.PromotionModifications, error) {
	const (
		inputField       = "input"
		outputField      = "output"
		packageTypeField = "package-type"
	)

	flagValue := ctx.GetStringFlagValue(commands.PathMappingFlag)
	if flagValue == "" {
		return nil, nil
	}

	entries := utils.ParseSliceFlag(flagValue)
	var mappings []model.PromotionPathMapping

	for i, entry := range entries {
		if entry == "" {
			return nil, errorutils.CheckErrorf("--%s entry %d is empty", commands.PathMappingFlag, i+1)
		}

		entryMap, err := utils.ParseKeyValueString(entry, ",")
		if err != nil {
			return nil, errorutils.CheckErrorf("--%s entry %d: %s", commands.PathMappingFlag, i+1, err.Error())
		}

		input, hasInput := entryMap[inputField]
		output, hasOutput := entryMap[outputField]

		if !hasInput || input == "" {
			return nil, errorutils.CheckErrorf("--%s entry %d: '%s' is required", commands.PathMappingFlag, i+1, inputField)
		}
		if !hasOutput || output == "" {
			return nil, errorutils.CheckErrorf("--%s entry %d: '%s' is required", commands.PathMappingFlag, i+1, outputField)
		}

		mapping := model.PromotionPathMapping{
			Input:  input,
			Output: output,
		}
		if pt, ok := entryMap[packageTypeField]; ok {
			mapping.PackageType = pt
		}

		mappings = append(mappings, mapping)
	}

	return &model.PromotionModifications{Mappings: mappings}, nil
}
