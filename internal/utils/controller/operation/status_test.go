package operation

import (
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1alpha1 "github.com/Azure/operation-cache-controller/api/v1alpha1"
)

func TestConditions(t *testing.T) {
	operation := &v1alpha1.Operation{}
	operation.Status.Conditions = []metav1.Condition{
		{
			Type:    "test",
			Status:  "test",
			Reason:  "test",
			Message: "test",
		},
	}
	ClearConditions(operation)
	assert.Equal(t, 0, len(operation.Status.Conditions))
}
