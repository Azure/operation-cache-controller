package operation

import (
	"sort"
	"strings"

	"github.com/google/uuid"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"

	appsv1 "github.com/Azure/operation-cache-controller/api/v1"
)

// NewOperationId generates a new operation id which is an UUID.
func NewOperationId() string {
	return strings.Replace(uuid.New().String(), "-", "", -1)
}

// DiffAppDeployments returns the difference between two slices of AppDeployment.
func DiffAppDeployments(expected, actual []appsv1.AppDeployment,
	equals func(a, b appsv1.AppDeployment) bool) (added, removed, updated []appsv1.AppDeployment) {
	// Find added and updated AppDeployments.
	for _, e := range expected {
		found := false
		for _, a := range actual {
			if a.Name == e.Name {
				found = true
				if !equals(a, e) {
					updated = append(updated, a)
				}
				break
			}
		}
		if !found {
			added = append(added, e)
		}
	}

	// Find removed AppDeployments.
	for _, a := range actual {
		found := false
		for _, e := range expected {
			if a.Name == e.Name {
				found = true
				break
			}
		}
		if !found {
			removed = append(removed, a)
		}
	}

	return added, removed, updated
}

func isJobResultIdentical(a, b batchv1.JobSpec) bool {
	if len(a.Template.Spec.Containers) == 0 && len(b.Template.Spec.Containers) == 0 {
		return true
	}
	if len(a.Template.Spec.Containers) == 0 || len(b.Template.Spec.Containers) == 0 {
		return false
	}
	if a.Template.Spec.Containers[0].Image != b.Template.Spec.Containers[0].Image {
		return false
	}
	if !isStringSliceIdentical(a.Template.Spec.Containers[0].Command, b.Template.Spec.Containers[0].Command) {
		return false
	}
	if !isStringSliceIdentical(a.Template.Spec.Containers[0].Args, b.Template.Spec.Containers[0].Args) {
		return false
	}
	if a.Template.Spec.Containers[0].WorkingDir != b.Template.Spec.Containers[0].WorkingDir {
		return false
	}
	if !isEnvVarsIdentical(a.Template.Spec.Containers[0].Env, b.Template.Spec.Containers[0].Env) {
		return false
	}

	return true
}

func isEnvVarsIdentical(a, b []corev1.EnvVar) bool {
	if len(a) != len(b) {
		return false
	}
	// sort environment variables to ensure consistent hashing
	sort.Slice(a, func(i, j int) bool {
		return a[i].Name < a[j].Name
	})
	sort.Slice(b, func(i, j int) bool {
		return b[i].Name < b[j].Name
	})
	for i, env := range a {
		if env.Name != b[i].Name || env.Value != b[i].Value {
			return false
		}
	}
	return true
}

func isStringSliceIdentical(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	sort.Strings(a)
	sort.Strings(b)
	for i, dep := range a {
		if dep != b[i] {
			return false
		}
	}
	return true
}

func CompareProvisionJobs(a, b appsv1.AppDeployment) bool {
	if sameProvisionJob := isJobResultIdentical(a.Spec.Provision, b.Spec.Provision); !sameProvisionJob {
		return false
	}
	// sort dependencies to ensure consistent hashing
	sort.Strings(a.Spec.Dependencies)
	sort.Strings(b.Spec.Dependencies)
	for i, dep := range a.Spec.Dependencies {
		if dep != b.Spec.Dependencies[i] {
			return false
		}
	}

	return true
}

func CompareTeardownJobs(a, b appsv1.AppDeployment) bool {
	if sameTeardownJob := isJobResultIdentical(a.Spec.Teardown, b.Spec.Teardown); !sameTeardownJob {
		return false
	}
	return true
}
