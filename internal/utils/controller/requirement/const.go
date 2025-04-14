package requirement

const (
	FinalizerName = "finalizer.requirement.devinfra.goms.io"

	// RequirementInitialized is the condition type for Requirement initialized
	ConditionRequirementInitialized = "RequirementInitialized"
	// Cache CRD for Requirement is existed
	ConditionCacheResourceFound = "CacheCRFound"
	// no available Operation CRD for Requirement in cache, create one
	ConditionCachedOperationAcquired = "CachedOpAcquired"
	// Operation CRD for Requirement is ready
	ConditionOperationReady = "OperationReady"

	ConditionReasonNoOperationAvailable = "NoOperationAvailable"
	ConditionReasonCacheCRNotFound      = "CacheCRNotFound"
	ConditionReasonCacheCRFound         = "CacheCRFound"
	ConditionReasonCacheHit             = "CacheHit"
	ConditionReasonCacheMiss            = "CacheMiss"

	PhaseEmpty         = ""
	PhaseCacheChecking = "CacheChecking"
	PhaseOperating     = "Operating"
	PhaseReady         = "Ready"
	PhaseDeleted       = "Deleted"
	PhaseDeleting      = "Deleting"
)
