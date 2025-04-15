package resources

import (
	appsv1 "github.com/Azure/operation-cache-controller/api/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var SampleCache appsv1.Cache = appsv1.Cache{
	ObjectMeta: metav1.ObjectMeta{Name: "cache-sample"},
	Spec: appsv1.CacheSpec{
		Strategy: "",
		OperationTemplate: appsv1.OperationSpec{
			Applications: []appsv1.ApplicationSpec{
				{
					Name: "test-app",
					Provision: batchv1.JobSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name:  "test-app-provision",
										Image: "aksartifactsmsftprodeastus.azurecr.io/devinfra/underlay:sha256:05c616236b999cb9e610cd7afede57adc90516f3db6f30408c72a907889ad8dc",
									},
								},
							},
						},
					},
					Teardown: batchv1.JobSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name:  "test-app-provision",
										Image: "aksartifactsmsftprodeastus.azurecr.io/devinfra/underlay:sha256:05c616236b999cb9e610cd7afede57adc90516f3db6f30408c72a907889ad8dc",
									},
								},
							},
						},
					},
				},
			},
		},
	},
}
