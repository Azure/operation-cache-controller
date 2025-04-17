package appdeployment

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/Azure/operation-cache-controller/api/v1alpha1"
)

func ClearConditions(ctx context.Context, appdeployment *v1alpha1.AppDeployment) {
	// Clear all conditions
	appdeployment.Status.Conditions = []metav1.Condition{}
}
