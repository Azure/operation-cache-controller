package appdeployment

import (
	"context"
	"testing"

	"github.com/Azure/operation-cache-controller/api/v1alpha1"
	"github.com/stretchr/testify/assert"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCheckJobStatus(t *testing.T) {
	tests := []struct {
		name     string
		job      *batchv1.Job
		expected JobStatus
	}{
		{
			name: "Should return succeeded when job is successful",
			job: &batchv1.Job{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-job",
					Namespace: "test-namespace",
					Labels: map[string]string{
						"app.kubernetes.io/name": "test-app-name",
					},
					Annotations: map[string]string{
						"app.kubernetes.io/version": "test-app-version",
					},
				},
				Status: batchv1.JobStatus{
					Failed:    0,
					Succeeded: 1,
				},
			},
			expected: JobStatusSucceeded,
		},
		{
			name: "Should return failed when job has failed",
			job: &batchv1.Job{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-job",
					Namespace: "test-namespace",
					Labels: map[string]string{
						"app.kubernetes.io/name": "test-app-name",
					},
					Annotations: map[string]string{
						"app.kubernetes.io/version": "test-app-version",
					},
				},
				Status: batchv1.JobStatus{
					Failed:    1,
					Succeeded: 0,
				},
			},
			expected: JobStatusFailed,
		},
		{
			name: "Should return in progress when job is still running",
			job: &batchv1.Job{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-job",
					Namespace: "test-namespace",
					Labels: map[string]string{
						"app.kubernetes.io/name": "test-app-name",
					},
					Annotations: map[string]string{
						"app.kubernetes.io/version": "test-app-version",
					},
				},
				Status: batchv1.JobStatus{
					Failed:    0,
					Succeeded: 0,
				},
			},
			expected: JobStatusRunning,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := CheckJobStatus(context.Background(), tt.job)
			assert.Equal(t, tt.expected, res)
		})
	}
}
func TestGetProvisionJobName(t *testing.T) {
	tests := []struct {
		name     string
		appName  string
		opId     string
		expected string
	}{
		{
			name:     "Valid app name and operation ID",
			appName:  "my-app",
			opId:     "op123",
			expected: "provision-my-app-op123",
		},
		{
			name:     "App name exceeds max length",
			appName:  "a-very-long-application-name-exceeding-limit",
			opId:     "op123",
			expected: "provision-a-very-long-application-name-exceedi-op123",
		},
		{
			name:     "Operation ID exceeds max length",
			appName:  "my-app",
			opId:     "operation-id-exceeding-length",
			expected: "provision-my-app-operation-id-excee",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appDeployment := &v1alpha1.AppDeployment{
				ObjectMeta: metav1.ObjectMeta{
					Name: tt.appName,
				},
				Spec: v1alpha1.AppDeploymentSpec{
					OpId: tt.opId,
				},
			}
			res := GetProvisionJobName(appDeployment)
			assert.Equal(t, tt.expected, res)
		})
	}
}

func TestGetTeardownJobName(t *testing.T) {
	tests := []struct {
		name     string
		appName  string
		opId     string
		expected string
	}{
		{
			name:     "Valid app name and operation ID",
			appName:  "my-app",
			opId:     "op123",
			expected: "teardown-my-app-op123",
		},
		{
			name:     "App name exceeds max length",
			appName:  "a-very-long-application-name-exceeding-limit",
			opId:     "op123",
			expected: "teardown-a-very-long-application-name-exceedi-op123",
		},
		{
			name:     "Operation ID exceeds max length",
			appName:  "my-app",
			opId:     "operation-id-exceeding-length",
			expected: "teardown-my-app-operation-id-excee",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appDeployment := &v1alpha1.AppDeployment{
				ObjectMeta: metav1.ObjectMeta{
					Name: tt.appName,
				},
				Spec: v1alpha1.AppDeploymentSpec{
					OpId: tt.opId,
				},
			}
			res := GetTeardownJobName(appDeployment)
			assert.Equal(t, tt.expected, res)
		})
	}
}
func TestOperationScopedAppDeployment(t *testing.T) {
	tests := []struct {
		name     string
		appName  string
		opId     string
		expected string
	}{
		{
			name:     "Both appName and opId within limits",
			appName:  "my-app",
			opId:     "op123",
			expected: "op123-my-app",
		},
		{
			name:     "appName exceeds max length",
			appName:  "a-very-long-application-name-exceeding-limit",
			opId:     "op123",
			expected: "op123-a-very-long-application-name-exceedi",
		},
		{
			name:     "opId exceeds max length",
			appName:  "my-app",
			opId:     "operation-id-exceeding-length",
			expected: "operation-id-excee-my-app",
		},
		{
			name:     "Both appName and opId exceed max length",
			appName:  "another-very-long-application-name-exceeding-limit",
			opId:     "another-operation-id-exceeding-length",
			expected: "another-operation--another-very-long-application-name-e",
		},
		{
			name:     "Empty appName and opId",
			appName:  "",
			opId:     "",
			expected: "",
		},
		{
			name:     "Empty appName",
			appName:  "",
			opId:     "op123",
			expected: "op123-",
		},
		{
			name:     "Empty opId",
			appName:  "my-app",
			opId:     "",
			expected: "my-app",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := OperationScopedAppDeployment(tt.appName, tt.opId)
			assert.Equal(t, tt.expected, res)
		})
	}
}
