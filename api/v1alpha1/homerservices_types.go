/*
Copyright 2024.

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

// HomerServicesSpec defines the desired state of HomerServices
type HomerServicesSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of HomerServices. Edit homerservices_types.go to remove/update
	Groups []Group `json:"groups,omitempty"`
}

// HomerServicesStatus defines the observed state of HomerServices
type HomerServicesStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// HomerServices is the Schema for the homerservices API
type HomerServices struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HomerServicesSpec   `json:"spec,omitempty"`
	Status HomerServicesStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// HomerServicesList contains a list of HomerServices
type HomerServicesList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HomerServices `json:"items"`
}

func init() {
	SchemeBuilder.Register(&HomerServices{}, &HomerServicesList{})
}

type Group struct {
	Name string `json:"name,omitempty"`
	Icon  string `json:"icon,omitempty"`
	Items  []Item `json:"items,omitempty"`
}

type Item struct {
	Name string `json:"name,omitempty"`
	Icon  string `json:"icon,omitempty"`
	Logo  string `json:"logo,omitempty"`
	TagStyle  string `json:"tagstyle,omitempty"`
	SubTitle  string `json:"subtitle,omitempty"`
	Tag  string `json:"tag,omitempty"`
	Keyword  string `json:"keywords,omitempty"`
	Url  string `json:"url,omitempty"`
	Target  string `json:"target,omitempty"`
	Background  string `json:"background,omitempty"`
}

