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

	// Groups is a map of group for items service
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
	// Name of group in homer dashboard
	Name string `json:"name,omitempty"`
	// Icon of group in homer dashboard, See https://fontawesome.com/v5/search for icons options
	Icon string `json:"icon,omitempty"`
	// Items is map of services in homer dashboard
	Items []Item `json:"items,omitempty"`
}

type Item struct {
	// Name of service in homer dashboard
	Name string `json:"name,omitempty"`
	// Icon of service in homer dashboard, See https://fontawesome.com/v5/search for icons options
	Icon string `json:"icon,omitempty"`
	// A path to an image can also be provided. Note that icon take precedence if both icon and logo are set.
	Logo string `json:"logo,omitempty"`
	// Tagstyle is the style of the tag in homer dashboard, See https://github.com/bastienwirtz/homer/blob/main/docs/configuration.md#style-options	for style options
	TagStyle string `json:"tagstyle,omitempty"`
	// SubTitle of the service in homer dashboard
	SubTitle string `json:"subtitle,omitempty"`
	// Tag of the service in homer dashboard
	Tag string `json:"tag,omitempty"`
	// Keywords of the service in homer dashboard
	Keyword string `json:"keywords,omitempty"`
	// Url of the service in homer dashboard
	Url string `json:"url,omitempty"`
	// Target of the service in homer dashboard
	Target string `json:"target,omitempty"`
	// Type of the service in homer dashboard. See https://github.com/bastienwirtz/homer/blob/64629742f7aec7b34e8fc9e6b83282e496c2fb74/docs/configuration.md?plain=1#L150
	Type string `json:"type,omitempty"`
	Clipboard string `json:"clipboard,omitempty"` 
	// Optional color for card to set color directly without custom stylesheet
	Background string `json:"background,omitempty"`
}
