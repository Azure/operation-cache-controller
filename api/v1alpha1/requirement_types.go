/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	RequirementOwnerKey = ".requirement.metadata.controller"

	RequirementFinalizerName = "finalizer.requirement.devinfra.goms.io"

	RequirementConditionRequirementInitialized  = "RequirementInitialized"
	RequirementConditionCacheResourceFound      = "CacheCRFound"
	RequirementConditionCachedOperationAcquired = "CachedOpAcquired"
	RequirementConditionOperationReady          = "OperationReady"

	RequirementConditionReasonNoOperationAvailable = "NoOperationAvailable"
	RequirementConditionReasonCacheCRNotFound      = "CacheCRNotFound"
	RequirementConditionReasonCacheCRFound         = "CacheCRFound"
	RequirementConditionReasonCacheHit             = "CacheHit"
	RequirementConditionReasonCacheMiss            = "CacheMiss"

	RequirementPhaseEmpty         = ""
	RequirementPhaseCacheChecking = "CacheChecking"
	RequirementPhaseOperating     = "Operating"
	RequirementPhaseReady         = "Ready"
	RequirementPhaseDeleted       = "Deleted"
	RequirementPhaseDeleting      = "Deleting"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// RequirementSpec defines the desired state of Requirement.
type RequirementSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// +kubebuilder:validation:Required
	Template OperationSpec `json:"template"`
	// +kubebuilder:validation:Required
	EnableCache bool `json:"enableCache"`
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Pattern:=`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$`
	ExpireAt string `json:"expireAt,omitempty"`
}

// RequirementStatus defines the observed state of Requirement.
type RequirementStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	OperationId   string             `json:"operationId"`
	OperationName string             `json:"operationName"`
	CacheKey      string             `json:"originalCacheKey"`
	Phase         string             `json:"phase"`
	Conditions    []metav1.Condition `json:"conditions"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Phase",type="string",JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="OperationId",type="string",JSONPath=`.status.operationId`

// Requirement is the Schema for the requirements API.
type Requirement struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RequirementSpec   `json:"spec,omitempty"`
	Status RequirementStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// RequirementList contains a list of Requirement.
type RequirementList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Requirement `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Requirement{}, &RequirementList{})
}
