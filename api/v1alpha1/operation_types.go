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
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	OperationOwnerKey = ".operation.metadata.controller"

	OperationFinalizerName         = "finalizer.operation.controller.azure.com"
	OperationAcquiredAnnotationKey = "operation.controller.azure.com/acquired"

	OperationPhaseEmpty       = ""
	OperationPhaseReconciling = "Reconciling"
	OperationPhaseReconciled  = "Reconciled"
	OperationPhaseDeleting    = "Deleting"
	OperationPhaseDeleted     = "Deleted"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type ApplicationSpec struct {
	Name      string          `json:"name"`
	Provision batchv1.JobSpec `json:"provision"`
	Teardown  batchv1.JobSpec `json:"teardown"`
	// +kubebuilder:validation:Optional
	Dependencies []string `json:"dependencies,omitempty"`
}

// OperationSpec defines the desired state of Operation.
type OperationSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinItems=1
	Applications []ApplicationSpec `json:"applications"`
	// +kubebuilder:validation:optional
	// +kubebuilder:validation:Pattern:=`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$`
	ExpireAt string `json:"expireAt,omitempty"`
}

// OperationStatus defines the observed state of Operation.
type OperationStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Conditions is a list of conditions to describe the status of the deploy
	Conditions  []metav1.Condition `json:"conditions"`
	Phase       string             `json:"phase"`
	CacheKey    string             `json:"cacheKey"`
	OperationID string             `json:"operationId"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Phase",type="string",JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Key",type="string",JSONPath=`.status.cacheKey`

// Operation is the Schema for the operations API.
type Operation struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OperationSpec   `json:"spec,omitempty"`
	Status OperationStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// OperationList contains a list of Operation.
type OperationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Operation `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Operation{}, &OperationList{})
}
