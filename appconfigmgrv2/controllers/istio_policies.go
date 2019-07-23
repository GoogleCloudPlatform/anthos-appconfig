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

package controllers

import (
	"context"
	"fmt"

	appconfig "github.com/GoogleCloudPlatform/anthos-appconfig/appconfigmgrv2/api/v1alpha1"

	istioauth "istio.io/api/authentication/v1alpha1"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// reconcileIstioHandlers reconciles istio Policy resources to support JWT auth
// functionality.
func (r *AppEnvConfigTemplateV2Reconciler) reconcileIstioPolicies(
	ctx context.Context,
	in *appconfig.AppEnvConfigTemplateV2,
) error {
	list, err := istioPolicies(in)
	if err != nil {
		return fmt.Errorf("building: %v", err)
	}

	gvr := istioPolicyGVR()

	for _, p := range list {
		if err := controllerutil.SetControllerReference(in, p, r.Scheme); err != nil {
			return err
		}

		if err := r.upsertUnstructured(ctx, p, gvr, true); err != nil {
			return fmt.Errorf("reconciling: %v", err)
		}
	}

	if err := r.garbageCollect(in, unstructuredNames(list), gvr); err != nil {
		return fmt.Errorf("garbage collecting: %v", err)
	}

	return nil
}

func istioPolicies(t *appconfig.AppEnvConfigTemplateV2) ([]*unstructured.Unstructured, error) {
	if t.Spec.Auth == nil || t.Spec.Auth.JWT == nil {
		return nil, nil
	}

	list := make([]*unstructured.Unstructured, 0, len(t.Spec.Services))

	issuer, jwksUri, err := resolveJWTIssuerJWKS(t.Spec.Auth.JWT)
	if err != nil {
		return nil, fmt.Errorf("resolving jwt config: %v", err)
	}

	gvk := istioPolicyGVK()

	for i := range t.Spec.Services {
		var triggerRules []*istioauth.Jwt_TriggerRule
		if t.Spec.Services[i].DisableAuth {
			triggerRules = append(triggerRules, &istioauth.Jwt_TriggerRule{
				ExcludedPaths: []*istioauth.StringMatch{
					{
						MatchType: &istioauth.StringMatch_Prefix{
							Prefix: "/",
						},
					},
				},
			})
		}

		var (
			meta = map[string]interface{}{
				"name":      istioPolicyName(t, i),
				"namespace": t.Namespace,
			}
			spec = &istioauth.Policy{
				Targets: []*istioauth.TargetSelector{
					{
						Name: serviceName(t, i),
					},
				},
				Peers: []*istioauth.PeerAuthenticationMethod{
					{
						Params: &istioauth.PeerAuthenticationMethod_Mtls{
							Mtls: &istioauth.MutualTls{},
						},
					},
				},
				Origins: []*istioauth.OriginAuthenticationMethod{
					{
						Jwt: &istioauth.Jwt{
							Issuer:       issuer,
							JwksUri:      jwksUri,
							TriggerRules: triggerRules,
						},
					},
				},
				PrincipalBinding: istioauth.PrincipalBinding_USE_ORIGIN,
			}
		)

		unst, err := unstructuredFromProto(gvk, meta, spec)
		if err != nil {
			return nil, fmt.Errorf("unstructured from proto: %v", err)
		}
		list = append(list, unst)
	}

	return list, nil
}

func resolveJWTIssuerJWKS(spec *appconfig.AppEnvConfigTemplateJWT) (issuer string, jwksUri string, err error) {
	switch typ := spec.Type; typ {
	case "google":
		issuer = "https://accounts.google.com"
		jwksUri = "https://www.googleapis.com/oauth2/v3/certs"
	case "firebase":
		const projectParam = "project"
		errParams := fmt.Errorf("missing required param: %v", projectParam)

		ps := spec.Params
		if ps == nil {
			return "", "", errParams
		}
		proj, ok := ps[projectParam]
		if !ok {
			return "", "", errParams
		}

		issuer = fmt.Sprintf("https://securetoken.google.com/%s", proj)
		jwksUri = "https://www.googleapis.com/service_accounts/v1/jwk/securetoken@system.gserviceaccount.com"
	default:
		return "", "", fmt.Errorf("unrecognized jwt auth type: %v", typ)
	}

	return
}

func istioPolicyName(t *appconfig.AppEnvConfigTemplateV2, i int) string {
	return fmt.Sprintf("%v-%v", t.Name, t.Spec.Services[i].Name)
}

func istioPolicyGVK() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   "authentication.istio.io",
		Version: "v1alpha1",
		Kind:    "Policy",
	}
}

func istioPolicyGVR() schema.GroupVersionResource {
	gvk := istioPolicyGVK()
	return schema.GroupVersionResource{
		Group:    gvk.Group,
		Version:  gvk.Version,
		Resource: "policies",
	}
}
