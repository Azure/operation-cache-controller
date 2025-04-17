package utils

import (
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/Azure/operation-cache-controller/api/v1alpha1"
)

const (
	TestNamespace = "operation-cache-controller-system"
)

func NewTestJobSpec(name string) batchv1.JobSpec {
	return batchv1.JobSpec{
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:    name,
						Image:   "mcr.microsoft.com/azurelinux/busybox:1.36",
						Command: []string{"echo", name + " job"},
					},
				},
			},
		},
	}
}

func NewTestApplicationSpec(name string) v1alpha1.ApplicationSpec {
	return v1alpha1.ApplicationSpec{
		Name:      name,
		Provision: NewTestJobSpec("provision"),
		Teardown:  NewTestJobSpec("teardown"),
	}
}

func NewSimpleOperationSpec(name string) *v1alpha1.OperationSpec {
	return &v1alpha1.OperationSpec{
		Applications: []v1alpha1.ApplicationSpec{NewTestApplicationSpec("app1")},
	}
}

func NewRequirement(name, namespace string) *v1alpha1.Requirement {
	return &v1alpha1.Requirement{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1alpha1.RequirementSpec{
			Template: v1alpha1.OperationSpec{
				Applications: []v1alpha1.ApplicationSpec{NewTestApplicationSpec("app1")},
			},
		},
	}
}
