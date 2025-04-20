package controller

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/Azure/operation-cache-controller/api/v1alpha1"
)

type RequirementHelper struct{}

func NewRequirementHelper() RequirementHelper { return RequirementHelper{} }
func (rh RequirementHelper) ClearConditions(r *v1alpha1.Requirement) {
	if r.Status.Conditions == nil {
		r.Status.Conditions = []metav1.Condition{}
	}
}

func (rh RequirementHelper) getCondition(r *v1alpha1.Requirement, conditionType string) (int, *metav1.Condition) {
	for i, condition := range r.Status.Conditions {
		if condition.Type == conditionType {
			return i, &condition
		}
	}
	return -1, nil
}

func (rh RequirementHelper) IsCacheMissed(r *v1alpha1.Requirement) bool {
	_, condition := rh.getCondition(r, v1alpha1.RequirementConditionCachedOperationAcquired)
	return condition == nil || condition.Status == metav1.ConditionFalse
}

func (rh RequirementHelper) UpdateCondition(r *v1alpha1.Requirement, conditionType string, conditionStatus metav1.ConditionStatus, reason, message string) bool {
	condition := &metav1.Condition{
		Type:               conditionType,
		Status:             conditionStatus,
		LastTransitionTime: metav1.Now(),
		Reason:             reason,
		Message:            message,
	}
	// Try to find this pod condition.
	idx, oldCondition := rh.getCondition(r, condition.Type)

	if oldCondition == nil {
		// We are adding new pod condition.
		r.Status.Conditions = append(r.Status.Conditions, *condition)
		return true
	}
	// We are updating an existing condition, so we need to check if it has changed.
	if condition.Status == oldCondition.Status {
		condition.LastTransitionTime = oldCondition.LastTransitionTime
	}

	isEqual := condition.Status == oldCondition.Status &&
		condition.Reason == oldCondition.Reason &&
		condition.Message == oldCondition.Message &&
		condition.LastTransitionTime.Equal(&oldCondition.LastTransitionTime)

	r.Status.Conditions[idx] = *condition
	// Return true if one of the fields have changed.
	return !isEqual
}
