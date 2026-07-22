//go:build e2e

package e2e

import (
	"strings"
	"testing"

	"github.com/jfrog/jfrog-cli-application/apptrust/model"
	"github.com/jfrog/jfrog-cli-application/e2e/utils"
	"github.com/stretchr/testify/assert"
)

func TestCreateApp(t *testing.T) {
	projectKey := utils.GetTestProjectKey(t)
	appKey := utils.GenerateUniqueKey("app-create")
	appName := "Full Test Application"
	description := "Application with all fields populated"
	businessCriticality := "critical"
	maturityLevel := "production"
	userOwners := []string{"admin", "developer"}
	groupOwners := []string{"devops-team", "security-team"}

	err := utils.AppTrustCli.Exec("app-create", appKey,
		"--project="+projectKey,
		"--application-name="+appName,
		"--desc="+description,
		"--business-criticality="+businessCriticality,
		"--maturity-level="+maturityLevel,
		"--labels=env=prod;team=devops",
		"--user-owners="+strings.Join(userOwners, ";"),
		"--group-owners="+strings.Join(groupOwners, ";"))
	assert.NoError(t, err)

	// Fetch and verify the application was created correctly
	app, _, err := utils.GetApplication(appKey)
	assert.NoError(t, err)
	assert.Equal(t, appKey, app.ApplicationKey)
	assert.Equal(t, appName, app.ApplicationName)
	assert.Equal(t, projectKey, app.ProjectKey)
	assert.Equal(t, description, *app.Description)
	assert.Equal(t, businessCriticality, *app.BusinessCriticality)
	assert.Equal(t, maturityLevel, *app.MaturityLevel)
	assert.ElementsMatch(t, []model.LabelEntry{{Key: "env", Value: "prod"}, {Key: "team", Value: "devops"}}, *app.Labels)
	assert.Equal(t, userOwners, *app.UserOwners)
	assert.Equal(t, groupOwners, *app.GroupOwners)

	utils.DeleteApplication(t, appKey)
}

func TestUpdateApp(t *testing.T) {
	projectKey := utils.GetTestProjectKey(t)
	appKey := utils.GenerateUniqueKey("app-update")

	utils.CreateBasicApplication(t, appKey)

	updatedAppName := "Updated Test Application"
	updatedDescription := "Updated description"
	updatedBusinessCriticality := "high"
	updatedMaturityLevel := "production"
	updatedUserOwners := []string{"app-admin", "frog"}
	updatedGroupOwners := []string{"dev-team", "security-team"}

	err := utils.AppTrustCli.Exec("app-update", appKey,
		"--application-name="+updatedAppName,
		"--desc="+updatedDescription,
		"--business-criticality="+updatedBusinessCriticality,
		"--maturity-level="+updatedMaturityLevel,
		"--labels=env=qa;team=dev",
		"--user-owners="+strings.Join(updatedUserOwners, ";"),
		"--group-owners="+strings.Join(updatedGroupOwners, ";"))
	assert.NoError(t, err)

	// Fetch and verify the application was updated correctly
	app, _, err := utils.GetApplication(appKey)
	assert.NoError(t, err)
	assert.Equal(t, appKey, app.ApplicationKey)
	assert.Equal(t, updatedAppName, app.ApplicationName)
	assert.Equal(t, projectKey, app.ProjectKey)
	assert.Equal(t, updatedDescription, *app.Description)
	assert.Equal(t, updatedBusinessCriticality, *app.BusinessCriticality)
	assert.Equal(t, updatedMaturityLevel, *app.MaturityLevel)
	assert.ElementsMatch(t, []model.LabelEntry{{Key: "env", Value: "qa"}, {Key: "team", Value: "dev"}}, *app.Labels)
	assert.Equal(t, updatedUserOwners, *app.UserOwners)
	assert.Equal(t, updatedGroupOwners, *app.GroupOwners)

	utils.DeleteApplication(t, appKey)
}

func TestCreateAppWithMonitorPolicy(t *testing.T) {
	projectKey := utils.GetTestProjectKey(t)

	tests := []struct {
		name          string
		monitorPolicy string
		expectedType  string
		expectedValue *int
	}{
		{
			name:          "version count",
			monitorPolicy: "type=version_count, value=5",
			expectedType:  model.MonitorPolicyTypeVersionCount,
			expectedValue: intPtr(5),
		},
		{
			name:          "time frame in months",
			monitorPolicy: "type=time_frame_in_months, value=3",
			expectedType:  model.MonitorPolicyTypeTimeframe,
			expectedValue: intPtr(3),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appKey := utils.GenerateUniqueKey("app-create-monitor-" + strings.ReplaceAll(tt.name, " ", "-"))

			err := utils.AppTrustCli.Exec("app-create", appKey,
				"--project="+projectKey,
				"--application-name="+appKey,
				"--monitor-policy="+tt.monitorPolicy)
			assert.NoError(t, err)

			app, _, err := utils.GetApplication(appKey)
			assert.NoError(t, err)
			assert.Equal(t, appKey, app.ApplicationKey)
			if assert.NotNil(t, app.MonitorPolicy) {
				assert.Equal(t, tt.expectedType, app.MonitorPolicy.Type)
				assert.Equal(t, tt.expectedValue, app.MonitorPolicy.Value)
			}

			utils.DeleteApplication(t, appKey)
		})
	}
}

func TestUpdateAppMonitorPolicy(t *testing.T) {
	tests := []struct {
		name          string
		monitorPolicy string
		expectedType  string
		expectedValue *int
	}{
		{
			name:          "time frame in months",
			monitorPolicy: "type=time_frame_in_months, value=6",
			expectedType:  model.MonitorPolicyTypeTimeframe,
			expectedValue: intPtr(6),
		},
		{
			name:          "version count",
			monitorPolicy: "type=version_count, value=4",
			expectedType:  model.MonitorPolicyTypeVersionCount,
			expectedValue: intPtr(4),
		},
		{
			name:          "none",
			monitorPolicy: "type=none",
			expectedType:  model.MonitorPolicyTypeNone,
			expectedValue: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appKey := utils.GenerateUniqueKey("app-update-monitor-" + strings.ReplaceAll(tt.name, " ", "-"))

			utils.CreateBasicApplication(t, appKey)

			err := utils.AppTrustCli.Exec("app-update", appKey,
				"--monitor-policy="+tt.monitorPolicy)
			assert.NoError(t, err)

			app, _, err := utils.GetApplication(appKey)
			assert.NoError(t, err)
			assert.Equal(t, appKey, app.ApplicationKey)
			if assert.NotNil(t, app.MonitorPolicy) {
				assert.Equal(t, tt.expectedType, app.MonitorPolicy.Type)
				assert.Equal(t, tt.expectedValue, app.MonitorPolicy.Value)
			}

			utils.DeleteApplication(t, appKey)
		})
	}
}

func intPtr(i int) *int {
	return &i
}

func TestDeleteApp(t *testing.T) {
	appKey := utils.GenerateUniqueKey("app-delete")
	utils.CreateBasicApplication(t, appKey)

	// Verify the application exists
	app, _, err := utils.GetApplication(appKey)
	assert.NoError(t, err)
	assert.Equal(t, appKey, app.ApplicationKey)

	// Delete the application
	err = utils.AppTrustCli.Exec("app-delete", appKey)
	assert.NoError(t, err)

	// Verify the application no longer exists
	_, statusCode, err := utils.GetApplication(appKey)
	assert.NoError(t, err)
	assert.Equal(t, 404, statusCode)
}
