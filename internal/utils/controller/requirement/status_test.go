package requirement

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1alpha1 "github.com/Azure/operation-cache-controller/api/v1alpha1"
)

// --- Test ClearConditions ---
func TestClearConditions(t *testing.T) {
	// Requirement with nil Conditions.
	req := &v1alpha1.Requirement{
		Status: v1alpha1.RequirementStatus{
			// Conditions nil for test.
		},
	}
	ClearConditions(req)
	require.NotNil(t, req.Status.Conditions)
	require.Equal(t, 0, len(req.Status.Conditions))
}

// --- Test IsCacheMissed ---
func TestIsCacheMissed(t *testing.T) {
	tests := []struct {
		name       string
		conditions []metav1.Condition
		wantMissed bool
	}{
		{
			name:       "no conditions",
			conditions: nil,
			wantMissed: true,
		},
		{
			name: "condition exists - false status",
			conditions: []metav1.Condition{
				{
					Type:   ConditionCachedOperationAcquired,
					Status: metav1.ConditionFalse,
				},
			},
			wantMissed: true,
		},
		{
			name: "condition exists - true status",
			conditions: []metav1.Condition{
				{
					Type:   ConditionCachedOperationAcquired,
					Status: metav1.ConditionTrue,
				},
			},
			wantMissed: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &v1alpha1.Requirement{
				Status: v1alpha1.RequirementStatus{
					Conditions: tt.conditions,
				},
			}
			got := IsCacheMissed(req)
			require.Equal(t, tt.wantMissed, got)
		})
	}
}

// --- Test UpdateCondition ---
func TestUpdateCondition(t *testing.T) {
	// Helper to create a Requirement with given conditions.
	newReq := func(conds []metav1.Condition) *v1alpha1.Requirement {
		return &v1alpha1.Requirement{
			Status: v1alpha1.RequirementStatus{
				Conditions: conds,
			},
		}
	}

	now := metav1.NewTime(time.Now())

	tests := []struct {
		name            string
		initialConds    []metav1.Condition
		conditionType   string
		conditionStatus metav1.ConditionStatus
		reason          string
		message         string
		wantChanged     bool
	}{
		{
			name:            "add new condition",
			initialConds:    []metav1.Condition{},
			conditionType:   "TestCondition",
			conditionStatus: metav1.ConditionTrue,
			reason:          "ReasonA",
			message:         "MessageA",
			wantChanged:     true,
		},
		{
			name: "no change when same condition exists",
			initialConds: []metav1.Condition{
				{
					Type:               "TestCondition",
					Status:             metav1.ConditionTrue,
					Reason:             "ReasonA",
					Message:            "MessageA",
					LastTransitionTime: now,
				},
			},
			conditionType:   "TestCondition",
			conditionStatus: metav1.ConditionTrue,
			reason:          "ReasonA",
			message:         "MessageA",
			wantChanged:     false,
		},
		{
			name: "update existing condition if field changes",
			initialConds: []metav1.Condition{
				{
					Type:               "TestCondition",
					Status:             metav1.ConditionFalse,
					Reason:             "OldReason",
					Message:            "OldMessage",
					LastTransitionTime: now,
				},
			},
			conditionType:   "TestCondition",
			conditionStatus: metav1.ConditionTrue,
			reason:          "NewReason",
			message:         "NewMessage",
			wantChanged:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := newReq(tt.initialConds)
			changed := UpdateCondition(req, tt.conditionType, tt.conditionStatus, tt.reason, tt.message)
			require.Equal(t, tt.wantChanged, changed)

			// Verify the condition is updated/added.
			var found *metav1.Condition
			for i := range req.Status.Conditions {
				if req.Status.Conditions[i].Type == tt.conditionType {
					found = &req.Status.Conditions[i]
					break
				}
			}
			require.NotNil(t, found)
			require.Equal(t, tt.conditionStatus, found.Status)
			require.Equal(t, tt.reason, found.Reason)
			require.Equal(t, tt.message, found.Message)
		})
	}
}
