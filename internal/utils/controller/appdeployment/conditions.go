package appdeployment

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	appv1 "github.com/Azure/operation-cache-controller/api/v1"
)

func ClearConditions(ctx context.Context, appdeployment *appv1.AppDeployment) {
	// Clear all conditions
	appdeployment.Status.Conditions = []metav1.Condition{}
}
