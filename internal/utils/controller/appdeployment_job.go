package controller

import (
	"context"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/Azure/operation-cache-controller/api/v1alpha1"
)

const (
	JobTypeProvision = "provision"
	JobTypeTeardown  = "teardown"
	JobNamePattern   = "%s-%s-%s" // jobType-appName-operationId
)

var (
	backOffLimit            int32 = 10
	ttlSecondsAfterFinished int32 = 3600
	MaxAppNameLength        int   = 36
)

func validJobName(appName, jobType string) string {
	originName := jobType + "-" + appName
	if len(originName) > MaxResourceNameLength {
		return originName[:MaxResourceNameLength]
	}
	return originName
}

func ProvisionJobFromAppDeploymentSpec(appDeployment *v1alpha1.AppDeployment) *batchv1.Job {
	return jobFromAppDeploymentSpec(appDeployment, JobTypeProvision)
}

func TeardownJobFromAppDeploymentSpec(appDeployment *v1alpha1.AppDeployment) *batchv1.Job {
	return jobFromAppDeploymentSpec(appDeployment, JobTypeTeardown)
}

func GetProvisionJobName(appDeployment *v1alpha1.AppDeployment) string {
	return validJobName(appDeployment.Name, JobTypeProvision)
}

func GetTeardownJobName(appDeployment *v1alpha1.AppDeployment) string {
	return validJobName(appDeployment.Name, JobTypeTeardown)
}

func jobFromAppDeploymentSpec(appDeployment *v1alpha1.AppDeployment, suffix string) *batchv1.Job {
	ops := jobOptions{
		name:        validJobName(appDeployment.Name, suffix),
		namespace:   appDeployment.Namespace,
		labels:      appDeployment.Labels,
		jobSpec:     appDeployment.Spec.Provision,
		operationID: appDeployment.Spec.OpId,
	}
	if suffix == JobTypeTeardown {
		ops.jobSpec = appDeployment.Spec.Teardown
	}
	return newJobWithOptions(ops)
}

type jobOptions struct {
	name        string
	namespace   string
	labels      map[string]string
	annotations map[string]string
	ownerRefs   []metav1.OwnerReference
	jobSpec     batchv1.JobSpec
	operationID string
}

func newJobWithOptions(options jobOptions) *batchv1.Job {
	options.jobSpec.Template.Spec.Containers[0].Env = append(options.jobSpec.Template.Spec.Containers[0].Env, corev1.EnvVar{
		Name:  OperationIDEnvKey,
		Value: options.operationID,
	})
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:        options.name,
			Namespace:   options.namespace,
			Labels:      options.labels,
			Annotations: options.annotations,
		},
		Spec: options.jobSpec,
	}
	job.Spec.Template.Spec.RestartPolicy = corev1.RestartPolicyOnFailure
	job.Spec.BackoffLimit = &backOffLimit
	job.Spec.TTLSecondsAfterFinished = &ttlSecondsAfterFinished

	if len(options.ownerRefs) > 0 {
		job.OwnerReferences = options.ownerRefs
	}

	return job
}

type JobStatus string

var (
	JobStatusSucceeded JobStatus = "Succeeded"
	JobStatusFailed    JobStatus = "Failed"
	JobStatusRunning   JobStatus = "Running"
)

func CheckJobStatus(ctx context.Context, job *batchv1.Job) JobStatus {
	if job.Status.Succeeded > 0 {
		return JobStatusSucceeded
	}
	if job.Status.Failed > 0 {
		return JobStatusFailed
	}
	return JobStatusRunning
}

func OperationScopedAppDeployment(appName, opId string) string {
	originalName := opId + "-" + appName
	if len(originalName) < MaxResourceNameLength {
		return originalName
	}
	if len(appName) > MaxAppNameLength {
		appName = appName[:MaxAppNameLength]
	}
	resisualLength := MaxResourceNameLength - len(appName) - 1 // -1 for the hyphen
	if len(opId) > resisualLength {
		opId = opId[:resisualLength]
	}
	return opId + "-" + appName
}
