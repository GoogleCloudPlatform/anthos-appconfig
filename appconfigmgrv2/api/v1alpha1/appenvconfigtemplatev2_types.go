// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// Copyright 2019 Google LLC. This software is provided as-is,
// without warranty or representation for any use or purpose.
//

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// AppEnvConfigTemplateServiceInfo defines the service info of AppEnvConfigTemplate
type AppEnvConfigTemplateServiceInfo struct {
	Name                   string                                  `json:"name,omitempty"`
	DeploymentApp          string                                  `json:"deploymentApp,omitempty"`
	DeploymentVersion      string                                  `json:"deploymentVersion,omitempty"`
	DeploymentPort         int32                                   `json:"deploymentPort,omitempty"`
	ServicePort            int32                                   `json:"servicePort,omitempty"`
	DeploymentPortProtocol corev1.Protocol                         `json:"deploymentPortProtocol,omitempty"`
	AllowedClients         []AppEnvConfigTemplateRelatedClientInfo `json:"allowedClients,omitempty"`

	DisableAuth bool `json:"disableAuth,omitempty"`
}

type AppEnvConfigTemplateRelatedClientInfo struct {
	Name string `json:"name,omitempty"`
}

type AppEnvConfigTemplateAuth struct {
	JWT       *AppEnvConfigTemplateJWT       `json:"jwt,omitempty"`
	GCPAccess *AppEnvConfigTemplateGCPAccess `json:"gcpAccess,omitempty"`
}

type AppEnvConfigTemplateJWT struct {
	Type   string            `json:"type,omitempty"`
	Params map[string]string `json:"params,omitempty"`
}

type AppEnvConfigTemplateGCPAccessSecretInfo struct {
	// name is the secret resource name"
	Name string `json:"name,omitempty"`
	// namespace where given secret resource name exists
	Namespace string `json:"namespace,omitempty"`
}

type AppEnvConfigTemplateGCPAccessVaultInfo struct {
	// Kubernetes service account to allow to enable dyamic credential generation for
	ServiceAccount string `json:"serviceAccount,omitempty"`
	// Vault Google Cloud Secrets Engine roleset name to enable
	RoleSet string `json:"roleSet,omitempty"`
	// Cluster-specific Vault path where Kubernetes Auth Method is enabled
	RolePath string `json:"rolePath,omitempty"`
}

type AppEnvConfigTemplateGCPAccess struct {
	// accessType defines the type of gcpAccess auth granted to the application.
	// must be one of [secret,vault]
	AccessType string `json:"accessType,omitempty"`
	// when accessType is 'secret', secretInfo declares the properties of the secret resource.
	SecretInfo *AppEnvConfigTemplateGCPAccessSecretInfo `json:"secretInfo,omitempty"`
	// when accessType is 'vault', vaultInfo declares the configured Google Cloud roleSet name
	// to be enabled via the given Kubernetes service accounts for use by the application.
	// see https://www.vaultproject.io/docs/secrets/gcp/index.html for details on creating roleSets
	VaultInfo *AppEnvConfigTemplateGCPAccessVaultInfo `json:"vaultInfo,omitempty"`
}

type AppEnvConfigTemplateAllowedEgress struct {
	Type  string   `json:"type,omitempty"`
	Hosts []string `json:"hosts,omitempty"`
}

type AppEnvConfigTemplateStatusConditionType string

// AppEnvConfigTemplateV2Spec defines the desired state of AppEnvConfigTemplateV2
type AppEnvConfigTemplateV2Spec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Services      []AppEnvConfigTemplateServiceInfo   `json:"services,omitempty"`
	AllowedEgress []AppEnvConfigTemplateAllowedEgress `json:"allowedEgress,omitempty"`
	Auth          *AppEnvConfigTemplateAuth           `json:"auth,omitempty"`
}

// AppEnvConfigTemplateV2Status defines the observed state of AppEnvConfigTemplateV2
type AppEnvConfigTemplateV2Status struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// AppEnvConfigTemplateV2 is the Schema for the appenvconfigtemplatev2s API
// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type AppEnvConfigTemplateV2 struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AppEnvConfigTemplateV2Spec   `json:"spec,omitempty"`
	Status AppEnvConfigTemplateV2Status `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:rbac:groups=appconfigmgr.cft.dev,resources=appenvconfigtemplatev2s,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=appconfigmgr.cft.dev,resources=appenvconfigtemplatev2s/status,verbs=get;update;patch

// +kubebuilder:rbac:groups=core,resources=namespaces,verbs=get;list;watch
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=list;get;watch
// +kubebuilder:rbac:groups=core,resources=secrets/status,verbs=list;get;watch
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=services/status,verbs=get;update;patch

// +kubebuilder:rbac:groups=networking.k8s.io,resources=networkpolicies,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=networking.k8s.io,resources=networkpolicies/status,verbs=get;update;patch

// +kubebuilder:rbac:groups=apiextensions.k8s.io,resources=customresourcedefinitions,verbs=get;list
// +kubebuilder:rbac:groups=apiextensions.k8s.io,resources=customresourcedefinitions/status,verbs=get;list

// +kubebuilder:rbac:groups=authentication.istio.io,resources=policies,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=authentication.istio.io,resources=policies/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=config.istio.io,resources=handlers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=config.istio.io,resources=handlers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=config.istio.io,resources=instances,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=config.istio.io,resources=instances/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=config.istio.io,resources=rules,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=config.istio.io,resources=rules/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=networking.istio.io,resources=virtualservices,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=networking.istio.io,resources=virtualservices/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=networking.istio.io,resources=serviceentries,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=networking.istio.io,resources=serviceentries/status,verbs=get;update;patch

// AppEnvConfigTemplateV2List contains a list of AppEnvConfigTemplateV2
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true
type AppEnvConfigTemplateV2List struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AppEnvConfigTemplateV2 `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AppEnvConfigTemplateV2{}, &AppEnvConfigTemplateV2List{})
}
