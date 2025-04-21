package controller

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/Azure/operation-cache-controller/api/v1alpha1"
)

const (
	// env keys
	OperationIDEnvKey = "OPERATION_ID"
)

type AppDeploymentHelper struct{}

func NewAppDeploymentHelper() AppDeploymentHelper { return AppDeploymentHelper{} }

func (ad AppDeploymentHelper) ClearConditions(ctx context.Context, appdeployment *v1alpha1.AppDeployment) {
	appdeployment.Status.Conditions = []metav1.Condition{}
}
