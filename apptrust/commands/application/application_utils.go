package application

import (
	"fmt"
	"strconv"

	"github.com/jfrog/jfrog-cli-application/apptrust/commands"
	"github.com/jfrog/jfrog-cli-application/apptrust/commands/utils"
	"github.com/jfrog/jfrog-cli-application/apptrust/model"
	"github.com/jfrog/jfrog-cli-core/v2/plugins/components"
)

func populateApplicationFromFlags(ctx *components.Context, descriptor *model.AppDescriptor) error {
	descriptor.ApplicationName = ctx.GetStringFlagValue(commands.ApplicationNameFlag)

	if ctx.IsFlagSet(commands.DescriptionFlag) {
		description := ctx.GetStringFlagValue(commands.DescriptionFlag)
		descriptor.Description = &description
	}

	if ctx.IsFlagSet(commands.BusinessCriticalityFlag) {
		businessCriticalityStr := ctx.GetStringFlagValue(commands.BusinessCriticalityFlag)
		businessCriticality, err := utils.ValidateEnumFlag(
			commands.BusinessCriticalityFlag,
			businessCriticalityStr,
			model.BusinessCriticalityUnspecified,
			model.BusinessCriticalityValues)
		if err != nil {
			return err
		}
		descriptor.BusinessCriticality = &businessCriticality
	}

	if ctx.IsFlagSet(commands.MaturityLevelFlag) {
		maturityLevelStr := ctx.GetStringFlagValue(commands.MaturityLevelFlag)
		maturityLevel, err := utils.ValidateEnumFlag(
			commands.MaturityLevelFlag,
			maturityLevelStr,
			model.MaturityLevelUnspecified,
			model.MaturityLevelValues)
		if err != nil {
			return err
		}
		descriptor.MaturityLevel = &maturityLevel
	}

	if ctx.IsFlagSet(commands.LabelsFlag) {
		labels, err := utils.ParseLabelKeyValuePairs(ctx.GetStringFlagValue(commands.LabelsFlag))
		if err != nil {
			return fmt.Errorf("failed to parse --%s: %w", commands.LabelsFlag, err)
		}
		descriptor.Labels = &labels
	}

	// Only set LabelUpdates if at least one of add-labels or remove-labels flags is set
	if ctx.IsFlagSet(commands.AddLabelsFlag) || ctx.IsFlagSet(commands.RemoveLabelsFlag) {
		labelUpdates := &model.LabelUpdates{}

		if ctx.IsFlagSet(commands.AddLabelsFlag) {
			addLabels, err := utils.ParseLabelKeyValuePairs(ctx.GetStringFlagValue(commands.AddLabelsFlag))
			if err != nil {
				return fmt.Errorf("failed to parse --%s: %w", commands.AddLabelsFlag, err)
			}
			labelUpdates.Add = addLabels
		}

		if ctx.IsFlagSet(commands.RemoveLabelsFlag) {
			removeLabels, err := utils.ParseLabelKeyValuePairs(ctx.GetStringFlagValue(commands.RemoveLabelsFlag))
			if err != nil {
				return fmt.Errorf("failed to parse --%s: %w", commands.RemoveLabelsFlag, err)
			}
			labelUpdates.Remove = removeLabels
		}

		descriptor.LabelUpdates = labelUpdates
	}

	if ctx.IsFlagSet(commands.UserOwnersFlag) {
		userOwners := utils.ParseSliceFlag(ctx.GetStringFlagValue(commands.UserOwnersFlag))
		descriptor.UserOwners = &userOwners
	}

	if ctx.IsFlagSet(commands.GroupOwnersFlag) {
		groupOwners := utils.ParseSliceFlag(ctx.GetStringFlagValue(commands.GroupOwnersFlag))
		descriptor.GroupOwners = &groupOwners
	}

	if err := populateMonitorPolicyFromFlags(ctx, descriptor); err != nil {
		return err
	}

	return nil
}

func populateMonitorPolicyFromFlags(ctx *components.Context, descriptor *model.AppDescriptor) error {
	if !ctx.IsFlagSet(commands.MonitorPolicyFlag) {
		return nil
	}

	fields, err := utils.ParseKeyValueString(ctx.GetStringFlagValue(commands.MonitorPolicyFlag), ",")
	if err != nil {
		return fmt.Errorf("failed to parse --%s: %w", commands.MonitorPolicyFlag, err)
	}

	for key := range fields {
		if key != "type" && key != "value" {
			return fmt.Errorf("invalid field '%s' in --%s (supported fields: type, value)", key, commands.MonitorPolicyFlag)
		}
	}

	typeStr, typeSet := fields["type"]
	if !typeSet {
		return fmt.Errorf("--%s requires a 'type' field", commands.MonitorPolicyFlag)
	}

	policyType, err := utils.ValidateEnumFlag(commands.MonitorPolicyFlag, typeStr, model.MonitorPolicyTypeNone, model.MonitorPolicyTypeValues)
	if err != nil {
		return err
	}

	policy := &model.MonitorPolicy{Type: policyType}
	valueStr, valueSet := fields["value"]

	if policyType == model.MonitorPolicyTypeNone {
		if valueSet {
			return fmt.Errorf("'value' must not be set in --%s when type is '%s'", commands.MonitorPolicyFlag, model.MonitorPolicyTypeNone)
		}
	} else {
		if !valueSet {
			return fmt.Errorf("'value' is required in --%s when type is '%s'", commands.MonitorPolicyFlag, policyType)
		}
		value, err := strconv.Atoi(valueStr)
		if err != nil {
			return fmt.Errorf("'value' in --%s must be an integer: %w", commands.MonitorPolicyFlag, err)
		}
		if value <= 0 {
			return fmt.Errorf("'value' in --%s must be a positive integer", commands.MonitorPolicyFlag)
		}
		policy.Value = &value
	}

	descriptor.MonitorPolicy = policy
	return nil
}
