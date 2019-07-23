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

	istionet "istio.io/api/networking/v1alpha3"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// reconcileIstioServiceEntries reconciles istio VirtualService resources.
// NOTE: These VirtualService entries are not currently providing any functionality.
func (r *AppEnvConfigTemplateV2Reconciler) reconcileIstioVirtualServices(
	ctx context.Context,
	in *appconfig.AppEnvConfigTemplateV2,
) error {
	list, err := istioVirtualServices(in)
	if err != nil {
		return fmt.Errorf("building: %v", err)
	}

	gvr := istioVirtualServiceGVR()

	for _, vs := range list {
		if err := controllerutil.SetControllerReference(in, vs, r.Scheme); err != nil {
			return err
		}

		if err := r.upsertUnstructured(ctx, vs, gvr, true); err != nil {
			return fmt.Errorf("reconciling: %v", err)
		}
	}

	if err := r.garbageCollect(in, unstructuredNames(list), gvr); err != nil {
		return fmt.Errorf("garbage collecting: %v", err)
	}

	return nil
}

func istioVirtualServices(t *appconfig.AppEnvConfigTemplateV2) ([]*unstructured.Unstructured, error) {
	list := make([]*unstructured.Unstructured, 0, len(t.Spec.Services))
	gvk := istioVirtualServiceGVK()

	for i := range t.Spec.Services {
		var (
			meta = map[string]interface{}{
				"name":      istioVirtualServiceName(t, i),
				"namespace": t.Namespace,
			}
			spec = &istionet.VirtualService{
				Hosts: []string{serviceName(t, i)},
				Http: []*istionet.HTTPRoute{{
					Match: []*istionet.HTTPMatchRequest{{
						Uri: &istionet.StringMatch{
							MatchType: &istionet.StringMatch_Prefix{
								Prefix: "/",
							},
						},
					}},
					Route: []*istionet.HTTPRouteDestination{{
						Destination: &istionet.Destination{
							Host: serviceName(t, i),
						},
					}},
				}},
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

func istioVirtualServiceName(t *appconfig.AppEnvConfigTemplateV2, i int) string {
	return fmt.Sprintf("%v-%v", t.Name, t.Spec.Services[i].Name)
}

func istioVirtualServiceGVK() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   "networking.istio.io",
		Version: "v1alpha3",
		Kind:    "VirtualService",
	}
}

func istioVirtualServiceGVR() schema.GroupVersionResource {
	gvk := istioVirtualServiceGVK()
	return schema.GroupVersionResource{
		Group:    gvk.Group,
		Version:  gvk.Version,
		Resource: "virtualservices",
	}
}
