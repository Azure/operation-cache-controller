package utils

import (
	appsv1 "github.com/Azure/operation-cache-controller/api/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func NewTestApplicationSpec(name string) appsv1.ApplicationSpec {
	return appsv1.ApplicationSpec{
		Name:      name,
		Provision: NewTestJobSpec("provision"),
		Teardown:  NewTestJobSpec("teardown"),
	}
}

func NewSimpleOperationSpec(name string) *appsv1.OperationSpec {
	return &appsv1.OperationSpec{
		Applications: []appsv1.ApplicationSpec{NewTestApplicationSpec("app1")},
	}
}

func NewRequirement(name, namespace string) *appsv1.Requirement {
	return &appsv1.Requirement{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: appsv1.RequirementSpec{
			Template: appsv1.OperationSpec{
				Applications: []appsv1.ApplicationSpec{NewTestApplicationSpec("app1")},
			},
		},
	}
}
