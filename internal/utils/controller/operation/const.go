package operation

const (
	FinalizerName         = "finalizer.operation.controller.azure.com"
	AcquiredAnnotationKey = "operation.controller.azure.com/acquired"

	// PhaseEmpty is the phase when the operation is empty.
	PhaseEmpty = ""
	// PhaseReconciling is the phase when the operation is reconciling.
	PhaseReconciling = "Reconciling"
	// PhaseReconciled is the phase when the operation is reconciled.
	PhaseReconciled = "Reconciled"
	// Deletion is the phase when the operation is deleting.
	PhaseDeleting = "Deleting"
	// PhaseDeleted is the phase when the operation is deleted.
	PhaseDeleted = "Deleted"
)
