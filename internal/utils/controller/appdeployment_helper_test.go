package controller_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/Azure/operation-cache-controller/api/v1alpha1"
	ctrlutils "github.com/Azure/operation-cache-controller/internal/utils/controller"
)

var helper = ctrlutils.NewAppDeploymentHelper()

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
			appDep := &v1alpha1.AppDeployment{
				Status: v1alpha1.AppDeploymentStatus{
					Conditions: tt.existingConditions,
				},
			}
			helper.ClearConditions(context.Background(), appDep)
			assert.Empty(t, appDep.Status.Conditions, "conditions should be cleared")
		})
	}
}
