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
	// Name of the service.
	Name string `json:"name,omitempty"`
	// Must match the "app" label on the corresponding deployed Pods.
	DeploymentApp string `json:"deploymentApp,omitempty"`
	// Must match the "version" label on the corresponding deployed Pods.
	DeploymentVersion string `json:"deploymentVersion,omitempty"`
	// Must match the port exposed on the corresponding deployed Pods.
	DeploymentPort int32 `json:"deploymentPort,omitempty"`
	// The port for the Kubernetes Service that will be created.
	ServicePort int32 `json:"servicePort,omitempty"`
	// Protocol to use for the service (i.e. "TCP").
	DeploymentPortProtocol corev1.Protocol `json:"deploymentPortProtocol,omitempty"`
	// The set of clients that are allowed to call the service.
	AllowedClients []AppEnvConfigTemplateRelatedClientInfo `json:"allowedClients,omitempty"`
	// Disables the application-wide auth policy (i.e. JWT) for this service.
	DisableAuth bool `json:"disableAuth,omitempty"`
	// Attaches a kubernetes service account to created pods.
	ServiceAccount string `json:"serviceAccount,omitempty"`
	// Specifies the ingress policy for this service (external access).
	Ingress *ServiceIngress `json:"ingress,omitempty"`
}

type ServiceIngress struct {
	Host string `json:"host,omitempty"`
	Path string `json:"path,omitempty"`
}

type AppEnvConfigTemplateRelatedClientInfo struct {
	// Name of the allowed client (corresponds to the "app" label on client Pod). It can be namespaced (i.e. "namespace/app") or it will default to the same namespace as the app config.
	Name string `json:"name,omitempty"`
}

type AppEnvConfigTemplateAuth struct {
	// Configuration for validating JWTs.
	JWT       *AppEnvConfigTemplateJWT       `json:"jwt,omitempty"`
	GCPAccess *AppEnvConfigTemplateGCPAccess `json:"gcpAccess,omitempty"`
}

type AppEnvConfigTemplateJWT struct {
	// Type of system to accept JWTs from (i.e. "firebase").
	Type string `json:"type,omitempty"`
	// Parameters used to identify project/etc. for a given type of system.
	Params map[string]string `json:"params,omitempty"`
}

type AppEnvConfigTemplateGCPAccessSecretInfo struct {
	// The name of the Secret.
	Name string `json:"name,omitempty"`
	// The namespace of the Secret.
	Namespace string `json:"namespace,omitempty"`
}

type AppEnvConfigTemplateGCPAccessVaultInfo struct {
	// Kubernetes service account name used in Vault authentication.
	ServiceAccount string `json:"serviceAccount,omitempty"`
	// Vault Google Cloud Secrets Engine mounted path.
	Path string `json:"path,omitempty"`
	// Vault Google Cloud Secrets Engine roleset name to retrieve credentials from.
	Roleset string `json:"roleset,omitempty"`
}

type AppEnvConfigTemplateGCPAccess struct {
	// Defines the type of GCP access auth granted to the application (must be "secret" or "vault").
	AccessType string `json:"accessType,omitempty"`
	// Used with accessType="secret". Declares the properties of the secret resource.
	SecretInfo *AppEnvConfigTemplateGCPAccessSecretInfo `json:"secretInfo,omitempty"`
	// Used with accessType="vault". Declares the configured Google Cloud roleSet name
	// to be enabled via the given Kubernetes service accounts for use by the application.
	// See https://www.vaultproject.io/docs/secrets/gcp/index.html for details on creating roleSets.
	VaultInfo *AppEnvConfigTemplateGCPAccessVaultInfo `json:"vaultInfo,omitempty"`
}

type AppEnvConfigTemplateAllowedEgress struct {
	// Type of egress traffic (i.e. "http").
	Type string `json:"type,omitempty"`
	// Hosts to allow egress to (i.e. "www.google.com").
	Hosts []string `json:"hosts,omitempty"`
}

type AppEnvConfigTemplateStatusConditionType string

// AppEnvConfigTemplateV2Spec defines the desired state of AppEnvConfigTemplateV2
type AppEnvConfigTemplateV2Spec struct {
	// Services that make up this application (set of services).
	Services []AppEnvConfigTemplateServiceInfo `json:"services,omitempty"`
	// Whitelisted destinations that services may initiate outgoing connections with.
	AllowedEgress []AppEnvConfigTemplateAllowedEgress `json:"allowedEgress,omitempty"`
	// Application-wide authentication configuration.
	Auth *AppEnvConfigTemplateAuth `json:"auth,omitempty"`
}

// AppEnvConfigTemplateV2Status defines the observed state of AppEnvConfigTemplateV2
type AppEnvConfigTemplateV2Status struct{}

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
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=list;get;watch
// +kubebuilder:rbac:groups=core,resources=serviceaccounts,verbs=list;get;watch
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=services/status,verbs=get;update;patch

// +kubebuilder:rbac:groups=networking.k8s.io,resources=networkpolicies,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=networking.k8s.io,resources=networkpolicies/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=extensions,resources=ingresses,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=extensions,resources=ingresses/status,verbs=get;update;patch

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
// +kubebuilder:rbac:groups=constraints.gatekeeper.sh,resources=appconfigrequiredlabels,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=constraints.gatekeeper.sh,resources=appconfigrequiredlabels/status,verbs=get;update;patch

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
