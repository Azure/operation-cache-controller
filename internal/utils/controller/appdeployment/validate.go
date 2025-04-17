package appdeployment

import (
	"errors"
	"fmt"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"

	"github.com/Azure/operation-cache-controller/api/v1alpha1"
)

type Validater func(*v1alpha1.AppDeployment) error

func Validate(ap *v1alpha1.AppDeployment) error {
	var errs error
	validaters := []Validater{
		validateJobSpec,
	}
	for _, v := range validaters {
		errs = errors.Join(errs, v(ap))
	}
	return errs
}

// validateJobSpec validates the container count in the AppDeployment Spec
// * container count in AppDeployment Spec should be 1
// * initCountainer is not allowed
func validateJobSpec(ap *v1alpha1.AppDeployment) error {
	if equality.Semantic.DeepEqual(ap.Spec, v1alpha1.AppDeploymentSpec{}) {
		return errors.New("spec of appdeployment is nil")
	}
	// provision job must be present
	if equality.Semantic.DeepEqual(ap.Spec.Provision, batchv1.JobSpec{}) {
		return errors.New("provision job is nil")
	}
	if jobConstraint(ap.Spec.Provision) != nil {
		return fmt.Errorf("provision: %w", jobConstraint(ap.Spec.Provision))
	}
	// teardown job is optional
	if !equality.Semantic.DeepEqual(ap.Spec.Teardown, batchv1.JobSpec{}) {
		if jobConstraint(ap.Spec.Teardown) != nil {
			return fmt.Errorf("teardown: %w", jobConstraint(ap.Spec.Teardown))
		}
	}
	return nil
}

// jobConstraint validates the JobSpec
// only container is allowed in the JobSpec
func jobConstraint(js batchv1.JobSpec) error {
	if js.ActiveDeadlineSeconds != nil {
		return fmt.Errorf("activeDeadlineSeconds is not allowed")
	}
	if js.BackoffLimit != nil {
		return fmt.Errorf("backoffLimit is not allowed")
	}
	if js.BackoffLimitPerIndex != nil {
		return fmt.Errorf("backoffLimitPerIndex is not allowed")
	}
	if js.Completions != nil {
		return fmt.Errorf("completions is not allowed")
	}
	if js.CompletionMode != nil {
		return fmt.Errorf("completionMode is not allowed")
	}
	if js.ManagedBy != nil {
		return fmt.Errorf("managedBy is not allowed")
	}
	if js.ManualSelector != nil {
		return fmt.Errorf("manualSelector is not allowed")
	}
	if js.MaxFailedIndexes != nil {
		return fmt.Errorf("maxFailedIndexes is not allowed")
	}
	if js.Parallelism != nil {
		return fmt.Errorf("parallelism is not allowed")
	}
	if js.PodFailurePolicy != nil {
		return fmt.Errorf("podFailurePolicy is not allowed")
	}
	if js.PodReplacementPolicy != nil {
		return fmt.Errorf("podReplacementPolicy is not allowed")
	}
	if js.Selector != nil {
		return fmt.Errorf("selector is not allowed")
	}
	if js.TTLSecondsAfterFinished != nil {
		return fmt.Errorf("ttlSecondsAfterFinished is not allowed")
	}
	if js.SuccessPolicy != nil {
		return fmt.Errorf("successPolicy is not allowed")
	}
	if js.Suspend != nil {
		return fmt.Errorf("suspend is not allowed")
	}

	if podConstraint(js.Template) != nil {
		return fmt.Errorf("pod template: %w", podConstraint(js.Template))
	}

	return nil
}

// podConstraint validates the PodSpec
func podConstraint(pt corev1.PodTemplateSpec) error {
	if len(pt.Name) > 0 {
		return fmt.Errorf("name is not allowed")
	}
	if len(pt.Namespace) > 0 {
		return fmt.Errorf("namespace is not allowed")
	}
	if len(pt.Spec.Volumes) > 0 {
		return fmt.Errorf("volumes are not allowed")
	}
	if len(pt.Spec.InitContainers) > 0 {
		return fmt.Errorf("initContainers are not allowed")
	}
	if len(pt.Spec.Containers) != 1 {
		return fmt.Errorf("container count should be 1")
	}
	if containerConstraint(pt.Spec.Containers[0]) != nil {
		return fmt.Errorf("container: %w", containerConstraint(pt.Spec.Containers[0]))
	}
	return nil
}

// containerConstraint validates the Container
func containerConstraint(c corev1.Container) error {
	if c.Image == "" {
		return fmt.Errorf("image is empty")
	}
	if len(c.VolumeMounts) > 0 {
		return fmt.Errorf("volumeMounts are not allowed")
	}

	return nil
}
