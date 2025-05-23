package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"

	"github.com/Azure/operation-cache-controller/api/v1alpha1"
)

var cacheHelper = NewCacheHelper()

func TestDefaultCacheExpireTime(t *testing.T) {
	result := cacheHelper.DefaultCacheExpireTime()
	require.NotEmpty(t, result)
}

func TestRandomSelectCachedOperation(t *testing.T) {
	tests := []struct {
		name        string
		caches      []string
		expectEmpty bool
	}{
		{"empty caches", nil, true},
		{"non-empty caches", []string{"cache1", "cache2", "cache3"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// ...existing setup code if any...
			cacheInstance := &v1alpha1.Cache{
				Status: v1alpha1.CacheStatus{
					AvailableCaches: tt.caches,
				},
			}

			result := cacheHelper.RandomSelectCachedOperation(cacheInstance)

			if tt.expectEmpty {
				require.Equal(t, "", result)
			} else {
				require.Contains(t, tt.caches, result)
			}
			// ...existing teardown code if any...
		})
	}
}
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
			actual := cacheHelper.NewCacheKeyFromApplications(tt.source)

			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestAppCacheFieldFromApplicationTeardown(t *testing.T) {
	tests := []struct {
		name     string
		app      v1alpha1.ApplicationSpec
		expected AppCacheField
	}{
		{
			name: "basic teardown",
			app: v1alpha1.ApplicationSpec{
				Name: "test-app-1",
				Teardown: batchv1.JobSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Image:      "nginx:cleanup",
									Command:    []string{"cleanup"},
									Args:       []string{"--all"},
									WorkingDir: "/cleanup",
									Env: []corev1.EnvVar{
										{Name: "CLEANUP_MODE", Value: "full"},
										{Name: "DEBUG", Value: "true"},
									},
								},
							},
						},
					},
				},
				Dependencies: []string{"dep-1", "dep-2"},
			},
			expected: AppCacheField{
				Name:         "test-app-1",
				Image:        "nginx:cleanup",
				Command:      []string{"cleanup"},
				Args:         []string{"--all"},
				WorkingDir:   "/cleanup",
				Env:          []corev1.EnvVar{{Name: "CLEANUP_MODE", Value: "full"}, {Name: "DEBUG", Value: "true"}},
				Dependencies: []string{"dep-1", "dep-2"},
			},
		},
		{
			name: "teardown with empty fields",
			app: v1alpha1.ApplicationSpec{
				Name: "test-app-minimal",
				Teardown: batchv1.JobSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Image: "minimal:latest",
								},
							},
						},
					},
				},
			},
			expected: AppCacheField{
				Name:         "test-app-minimal",
				Image:        "minimal:latest",
				Command:      nil,
				Args:         nil,
				WorkingDir:   "",
				Env:          nil,
				Dependencies: nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := cacheHelper.AppCacheFieldFromApplicationTeardown(tt.app)

			assert.Equal(t, tt.expected.Name, actual.Name)
			assert.Equal(t, tt.expected.Image, actual.Image)
			assert.Equal(t, tt.expected.Command, actual.Command)
			assert.Equal(t, tt.expected.Args, actual.Args)
			assert.Equal(t, tt.expected.WorkingDir, actual.WorkingDir)
			assert.Equal(t, tt.expected.Env, actual.Env)
			assert.Equal(t, tt.expected.Dependencies, actual.Dependencies)
		})
	}
}

func TestAppCacheFieldFromApplicationProvision(t *testing.T) {
	tests := []struct {
		name     string
		app      v1alpha1.ApplicationSpec
		expected AppCacheField
	}{
		{
			name: "basic provision",
			app: v1alpha1.ApplicationSpec{
				Name: "test-app-1",
				Provision: batchv1.JobSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Image:      "nginx:latest",
									Command:    []string{"nginx"},
									Args:       []string{"-g", "daemon off;"},
									WorkingDir: "/app",
									Env: []corev1.EnvVar{
										{Name: "ENV_VAR1", Value: "value1"},
										{Name: "ENV_VAR2", Value: "value2"},
									},
								},
							},
						},
					},
				},
				Dependencies: []string{"dep-1", "dep-2"},
			},
			expected: AppCacheField{
				Name:         "test-app-1",
				Image:        "nginx:latest",
				Command:      []string{"nginx"},
				Args:         []string{"-g", "daemon off;"},
				WorkingDir:   "/app",
				Env:          []corev1.EnvVar{{Name: "ENV_VAR1", Value: "value1"}, {Name: "ENV_VAR2", Value: "value2"}},
				Dependencies: []string{"dep-1", "dep-2"},
			},
		},
		{
			name: "provision with empty fields",
			app: v1alpha1.ApplicationSpec{
				Name: "test-app-minimal",
				Provision: batchv1.JobSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Image: "minimal:latest",
								},
							},
						},
					},
				},
			},
			expected: AppCacheField{
				Name:         "test-app-minimal",
				Image:        "minimal:latest",
				Command:      nil,
				Args:         nil,
				WorkingDir:   "",
				Env:          nil,
				Dependencies: nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := cacheHelper.AppCacheFieldFromApplicationProvision(tt.app)

			assert.Equal(t, tt.expected.Name, actual.Name)
			assert.Equal(t, tt.expected.Image, actual.Image)
			assert.Equal(t, tt.expected.Command, actual.Command)
			assert.Equal(t, tt.expected.Args, actual.Args)
			assert.Equal(t, tt.expected.WorkingDir, actual.WorkingDir)
			assert.Equal(t, tt.expected.Env, actual.Env)
			assert.Equal(t, tt.expected.Dependencies, actual.Dependencies)
		})
	}
}
