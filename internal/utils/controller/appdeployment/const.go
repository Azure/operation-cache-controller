package appdeployment

const (
	FinalizerName = "finalizer.appdeployment.devinfra.goms.io"

	// phase types
	PhaseEmpty     = ""
	PhasePending   = "Pending"
	PhaseDeploying = "Deploying"
	PhaseReady     = "Ready"
	PhaseDeleting  = "Deleting"
	PhaseDeleted   = "Deleted"

	// log keys
	LogKeyJobName           = "jobName"
	LogKeyAppDeploymentName = "appDeploymentName"
)
