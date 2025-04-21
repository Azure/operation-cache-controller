package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/Azure/operation-cache-controller/api/v1alpha1"
)

var helper = NewOperationHelper()

func TestNewOperationId(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "basic"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id := helper.NewOperationId()
			assert.NotEmpty(t, id)
		})
	}
}

func TestDiffAppDeployments(t *testing.T) {
	tests := []struct {
		name        string
		expected    []v1alpha1.AppDeployment
		actual      []v1alpha1.AppDeployment
		equals      func(a, b v1alpha1.AppDeployment) bool
		wantAdded   int
		wantRemoved int
		wantUpdated int
	}{
		{
			name:      "basic example",
			expected:  []v1alpha1.AppDeployment{{ObjectMeta: metav1.ObjectMeta{Name: "app1"}}, {ObjectMeta: metav1.ObjectMeta{Name: "app2"}}},
			actual:    []v1alpha1.AppDeployment{{ObjectMeta: metav1.ObjectMeta{Name: "app2"}}, {ObjectMeta: metav1.ObjectMeta{Name: "app3"}}},
			equals:    func(a, b v1alpha1.AppDeployment) bool { return a.Name == b.Name },
			wantAdded: 1, wantRemoved: 1, wantUpdated: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			added, removed, updated := helper.DiffAppDeployments(tt.expected, tt.actual, tt.equals)
			assert.Equal(t, tt.wantAdded, len(added))
			assert.Equal(t, tt.wantRemoved, len(removed))
			assert.Equal(t, tt.wantUpdated, len(updated))
		})
	}
}

func TestCompareProvisionJobs(t *testing.T) {
	tests := []struct {
		name string
		a    v1alpha1.AppDeployment
		b    v1alpha1.AppDeployment
		want bool
	}{
		{
			name: "different images",
			a: v1alpha1.AppDeployment{
				Spec: v1alpha1.AppDeploymentSpec{
					Provision: batchv1.JobSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{{Image: "image1"}},
							},
						},
					},
					Dependencies: []string{"dep1", "dep2"},
				},
			},
			b: v1alpha1.AppDeployment{
				Spec: v1alpha1.AppDeploymentSpec{
					Provision: batchv1.JobSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{{Image: "image2"}},
							},
						},
					},
					Dependencies: []string{"dep2", "dep1"},
				},
			},
			want: false,
		},
		{
			name: "identical provision jobs with same dependencies",
			a: v1alpha1.AppDeployment{
				Spec: v1alpha1.AppDeploymentSpec{
					Provision: batchv1.JobSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{{Image: "image1"}},
							},
						},
					},
					Dependencies: []string{"dep1", "dep2"},
				},
			},
			b: v1alpha1.AppDeployment{
				Spec: v1alpha1.AppDeploymentSpec{
					Provision: batchv1.JobSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{{Image: "image1"}},
							},
						},
					},
					Dependencies: []string{"dep2", "dep1"},
				},
			},
			want: true,
		},
		{
			name: "identical provision jobs with different dependencies",
			a: v1alpha1.AppDeployment{
				Spec: v1alpha1.AppDeploymentSpec{
					Provision: batchv1.JobSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{{Image: "image1"}},
							},
						},
					},
					Dependencies: []string{"dep1", "dep2"},
				},
			},
			b: v1alpha1.AppDeployment{
				Spec: v1alpha1.AppDeploymentSpec{
					Provision: batchv1.JobSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{{Image: "image1"}},
							},
						},
					},
					Dependencies: []string{"dep1", "dep3"},
				},
			},
			want: false,
		},
		{
			name: "identical provision jobs with different number of dependencies",
			a: v1alpha1.AppDeployment{
				Spec: v1alpha1.AppDeploymentSpec{
					Provision: batchv1.JobSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{{Image: "image1"}},
							},
						},
					},
					Dependencies: []string{"dep1", "dep2", "dep3"},
				},
			},
			b: v1alpha1.AppDeployment{
				Spec: v1alpha1.AppDeploymentSpec{
					Provision: batchv1.JobSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{{Image: "image1"}},
							},
						},
					},
					Dependencies: []string{"dep1", "dep2"},
				},
			},
			want: false,
		},
		{
			name: "identical provision jobs with empty dependencies",
			a: v1alpha1.AppDeployment{
				Spec: v1alpha1.AppDeploymentSpec{
					Provision: batchv1.JobSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{{Image: "image1"}},
							},
						},
					},
					Dependencies: []string{},
				},
			},
			b: v1alpha1.AppDeployment{
				Spec: v1alpha1.AppDeploymentSpec{
					Provision: batchv1.JobSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{{Image: "image1"}},
							},
						},
					},
					Dependencies: []string{},
				},
			},
			want: true,
		},
		{
			name: "identical provision jobs with nil and empty dependencies",
			a: v1alpha1.AppDeployment{
				Spec: v1alpha1.AppDeploymentSpec{
					Provision: batchv1.JobSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{{Image: "image1"}},
							},
						},
					},
					Dependencies: nil,
				},
			},
			b: v1alpha1.AppDeployment{
				Spec: v1alpha1.AppDeploymentSpec{
					Provision: batchv1.JobSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{{Image: "image1"}},
							},
						},
					},
					Dependencies: []string{},
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := helper.CompareProvisionJobs(tt.a, tt.b)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCompareProvisionJobsWithDependencies(t *testing.T) {
	tests := []struct {
		name string
		a    v1alpha1.AppDeployment
		b    v1alpha1.AppDeployment
		want bool
	}{
		{
			name: "different images",
			a: v1alpha1.AppDeployment{
				Spec: v1alpha1.AppDeploymentSpec{
					Provision: batchv1.JobSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{{Image: "image1"}},
							},
						},
					},
					Dependencies: []string{"dep1", "dep2"},
				},
			},
			b: v1alpha1.AppDeployment{
				Spec: v1alpha1.AppDeploymentSpec{
					Provision: batchv1.JobSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{{Image: "image2"}},
							},
						},
					},
					Dependencies: []string{"dep2", "dep1"},
				},
			},
			want: false,
		},
		{
			name: "identical provision jobs with same dependencies in different order",
			a: v1alpha1.AppDeployment{
				Spec: v1alpha1.AppDeploymentSpec{
					Provision: batchv1.JobSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{{Image: "image1"}},
							},
						},
					},
					Dependencies: []string{"dep1", "dep2"},
				},
			},
			b: v1alpha1.AppDeployment{
				Spec: v1alpha1.AppDeploymentSpec{
					Provision: batchv1.JobSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{{Image: "image1"}},
							},
						},
					},
					Dependencies: []string{"dep2", "dep1"},
				},
			},
			want: true,
		},
		{
			name: "identical provision jobs with different dependencies",
			a: v1alpha1.AppDeployment{
				Spec: v1alpha1.AppDeploymentSpec{
					Provision: batchv1.JobSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{{Image: "image1"}},
							},
						},
					},
					Dependencies: []string{"dep1", "dep2"},
				},
			},
			b: v1alpha1.AppDeployment{
				Spec: v1alpha1.AppDeploymentSpec{
					Provision: batchv1.JobSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{{Image: "image1"}},
							},
						},
					},
					Dependencies: []string{"dep1", "dep3"},
				},
			},
			want: false,
		},
		{
			name: "identical provision jobs with different number of dependencies",
			a: v1alpha1.AppDeployment{
				Spec: v1alpha1.AppDeploymentSpec{
					Provision: batchv1.JobSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{{Image: "image1"}},
							},
						},
					},
					Dependencies: []string{"dep1", "dep2", "dep3"},
				},
			},
			b: v1alpha1.AppDeployment{
				Spec: v1alpha1.AppDeploymentSpec{
					Provision: batchv1.JobSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{{Image: "image1"}},
							},
						},
					},
					Dependencies: []string{"dep1", "dep2"},
				},
			},
			want: false,
		},
		{
			name: "identical provision jobs with empty dependencies",
			a: v1alpha1.AppDeployment{
				Spec: v1alpha1.AppDeploymentSpec{
					Provision: batchv1.JobSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{{Image: "image1"}},
							},
						},
					},
					Dependencies: []string{},
				},
			},
			b: v1alpha1.AppDeployment{
				Spec: v1alpha1.AppDeploymentSpec{
					Provision: batchv1.JobSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{{Image: "image1"}},
							},
						},
					},
					Dependencies: []string{},
				},
			},
			want: true,
		},
		{
			name: "identical provision jobs with nil and empty dependencies",
			a: v1alpha1.AppDeployment{
				Spec: v1alpha1.AppDeploymentSpec{
					Provision: batchv1.JobSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{{Image: "image1"}},
							},
						},
					},
					Dependencies: nil,
				},
			},
			b: v1alpha1.AppDeployment{
				Spec: v1alpha1.AppDeploymentSpec{
					Provision: batchv1.JobSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{{Image: "image1"}},
							},
						},
					},
					Dependencies: []string{},
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := helper.CompareProvisionJobs(tt.a, tt.b)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCompareTeardownJobs(t *testing.T) {
	tests := []struct {
		name string
		a    v1alpha1.AppDeployment
		b    v1alpha1.AppDeployment
		want bool
	}{
		{
			name: "different teardown images",
			a: v1alpha1.AppDeployment{
				Spec: v1alpha1.AppDeploymentSpec{
					Teardown: batchv1.JobSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{{Image: "image1"}},
							},
						},
					},
				},
			},
			b: v1alpha1.AppDeployment{
				Spec: v1alpha1.AppDeploymentSpec{
					Teardown: batchv1.JobSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{{Image: "image2"}},
							},
						},
					},
				},
			},
			want: false,
		},
		{
			name: "identical teardown jobs",
			a: v1alpha1.AppDeployment{
				Spec: v1alpha1.AppDeploymentSpec{
					Teardown: batchv1.JobSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{{Image: "same-image"}},
							},
						},
					},
				},
			},
			b: v1alpha1.AppDeployment{
				Spec: v1alpha1.AppDeploymentSpec{
					Teardown: batchv1.JobSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{{Image: "same-image"}},
							},
						},
					},
				},
			},
			want: true,
		},
		{
			name: "both empty teardown containers",
			a: v1alpha1.AppDeployment{
				Spec: v1alpha1.AppDeploymentSpec{
					Teardown: batchv1.JobSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{},
							},
						},
					},
				},
			},
			b: v1alpha1.AppDeployment{
				Spec: v1alpha1.AppDeploymentSpec{
					Teardown: batchv1.JobSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{},
							},
						},
					},
				},
			},
			want: true,
		},
		{
			name: "one empty teardown container",
			a: v1alpha1.AppDeployment{
				Spec: v1alpha1.AppDeploymentSpec{
					Teardown: batchv1.JobSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{},
							},
						},
					},
				},
			},
			b: v1alpha1.AppDeployment{
				Spec: v1alpha1.AppDeploymentSpec{
					Teardown: batchv1.JobSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{{Image: "image"}},
							},
						},
					},
				},
			},
			want: false,
		},
		{
			name: "different commands in teardown",
			a: v1alpha1.AppDeployment{
				Spec: v1alpha1.AppDeploymentSpec{
					Teardown: batchv1.JobSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{{
									Image:   "same-image",
									Command: []string{"cmd1", "cmd2"},
								}},
							},
						},
					},
				},
			},
			b: v1alpha1.AppDeployment{
				Spec: v1alpha1.AppDeploymentSpec{
					Teardown: batchv1.JobSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{{
									Image:   "same-image",
									Command: []string{"cmd3", "cmd4"},
								}},
							},
						},
					},
				},
			},
			want: false,
		},
		{
			name: "different args in teardown",
			a: v1alpha1.AppDeployment{
				Spec: v1alpha1.AppDeploymentSpec{
					Teardown: batchv1.JobSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{{
									Image: "same-image",
									Args:  []string{"arg1", "arg2"},
								}},
							},
						},
					},
				},
			},
			b: v1alpha1.AppDeployment{
				Spec: v1alpha1.AppDeploymentSpec{
					Teardown: batchv1.JobSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{{
									Image: "same-image",
									Args:  []string{"arg3", "arg4"},
								}},
							},
						},
					},
				},
			},
			want: false,
		},
		{
			name: "different working dir in teardown",
			a: v1alpha1.AppDeployment{
				Spec: v1alpha1.AppDeploymentSpec{
					Teardown: batchv1.JobSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{{
									Image:      "same-image",
									WorkingDir: "/dir1",
								}},
							},
						},
					},
				},
			},
			b: v1alpha1.AppDeployment{
				Spec: v1alpha1.AppDeploymentSpec{
					Teardown: batchv1.JobSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{{
									Image:      "same-image",
									WorkingDir: "/dir2",
								}},
							},
						},
					},
				},
			},
			want: false,
		},
		{
			name: "different env vars in teardown",
			a: v1alpha1.AppDeployment{
				Spec: v1alpha1.AppDeploymentSpec{
					Teardown: batchv1.JobSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{{
									Image: "same-image",
									Env: []corev1.EnvVar{
										{Name: "VAR1", Value: "value1"},
									},
								}},
							},
						},
					},
				},
			},
			b: v1alpha1.AppDeployment{
				Spec: v1alpha1.AppDeploymentSpec{
					Teardown: batchv1.JobSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{{
									Image: "same-image",
									Env: []corev1.EnvVar{
										{Name: "VAR1", Value: "different"},
									},
								}},
							},
						},
					},
				},
			},
			want: false,
		},
		{
			name: "complex identical teardown jobs",
			a: v1alpha1.AppDeployment{
				Spec: v1alpha1.AppDeploymentSpec{
					Teardown: batchv1.JobSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{{
									Image:      "same-image",
									Command:    []string{"cmd1", "cmd2"},
									Args:       []string{"arg1", "arg2"},
									WorkingDir: "/workdir",
									Env: []corev1.EnvVar{
										{Name: "VAR1", Value: "value1"},
										{Name: "VAR2", Value: "value2"},
									},
								}},
							},
						},
					},
				},
			},
			b: v1alpha1.AppDeployment{
				Spec: v1alpha1.AppDeploymentSpec{
					Teardown: batchv1.JobSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{{
									Image:      "same-image",
									Command:    []string{"cmd1", "cmd2"},
									Args:       []string{"arg1", "arg2"},
									WorkingDir: "/workdir",
									Env: []corev1.EnvVar{
										{Name: "VAR2", Value: "value2"},
										{Name: "VAR1", Value: "value1"},
									},
								}},
							},
						},
					},
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := helper.CompareTeardownJobs(tt.a, tt.b)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsOperationReady(t *testing.T) {
	tests := []struct {
		name      string
		operation *v1alpha1.Operation
		want      bool
	}{
		{
			name:      "nil operation",
			operation: nil,
			want:      false,
		},
		{
			name: "operation not ready",
			operation: &v1alpha1.Operation{
				Status: v1alpha1.OperationStatus{
					Phase: v1alpha1.OperationPhaseReconciling,
				},
			},
			want: false,
		},
		{
			name: "operation ready",
			operation: &v1alpha1.Operation{
				Status: v1alpha1.OperationStatus{
					Phase: v1alpha1.OperationPhaseReconciled,
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := helper.IsOperationReady(tt.operation)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestClearOperationConditions(t *testing.T) {
	t.Run("clear conditions", func(t *testing.T) {
		operation := &v1alpha1.Operation{
			Status: v1alpha1.OperationStatus{
				Conditions: []metav1.Condition{
					{
						Type:    "Test",
						Status:  "True",
						Reason:  "Testing",
						Message: "Test message",
					},
				},
			},
		}

		helper.ClearConditions(operation)
		assert.Empty(t, operation.Status.Conditions)
	})
}

func TestIsStringSliceIdentical(t *testing.T) {
	tests := []struct {
		name string
		a    []string
		b    []string
		want bool
	}{
		{
			name: "different length",
			a:    []string{"a", "b"},
			b:    []string{"a"},
			want: false,
		},
		{
			name: "same elements different order",
			a:    []string{"a", "b", "c"},
			b:    []string{"c", "a", "b"},
			want: true,
		},
		{
			name: "same elements same order",
			a:    []string{"a", "b", "c"},
			b:    []string{"a", "b", "c"},
			want: true,
		},
		{
			name: "different elements",
			a:    []string{"a", "b", "c"},
			b:    []string{"a", "b", "d"},
			want: false,
		},
		{
			name: "empty slices",
			a:    []string{},
			b:    []string{},
			want: true,
		},
		{
			name: "nil slices",
			a:    nil,
			b:    nil,
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create copies to verify original slices aren't modified
			aCopy := make([]string, len(tt.a))
			bCopy := make([]string, len(tt.b))
			if len(tt.a) > 0 {
				copy(aCopy, tt.a)
			}
			if len(tt.b) > 0 {
				copy(bCopy, tt.b)
			}

			got := isStringSliceIdentical(aCopy, bCopy)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsEnvVarsIdentical(t *testing.T) {
	tests := []struct {
		name string
		a    []corev1.EnvVar
		b    []corev1.EnvVar
		want bool
	}{
		{
			name: "different length",
			a: []corev1.EnvVar{
				{Name: "VAR1", Value: "value1"},
				{Name: "VAR2", Value: "value2"},
			},
			b: []corev1.EnvVar{
				{Name: "VAR1", Value: "value1"},
			},
			want: false,
		},
		{
			name: "same vars different order",
			a: []corev1.EnvVar{
				{Name: "VAR1", Value: "value1"},
				{Name: "VAR2", Value: "value2"},
			},
			b: []corev1.EnvVar{
				{Name: "VAR2", Value: "value2"},
				{Name: "VAR1", Value: "value1"},
			},
			want: true,
		},
		{
			name: "different values",
			a: []corev1.EnvVar{
				{Name: "VAR1", Value: "value1"},
				{Name: "VAR2", Value: "value2"},
			},
			b: []corev1.EnvVar{
				{Name: "VAR1", Value: "value1"},
				{Name: "VAR2", Value: "different"},
			},
			want: false,
		},
		{
			name: "different names",
			a: []corev1.EnvVar{
				{Name: "VAR1", Value: "value1"},
				{Name: "VAR2", Value: "value2"},
			},
			b: []corev1.EnvVar{
				{Name: "VAR1", Value: "value1"},
				{Name: "DIFFERENT", Value: "value2"},
			},
			want: false,
		},
		{
			name: "empty slices",
			a:    []corev1.EnvVar{},
			b:    []corev1.EnvVar{},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := helper.isEnvVarsIdentical(tt.a, tt.b)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsJobResultIdentical(t *testing.T) {
	tests := []struct {
		name string
		a    batchv1.JobSpec
		b    batchv1.JobSpec
		want bool
	}{
		{
			name: "both empty containers",
			a: batchv1.JobSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{},
					},
				},
			},
			b: batchv1.JobSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{},
					},
				},
			},
			want: true,
		},
		{
			name: "one empty containers",
			a: batchv1.JobSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{},
					},
				},
			},
			b: batchv1.JobSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{{
							Image: "test-image",
						}},
					},
				},
			},
			want: false,
		},
		{
			name: "different images",
			a: batchv1.JobSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{{
							Image: "test-image1",
						}},
					},
				},
			},
			b: batchv1.JobSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{{
							Image: "test-image2",
						}},
					},
				},
			},
			want: false,
		},
		{
			name: "different commands",
			a: batchv1.JobSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{{
							Image:   "test-image",
							Command: []string{"cmd1", "cmd2"},
						}},
					},
				},
			},
			b: batchv1.JobSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{{
							Image:   "test-image",
							Command: []string{"cmd1", "cmd3"},
						}},
					},
				},
			},
			want: false,
		},
		{
			name: "different args",
			a: batchv1.JobSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{{
							Image: "test-image",
							Args:  []string{"arg1", "arg2"},
						}},
					},
				},
			},
			b: batchv1.JobSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{{
							Image: "test-image",
							Args:  []string{"arg1", "arg3"},
						}},
					},
				},
			},
			want: false,
		},
		{
			name: "different working directory",
			a: batchv1.JobSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{{
							Image:      "test-image",
							WorkingDir: "/dir1",
						}},
					},
				},
			},
			b: batchv1.JobSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{{
							Image:      "test-image",
							WorkingDir: "/dir2",
						}},
					},
				},
			},
			want: false,
		},
		{
			name: "different env vars",
			a: batchv1.JobSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{{
							Image: "test-image",
							Env: []corev1.EnvVar{
								{Name: "VAR1", Value: "value1"},
							},
						}},
					},
				},
			},
			b: batchv1.JobSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{{
							Image: "test-image",
							Env: []corev1.EnvVar{
								{Name: "VAR1", Value: "different"},
							},
						}},
					},
				},
			},
			want: false,
		},
		{
			name: "identical jobs",
			a: batchv1.JobSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{{
							Image:      "test-image",
							Command:    []string{"cmd1", "cmd2"},
							Args:       []string{"arg1", "arg2"},
							WorkingDir: "/workdir",
							Env: []corev1.EnvVar{
								{Name: "VAR1", Value: "value1"},
								{Name: "VAR2", Value: "value2"},
							},
						}},
					},
				},
			},
			b: batchv1.JobSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{{
							Image:      "test-image",
							Command:    []string{"cmd1", "cmd2"},
							Args:       []string{"arg1", "arg2"},
							WorkingDir: "/workdir",
							Env: []corev1.EnvVar{
								{Name: "VAR2", Value: "value2"},
								{Name: "VAR1", Value: "value1"},
							},
						}},
					},
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := helper.isJobResultIdentical(tt.a, tt.b)
			assert.Equal(t, tt.want, got)
		})
	}
}
