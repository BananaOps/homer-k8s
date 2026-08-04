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

// HomerPageSpec defines the desired state of HomerPage
type HomerPageSpec struct {
	// PageName is the name of the page file (without extension).
	// Homer will load this page when the user navigates to "#<pageName>".
	// Example: "page2" will create the file "page2.yml" and be accessible via "#page2".
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern=`^[a-z0-9]([a-z0-9\-]*[a-z0-9])?$`
	PageName string `json:"pageName"`

	// Groups is the list of service groups displayed on this page.
	Groups []Group `json:"groups,omitempty"`

	// AutoLink controls whether a link to this page is automatically added to the
	// main config.yml navbar links. Defaults to false.
	// +kubebuilder:default=false
	AutoLink bool `json:"autoLink,omitempty"`

	// LinkName is the display name used in the navbar when AutoLink is true.
	// Defaults to the PageName if not set.
	// +optional
	LinkName string `json:"linkName,omitempty"`

	// LinkIcon is the FontAwesome icon used in the navbar when AutoLink is true.
	// Example: "fas fa-file-alt". See https://fontawesome.com/v5/search for options.
	// +optional
	LinkIcon string `json:"linkIcon,omitempty"`
}

// HomerPageStatus defines the observed state of HomerPage
type HomerPageStatus struct {
	// Ready indicates whether the page file has been successfully written.
	// +optional
	Ready bool `json:"ready,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Page",type="string",JSONPath=".spec.pageName"
//+kubebuilder:printcolumn:name="AutoLink",type="boolean",JSONPath=".spec.autoLink"
//+kubebuilder:printcolumn:name="Ready",type="boolean",JSONPath=".status.ready"

// HomerPage is the Schema for the homerpages API.
// A HomerPage creates a dedicated Homer page file (<pageName>.yml) in the assets
// directory, allowing Homer to serve multiple dashboard pages.
type HomerPage struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HomerPageSpec   `json:"spec,omitempty"`
	Status HomerPageStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// HomerPageList contains a list of HomerPage
type HomerPageList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HomerPage `json:"items"`
}

func init() {
	SchemeBuilder.Register(&HomerPage{}, &HomerPageList{})
}
