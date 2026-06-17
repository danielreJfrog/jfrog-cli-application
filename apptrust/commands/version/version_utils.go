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

// ParsePathMappings extracts path mapping rules from --map-in, --map-out, and --map-type flags.
// Returns nil if no mapping flags are provided.
func ParsePathMappings(ctx *components.Context) (*model.PromotionModifications, error) {
	mapInStr := ctx.GetStringFlagValue(commands.MapInFlag)
	mapOutStr := ctx.GetStringFlagValue(commands.MapOutFlag)
	mapTypeStr := ctx.GetStringFlagValue(commands.MapTypeFlag)

	if mapInStr == "" && mapOutStr == "" && mapTypeStr == "" {
		return nil, nil
	}

	if mapInStr == "" || mapOutStr == "" {
		return nil, errorutils.CheckErrorf("--%s and --%s must be provided together (both are required for path mappings)",
			commands.MapInFlag, commands.MapOutFlag)
	}

	inputs := utils.ParseSliceFlag(mapInStr)
	outputs := utils.ParseSliceFlag(mapOutStr)

	for i, v := range inputs {
		if v == "" {
			return nil, errorutils.CheckErrorf("--%s entry %d is empty", commands.MapInFlag, i+1)
		}
	}
	for i, v := range outputs {
		if v == "" {
			return nil, errorutils.CheckErrorf("--%s entry %d is empty", commands.MapOutFlag, i+1)
		}
	}

	if len(inputs) != len(outputs) {
		return nil, errorutils.CheckErrorf("--%s and --%s must have the same number of entries (got %d and %d)",
			commands.MapInFlag, commands.MapOutFlag, len(inputs), len(outputs))
	}

	var packageTypes []string
	if mapTypeStr != "" {
		packageTypes = utils.ParseSliceFlag(mapTypeStr)
		for i, v := range packageTypes {
			if v == "" {
				return nil, errorutils.CheckErrorf("--%s entry %d is empty", commands.MapTypeFlag, i+1)
			}
		}
		if len(packageTypes) > len(inputs) {
			return nil, errorutils.CheckErrorf("--%s has more entries (%d) than --%s (%d)",
				commands.MapTypeFlag, len(packageTypes), commands.MapInFlag, len(inputs))
		}
	}

	mappings := make([]model.PromotionPathMapping, len(inputs))
	for i := range inputs {
		mappings[i] = model.PromotionPathMapping{
			Input:  inputs[i],
			Output: outputs[i],
		}
		if i < len(packageTypes) {
			mappings[i].PackageType = packageTypes[i]
		}
	}

	return &model.PromotionModifications{Mappings: mappings}, nil
}
