package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CollectorAgentSpec defines desired collector agent state.
type CollectorAgentSpec struct {
	Name          string `json:"name"`
	Environment   string `json:"environment,omitempty"`
	Version       string `json:"version"`
	PolicyID      string `json:"policyId,omitempty"`
	CanaryVersion string `json:"canaryVersion,omitempty"`
}

// CollectorAgentPhase is reconciliation phase.
type CollectorAgentPhase string

const (
	PhasePending  CollectorAgentPhase = "Pending"
	PhaseSynced   CollectorAgentPhase = "Synced"
	PhaseUpgrading CollectorAgentPhase = "Upgrading"
	PhaseFailed   CollectorAgentPhase = "Failed"
)

// CollectorAgentStatus defines observed state.
type CollectorAgentStatus struct {
	Phase               CollectorAgentPhase `json:"phase,omitempty"`
	LastReconcileTime   metav1.Time         `json:"lastReconcileTime,omitempty"`
	ObservedVersion     string              `json:"observedVersion,omitempty"`
	GatewayAgentID      string              `json:"gatewayAgentId,omitempty"`
	Message             string              `json:"message,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,shortName=colagent

// CollectorAgent is the Schema for collectoragents API.
type CollectorAgent struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CollectorAgentSpec   `json:"spec,omitempty"`
	Status CollectorAgentStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CollectorAgentList contains a list of CollectorAgent.
type CollectorAgentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CollectorAgent `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CollectorAgent{}, &CollectorAgentList{})
}
