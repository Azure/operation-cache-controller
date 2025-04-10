package appdeployment_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	appv1 "github.com/Azure/operation-cache-controller/api/v1"
	"github.com/Azure/operation-cache-controller/internal/utils/controller/appdeployment"
)

func TestClearConditions(t *testing.T) {
	tests := []struct {
		name               string
		existingConditions []metav1.Condition
	}{
		{
			name:               "No existing conditions",
			existingConditions: []metav1.Condition{},
		},
		{
			name: "Multiple existing conditions",
			existingConditions: []metav1.Condition{
				{
					Type:    "TestType",
					Status:  metav1.ConditionTrue,
					Reason:  "TestReason",
					Message: "TestMessage",
				},
				{
					Type:    "AnotherTestType",
					Status:  metav1.ConditionFalse,
					Reason:  "AnotherTestReason",
					Message: "AnotherTestMessage",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appDep := &appv1.AppDeployment{
				Status: appv1.AppDeploymentStatus{
					Conditions: tt.existingConditions,
				},
			}
			appdeployment.ClearConditions(context.Background(), appDep)
			assert.Empty(t, appDep.Status.Conditions, "conditions should be cleared")
		})
	}
}
