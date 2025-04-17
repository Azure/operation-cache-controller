package operation

import (
	"testing"

	"github.com/stretchr/testify/assert"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1alpha1 "github.com/Azure/operation-cache-controller/api/v1alpha1"
)

func TestNewOperationId(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "basic"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id := NewOperationId()
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
			added, removed, updated := DiffAppDeployments(tt.expected, tt.actual, tt.equals)
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CompareProvisionJobs(tt.a, tt.b)
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CompareTeardownJobs(tt.a, tt.b)
			assert.Equal(t, tt.want, got)
		})
	}
}
