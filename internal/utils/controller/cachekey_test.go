package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"

	v1alpha1 "github.com/Azure/operation-cache-controller/api/v1alpha1"
)

func TestNewCacheKey(t *testing.T) {
	tests := []struct {
		name     string
		source   AppCacheField
		expected string
	}{
		{
			name: "basic",
			source: AppCacheField{
				Name:       "test-app-1",
				Image:      "nginx:latest",
				Command:    []string{"nginx"},
				Args:       []string{"-g", "daemon off;"},
				WorkingDir: "/",
				Env: []corev1.EnvVar{
					{Name: "ENV_VAR1", Value: "value1"},
					{Name: "ENV_VAR2", Value: "value2"},
				},
			},
			expected: "126075cdb21390e83c4703ace30f4e028568f38c193c9075171fea907570d23e", // Set the expected value after running the test once
		},
		{
			name: "basic",
			source: AppCacheField{
				Name:       "test-app-1",
				Image:      "nginx:latest",
				Command:    []string{"nginx"},
				Args:       []string{"-g", "daemon off;"},
				WorkingDir: "/",
				Env: []corev1.EnvVar{
					{Name: "ENV_VAR2", Value: "value2"},
					{Name: "ENV_VAR1", Value: "value1"},
				},
			},
			expected: "126075cdb21390e83c4703ace30f4e028568f38c193c9075171fea907570d23e", // Set the expected value after running the test once
		},
		{
			name: "basic with dependencies",
			source: AppCacheField{
				Name:       "test-app-2",
				Image:      "nginx:latest",
				Command:    []string{"nginx"},
				Args:       []string{"-g", "daemon off;"},
				WorkingDir: "/",
				Env: []corev1.EnvVar{
					{Name: "ENV_VAR1", Value: "value1"},
					{Name: "ENV_VAR2", Value: "value2"},
				},
				Dependencies: []string{"test-app-1"},
			},
			expected: "b5eab152dfbbb5d9d0f0b6c2920204b72991a86af0d6ebd673aaf4b80761b207",
		},
		{
			name: "basic with dependencies",
			source: AppCacheField{
				Name:       "test-app-2",
				Image:      "nginx:latest",
				Command:    []string{"nginx"},
				Args:       []string{"-g", "daemon off;"},
				WorkingDir: "/",
				Env: []corev1.EnvVar{
					{Name: "ENV_VAR2", Value: "value2"},
					{Name: "ENV_VAR1", Value: "value1"},
				},
				Dependencies: []string{"test-app-1"},
			},
			expected: "b5eab152dfbbb5d9d0f0b6c2920204b72991a86af0d6ebd673aaf4b80761b207",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.source.NewCacheKey()

			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestNewCacheKeyFromApplications(t *testing.T) {
	tests := []struct {
		name     string
		source   []v1alpha1.ApplicationSpec
		expected string
	}{
		{
			name: "basic",
			source: []v1alpha1.ApplicationSpec{
				{
					Name: "test-app-1",
					Provision: batchv1.JobSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
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
			expected: "fac2c1ee35f29e0cf01df3d2d087aa63f816cddcf37519b86f9788a8fe437db0",
		},
		{
			name: "basic with dependencies",
			source: []v1alpha1.ApplicationSpec{
				{
					Name: "test-app-1",
					Provision: batchv1.JobSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
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
				{
					Name: "test-app-2",
					Provision: batchv1.JobSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Image:   "nginx:latest",
										Command: []string{"nginx"},
									},
								},
							},
						},
					},
					Dependencies: []string{"test-app-1"},
				},
			},
			expected: "4ad519825833e792749137e9ff7ad3efcab0907f9d3adc8751eef2cf871648a4",
		},

		{
			name: "basic with multiple dependencies",
			source: []v1alpha1.ApplicationSpec{
				{
					Name: "test-app-1",
					Provision: batchv1.JobSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
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
				{
					Name: "test-app-2",
					Provision: batchv1.JobSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Image:   "nginx:latest",
										Command: []string{"nginx"},
									},
								},
							},
						},
					},
					Dependencies: []string{"test-app-1"},
				},
				{
					Name: "test-app-3",
					Provision: batchv1.JobSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Image:   "nginx:latest",
										Command: []string{"nginx"},
									},
								},
							},
						},
					},
					Dependencies: []string{"test-app-1", "test-app-2"},
				},
			},
			expected: "d9dc8143d50f0758ab226d2a199d3f884569750fa7c0a1928b2597af3eb3f7f0",
		},
		{
			name: "basic with multiple dependencies and different order",
			source: []v1alpha1.ApplicationSpec{
				{
					Name: "test-app-1",
					Provision: batchv1.JobSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
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
				{
					Name: "test-app-3",
					Provision: batchv1.JobSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Image:   "nginx:latest",
										Command: []string{"nginx"},
									},
								},
							},
						},
					},
					Dependencies: []string{"test-app-2", "test-app-1"},
				},
				{
					Name: "test-app-2",
					Provision: batchv1.JobSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Image:   "nginx:latest",
										Command: []string{"nginx"},
									},
								},
							},
						},
					},
					Dependencies: []string{"test-app-1"},
				},
			},

			expected: "d9dc8143d50f0758ab226d2a199d3f884569750fa7c0a1928b2597af3eb3f7f0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := NewCacheKeyFromApplications(tt.source)

			assert.Equal(t, tt.expected, actual)
		})
	}
}
