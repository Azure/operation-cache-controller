package controller

import (
	"context"
	"strings"
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
			appName:  "op123-my-app",
			expected: "provision-op123-my-app",
		},
		{
			name:     "App name exceeds max length",
			appName:  "op1234567890-a-very-long-application-name-exceeding-limit",
			expected: "provision-op1234567890-a-very-long-application-name-exceeding-l",
		},
		{
			name:     "Operation ID exceeds max length",
			appName:  "operationid1234567890123456789012345678901234567890-my-application",
			expected: "provision-operationid1234567890123456789012345678901234567890-m",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appDeployment := &v1alpha1.AppDeployment{
				ObjectMeta: metav1.ObjectMeta{
					Name: tt.appName,
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
			expected: "teardown-my-app",
		},
		{
			name:     "App name exceeds max length",
			appName:  "op1234567890-a-very-long-application-name-exceeding-limit",
			expected: "teardown-op1234567890-a-very-long-application-name-exceeding-li",
		},
		{
			name:     "Operation ID exceeds max length",
			appName:  "operationid1234567890123456789012345678901234567890-my-application",
			expected: "teardown-operationid1234567890123456789012345678901234567890-my",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appDeployment := &v1alpha1.AppDeployment{
				ObjectMeta: metav1.ObjectMeta{
					Name: tt.appName,
				},
			}
			res := GetTeardownJobName(appDeployment)
			assert.Equal(t, tt.expected, res)
		})
	}
}
func TestOperationScopedAppDeployment(t *testing.T) {
	tests := []struct {
		name    string
		appName string
		opId    string
		want    string
	}{
		{
			name:    "normal case - combined length within limit",
			appName: "my-app",
			opId:    "op-12345",
			want:    "op-12345-my-app",
		},
		{
			name:    "app name truncated when too long",
			appName: strings.Repeat("a", 60), // 60 chars, exceeds MaxAppNameLength
			opId:    "op-12345",
			want:    "op-12345-" + strings.Repeat("a", MaxAppNameLength),
		},
		{
			name:    "operation ID truncated when combined length exceeds limit",
			appName: strings.Repeat("a", MaxAppNameLength),
			opId:    strings.Repeat("b", 30), // This will exceed when combined
			want:    strings.Repeat("b", MaxResourceNameLength-MaxAppNameLength-1) + "-" + strings.Repeat("a", MaxAppNameLength),
		},
		{
			name:    "both truncated for very long inputs",
			appName: strings.Repeat("a", 100),
			opId:    strings.Repeat("b", 100),
			want:    strings.Repeat("b", MaxResourceNameLength-MaxAppNameLength-1) + "-" + strings.Repeat("a", MaxAppNameLength),
		},
		{
			name:    "empty app name",
			appName: "",
			opId:    "op-12345",
			want:    "op-12345-",
		},
		{
			name:    "empty operation ID",
			appName: "my-app",
			opId:    "",
			want:    "-my-app",
		},
		{
			name:    "both empty",
			appName: "",
			opId:    "",
			want:    "-",
		},
		{
			name:    "exact max length",
			appName: "app",
			opId:    strings.Repeat("x", MaxResourceNameLength-4), // -4 for "-app"
			want:    strings.Repeat("x", MaxResourceNameLength-4) + "-app",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := OperationScopedAppDeployment(tt.appName, tt.opId)
			assert.Equal(t, tt.want, got, "OperationScopedAppDeployment() mismatch for %s", tt.name)
		})
	}
}

func TestOperationScopedAppDeployment_LengthConstraints(t *testing.T) {
	// Test that the function always respects the maximum length constraint
	testCases := []struct {
		appNameLen int
		opIdLen    int
	}{
		{10, 10},
		{MaxAppNameLength, 10},
		{50, 50},
		{100, 100},
		{MaxResourceNameLength, MaxResourceNameLength},
	}

	for _, tc := range testCases {
		appName := strings.Repeat("a", tc.appNameLen)
		opId := strings.Repeat("b", tc.opIdLen)

		result := OperationScopedAppDeployment(appName, opId)

		if len(result) > MaxResourceNameLength {
			t.Errorf("Result length %d exceeds MaxResourceNameLength %d for appName length %d and opId length %d",
				len(result), MaxResourceNameLength, tc.appNameLen, tc.opIdLen)
		}
	}
}
