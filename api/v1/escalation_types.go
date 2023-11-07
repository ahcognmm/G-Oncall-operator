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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// TODO: teamid default is nothing
// EscalationSpec defines the desired state of Escalation
type EscalationSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	ID                string             `json:"id,omitempty"`
	Name              string             `json:"name,omitempty"`
	EscalationPolices []EscalationPolicy `json:"escalation_policies,omitempty"`
}

type EscalationPolicy struct {
	ID        string `json:"id,omitempty"`
	Position  string `json:"position,omitempty"`
	Type      string `json:"type,omitempty"`
	Important string `json:"important,omitempty"`

	// TODO: auto convert any number to valid values

	// +kubebuilder:validation:Enum=60;300;900;1800;3600
	Duration int `json:"duration,omitempty"`
	// Out going webhook name
	ActionToTrigger             string   `json:"action_to_trigger,omitempty"`
	GroupToNotify               string   `json:"group_to_notify,omitempty"`
	PersonsToNotify             []string `json:"persons_to_notify,omitempty"`
	PersonsToNotifyNextEachTime []string `json:"persons_to_notify_next_each_time,omitempty"`
	NotifyOnCallFromSchedule    string   `json:"notify_on_call_from_schedule,omitempty"`
	NotifyIfTimeFrom            string   `json:"notify_if_time_from,omitempty"`
	NotifyIfTimeTo              string   `json:"notify_if_time_to,omitempty"`
}

// EscalationStatus defines the observed state of Escalation
type EscalationStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Escalation is the Schema for the escalations API
type Escalation struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   EscalationSpec   `json:"spec,omitempty"`
	Status EscalationStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// EscalationList contains a list of Escalation
type EscalationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Escalation `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Escalation{}, &EscalationList{})
}
