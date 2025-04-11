package operation

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	appsv1 "github.com/Azure/operation-cache-controller/api/v1"
)

func ClearConditions(operation *appsv1.Operation) {
	operation.Status.Conditions = []metav1.Condition{}
}
