/*
Copyright 2023.

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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func init() {
	SchemeBuilder.Register(&Integration{}, &IntegrationList{})
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Integration is the Schema for the integrations API
type Integration struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   IntegrationSpec   `json:"spec,omitempty"`
	Status IntegrationStatus `json:"status,omitempty"`
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// IntegrationSpec defines the desired state of Integration
type IntegrationSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Value is immutable"
	Name string `json:"name,omitempty"`

	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Value is immutable"
	ID           string    `json:"id,omitempty"`
	Type         string    `json:"type"`
	Templates    Templates `json:"templates,omitempty"`
	DefaultRoute Route     `json:"default_route,omitempty"`
	Routes       []Route   `json:"routes,omitempty"`
}

// TODO: add remaining templates: sms, slack, phone_call, telegram
type Templates struct {
	GroupingKey       string      `json:"grouping_key,omitempty"`
	ResolveSignal     string      `json:"resolve_signal,omitempty"`
	AcknowledgeSignal string      `json:"acknowledge_signal,omitempty"`
	SourceLink        string      `json:"source_link,omitempty"`
	Web               TemplateWeb `json:"web,omitempty"`
}

type TemplateWeb struct {
	Title    string `json:"title,omitempty"`
	Message  string `json:"message,omitempty"`
	ImageURL string `json:"image_url,omitempty"`
}

// TODO: add slack setting for routes
type Route struct {

	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Value is immutable"
	ID              string `json:"id,omitempty"`
	EscalationChain string `json:"escalation_chain,omitempty"`
	RoutingRegex    string `json:"routing_regex,omitempty"`
	Position        int    `json:"position"`
}

// IntegrationStatus defines the observed state of Integration
type IntegrationStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	HttpEndpoint string   `json:"http_endpoint"`
	RouteIDs     []string `json:"route_ids,omitempty"`
}

//+kubebuilder:object:root=true

// IntegrationList contains a list of Integration
type IntegrationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Integration `json:"items"`
}
