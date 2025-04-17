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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// CacheSpec defines the desired state of Cache.
type CacheSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	OperationTemplate OperationSpec `json:"operationTemplate"`

	// Strategy is the cache strategy
	// +kubebuilder:validation:optional
	Strategy string `json:"strategy,omitempty"`

	// ExpireTime is the RFC3339-format time when the cache will be expired. If not set, the cache is never expired.
	// +kubebuilder:validation:optional
	// +kubebuilder:validation:Pattern:=`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$`
	ExpireTime string `json:"expireTime,omitempty"`
}

// CacheStatus defines the observed state of Cache.
type CacheStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	CacheKey        string   `json:"cacheKey"`
	KeepAliveCount  int32    `json:"keepAlive"`
	AvailableCaches []string `json:"availableCaches,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Cache is the Schema for the caches API.
type Cache struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CacheSpec   `json:"spec,omitempty"`
	Status CacheStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CacheList contains a list of Cache.
type CacheList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Cache `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Cache{}, &CacheList{})
}
