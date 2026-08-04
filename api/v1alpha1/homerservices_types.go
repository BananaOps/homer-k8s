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
	// Type of the service in homer dashboard
	Type string `json:"type,omitempty"`
	// Text to copy when type is clipboard
	Clipboard string `json:"clipboard,omitempty"`
	// Optional color for card to set color directly without custom stylesheet
	Background string `json:"background,omitempty"`
	// Optional API key used to fetch service data when necessary.
	ApiKey string `json:"apikey,omitempty"`
	// Optional API version used to fetch service data when necessary.
	ApiVersion string `json:"apiVersion,omitempty"`
	// Optional check interval in seconds to fetch service data when necessary.
	CheckInterval int `json:"checkInterval,omitempty"`
	// Optional - text or bar. Default to text.
	Display string `json:"display,omitempty"`
	// Optional download interval
	DownloadInterval int `json:"downloadInterval,omitempty"`
	// Optional alternative base URL used to fetch service data when necessary.
	Endpoint string `json:"endpoint,omitempty"`
	// Optional environments variables used to fetch service data when necessary.
	Environments []string `json:"environments,omitempty"`
	// Optional interval to fetch service data when necessary.
	Interval int `json:"interval,omitempty"`
	// Displays Home Assistant instance status, version, location, and entity count
	Items []string `json:"items,omitempty"`
	// Optional library type for emby/jellyfin
	LibraryType string `json:"libraryType,omitempty"`
	// Optional limit for number of items to fetch when necessary.
	Limit int `json:"limit,omitempty"`
	// Optional location for weather
	Location string `json:"location,omitempty"`
	// Optional location id for weather
	LocationId string `json:"locationId,omitempty"`
	// Optional password used to fetch service data when necessary.
	Password string `json:"password,omitempty"`
	// Optional query used to fetch service data when necessary.
	Query string `json:"query,omitempty"`
	// Optional rate interval
	RateInterval int `json:"rateInterval,omitempty"`
	// Optional separator for home assistant
	Separator string `json:"separator,omitempty"`
	// Optional slug used to fetch service data when necessary.
	Slug string `json:"slug,omitempty"`
	// Optional - either status or counter for miniflux
	Style string `json:"style,omitempty"`
	// Optional token used to fetch service data when necessary.
	Token string `json:"token,omitempty"`
	// Optional torrent interval
	TorrentInterval int `json:"torrentInterval,omitempty"`
	// Optional units for weather
	Units string `json:"units,omitempty"`
	// Optional update interval in seconds to fetch service data when necessary.
	UpdateInterval int `json:"updateInterval,omitempty"`
	// Optional username used to fetch service data when necessary.
	Username string `json:"username,omitempty"`
	// Optional xmlrpc used to fetch service data when necessary.
	Xmlrpc string `json:"xmlrpc,omitempty"`
}
