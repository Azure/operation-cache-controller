package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/Azure/operation-cache-controller/api/v1alpha1"
	"github.com/Azure/operation-cache-controller/internal/utils/ptr"
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		app     v1alpha1.AppDeployment
		wantErr bool
	}{
		{
			name: "Valid AppDeployment",
			app: v1alpha1.AppDeployment{
				Spec: v1alpha1.AppDeploymentSpec{
					Provision: batchv1.JobSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								RestartPolicy: corev1.RestartPolicyNever,
								Containers: []corev1.Container{
									{
										Image:   "nginx:latest",
										Command: []string{"nginx"},
									},
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name:    "Empty Spec",
			app:     v1alpha1.AppDeployment{},
			wantErr: true,
		},
		{
			name: "Invalid Provision Job",
			app: v1alpha1.AppDeployment{
				Spec: v1alpha1.AppDeploymentSpec{
					OpId: "test",
					Provision: batchv1.JobSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Image: "nginx:latest",
										Command: []string{
											"nginx",
										},
									},
								},
							},
						},
						SuccessPolicy: &batchv1.SuccessPolicy{
							Rules: []batchv1.SuccessPolicyRule{
								{
									SucceededIndexes: ptr.Of(string("test")),
									SucceededCount:   ptr.Of(int32(1)),
								},
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Invalid TearDown Job",
			app: v1alpha1.AppDeployment{
				Spec: v1alpha1.AppDeploymentSpec{
					OpId: "test",
					Provision: batchv1.JobSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Image: "nginx:latest",
										Command: []string{
											"nginx",
										},
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
										Command: []string{
											"nginx",
										},
									},
								},
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "jobConstraint activeDeadlineSeconds",
			app: v1alpha1.AppDeployment{
				Spec: v1alpha1.AppDeploymentSpec{
					Provision: batchv1.JobSpec{
						ActiveDeadlineSeconds: ptr.Of(int64(1)),
					},
				},
			},
			wantErr: true,
		},
		{
			name: "jobConstraint backoffLimit",
			app: v1alpha1.AppDeployment{
				Spec: v1alpha1.AppDeploymentSpec{
					Provision: batchv1.JobSpec{
						BackoffLimit: ptr.Of(int32(1)),
					},
				},
			},
			wantErr: true,
		},
		{
			name: "jobConstraint backoffLimitPerIndex",
			app: v1alpha1.AppDeployment{
				Spec: v1alpha1.AppDeploymentSpec{
					Provision: batchv1.JobSpec{
						BackoffLimitPerIndex: ptr.Of(int32(1)),
					},
				},
			},
			wantErr: true,
		},
		{
			name: "jobConstraint completions",
			app: v1alpha1.AppDeployment{
				Spec: v1alpha1.AppDeploymentSpec{
					Provision: batchv1.JobSpec{
						Completions: ptr.Of(int32(1)),
					},
				},
			},
			wantErr: true,
		},
		{
			name: "jobConstraint completionMode",
			app: v1alpha1.AppDeployment{
				Spec: v1alpha1.AppDeploymentSpec{
					Provision: batchv1.JobSpec{
						CompletionMode: ptr.Of(batchv1.NonIndexedCompletion),
					},
				},
			},
			wantErr: true,
		},
		{
			name: "jobConstraint managedBy",
			app: v1alpha1.AppDeployment{
				Spec: v1alpha1.AppDeploymentSpec{
					Provision: batchv1.JobSpec{
						ManagedBy: ptr.Of("test"),
					},
				},
			},
			wantErr: true,
		},
		{
			name: "jobConstraint manualSelector",
			app: v1alpha1.AppDeployment{
				Spec: v1alpha1.AppDeploymentSpec{
					Provision: batchv1.JobSpec{
						ManualSelector: ptr.Of(true),
					},
				},
			},
			wantErr: true,
		},
		{
			name: "jobConstraint maxFailedIndexes",
			app: v1alpha1.AppDeployment{
				Spec: v1alpha1.AppDeploymentSpec{
					Provision: batchv1.JobSpec{
						MaxFailedIndexes: ptr.Of(int32(1)),
					},
				},
			},
			wantErr: true,
		},
		{
			name: "jobConstraint parallelism",
			app: v1alpha1.AppDeployment{
				Spec: v1alpha1.AppDeploymentSpec{
					Provision: batchv1.JobSpec{
						Parallelism: ptr.Of(int32(1)),
					},
				},
			},
			wantErr: true,
		},
		{
			name: "jobConstraint podFailurePolicy",
			app: v1alpha1.AppDeployment{
				Spec: v1alpha1.AppDeploymentSpec{
					Provision: batchv1.JobSpec{
						PodFailurePolicy: ptr.Of(batchv1.PodFailurePolicy{
							Rules: []batchv1.PodFailurePolicyRule{
								{
									Action: batchv1.PodFailurePolicyActionFailJob,
								},
							},
						}),
					},
				},
			},
			wantErr: true,
		},
		{
			name: "jobConstraint podReplacementPolicy",
			app: v1alpha1.AppDeployment{
				Spec: v1alpha1.AppDeploymentSpec{
					Provision: batchv1.JobSpec{
						PodReplacementPolicy: ptr.Of(batchv1.Failed),
					},
				},
			},
			wantErr: true,
		},
		{
			name: "jobConstraint Selectors",
			app: v1alpha1.AppDeployment{
				Spec: v1alpha1.AppDeploymentSpec{
					Provision: batchv1.JobSpec{
						Selector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"app": "test",
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "jobConstraint TTLSecondsAfterFinished",
			app: v1alpha1.AppDeployment{
				Spec: v1alpha1.AppDeploymentSpec{
					Provision: batchv1.JobSpec{
						TTLSecondsAfterFinished: ptr.Of(int32(1)),
					},
				},
			},
			wantErr: true,
		},
		{
			name: "jobConstraint SuccessPolicy",
			app: v1alpha1.AppDeployment{
				Spec: v1alpha1.AppDeploymentSpec{
					Provision: batchv1.JobSpec{
						CompletionMode: new(batchv1.CompletionMode),
					},
				},
			},
			wantErr: true,
		},
		{
			name: "jobConstraint Suspend",
			app: v1alpha1.AppDeployment{
				Spec: v1alpha1.AppDeploymentSpec{
					Provision: batchv1.JobSpec{
						Suspend: ptr.Of(true),
					},
				},
			},
			wantErr: true,
		},
		{
			name: "jobConstraint PodTemplate",
			app: v1alpha1.AppDeployment{
				Spec: v1alpha1.AppDeploymentSpec{
					OpId: "test",
					Provision: batchv1.JobSpec{
						Template: corev1.PodTemplateSpec{},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "podConstraint name",
			app: v1alpha1.AppDeployment{
				Spec: v1alpha1.AppDeploymentSpec{
					Provision: batchv1.JobSpec{
						Template: corev1.PodTemplateSpec{
							ObjectMeta: metav1.ObjectMeta{
								Name: "app",
							},
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Image: "nginx:latest",
										Command: []string{
											"nginx",
										},
									},
								},
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "podConstraint namespace",
			app: v1alpha1.AppDeployment{
				Spec: v1alpha1.AppDeploymentSpec{
					Provision: batchv1.JobSpec{
						Template: corev1.PodTemplateSpec{
							ObjectMeta: metav1.ObjectMeta{
								Namespace: "testns",
							},
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Image: "nginx:latest",
										Command: []string{
											"nginx",
										},
									},
								},
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "podConstraint volumes",
			app: v1alpha1.AppDeployment{
				Spec: v1alpha1.AppDeploymentSpec{
					Provision: batchv1.JobSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Volumes: []corev1.Volume{
									{
										Name: "test",
									},
								},
								Containers: []corev1.Container{
									{
										Image: "nginx:latest",
										Command: []string{
											"nginx",
										},
									},
								},
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "podConstraint initContainers",
			app: v1alpha1.AppDeployment{
				Spec: v1alpha1.AppDeploymentSpec{
					Provision: batchv1.JobSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								InitContainers: []corev1.Container{
									{
										Image: "nginx:latest",
										Command: []string{
											"nginx",
										},
									},
								},
								Containers: []corev1.Container{
									{
										Image: "nginx:latest",
										Command: []string{
											"nginx",
										},
									},
								},
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "podConstraint container count should be 1",
			app: v1alpha1.AppDeployment{
				Spec: v1alpha1.AppDeploymentSpec{
					Provision: batchv1.JobSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Image: "nginx:latest",
										Command: []string{
											"nginx",
										},
									},
									{
										Image: "nginx:latest",
										Command: []string{
											"nginx",
										},
									},
								},
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "containerConstraint image is empty",
			app: v1alpha1.AppDeployment{
				Spec: v1alpha1.AppDeploymentSpec{
					Provision: batchv1.JobSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Command: []string{
											"nginx",
										},
									},
								},
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "containerConstraint volumeMounts are not allowed",
			app: v1alpha1.AppDeployment{
				Spec: v1alpha1.AppDeploymentSpec{
					Provision: batchv1.JobSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Image: "nginx:latest",
										Command: []string{
											"nginx",
										},
										VolumeMounts: []corev1.VolumeMount{
											{
												Name: "test",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(&tt.app)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
