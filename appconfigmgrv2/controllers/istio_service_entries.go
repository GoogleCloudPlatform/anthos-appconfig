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
	"strings"

	appconfig "github.com/GoogleCloudPlatform/anthos-appconfig/appconfigmgrv2/api/v1alpha1"

	istionet "istio.io/api/networking/v1alpha3"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// reconcileIstioServiceEntries recociles istio ServiceEntry resources to provide
// allowedEgress functionality.
func (r *AppEnvConfigTemplateV2Reconciler) reconcileIstioServiceEntries(
	ctx context.Context,
	cfg Config,
	in *appconfig.AppEnvConfigTemplateV2,
) error {
	list, err := istioServiceEntries(cfg, in)
	if err != nil {
		return fmt.Errorf("building: %v", err)
	}

	gvr := istioServiceEntryGVR()

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

func istioServiceEntries(cfg Config, t *appconfig.AppEnvConfigTemplateV2) ([]*unstructured.Unstructured, error) {
	list := make([]*unstructured.Unstructured, 0, len(t.Spec.AllowedEgress))

	gvk := istioServiceEntryGVK()

	// TODO: Validate no duplicate spec.allowedEgress.type fields.

	for i := range t.Spec.AllowedEgress {
		entry := t.Spec.AllowedEgress[i]

		ports, ok := cfg.EgressTypes[entry.Type]
		if !ok {
			return nil, fmt.Errorf("unknown allowedEgress.type: %v", entry.Type)
		}

		res := istionet.ServiceEntry_DNS
		for _, h := range entry.Hosts {
			if strings.Contains(h, "*") {
				res = istionet.ServiceEntry_NONE
				break
			}
		}

		meta := map[string]interface{}{
			"name":      istioServiceEntryName(t, i),
			"namespace": t.Namespace,
		}
		spec := &istionet.ServiceEntry{
			Hosts:    entry.Hosts,
			Location: istionet.ServiceEntry_MESH_EXTERNAL,
			// TODO: Validation on known types.
			Ports:      ports,
			Resolution: res,
			// Apply to same namespace only:
			ExportTo: []string{"."},
		}

		unst, err := unstructuredFromProto(gvk, meta, spec)
		if err != nil {
			return nil, fmt.Errorf("unstructured from proto: %v", err)
		}
		list = append(list, unst)
	}

	return list, nil
}

func istioServiceEntryName(t *appconfig.AppEnvConfigTemplateV2, i int) string {
	return fmt.Sprintf("%v-%v", t.Name, t.Spec.AllowedEgress[i].Type)
}

func istioServiceEntryGVK() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   "networking.istio.io",
		Version: "v1alpha3",
		Kind:    "ServiceEntry",
	}
}

func istioServiceEntryGVR() schema.GroupVersionResource {
	gvk := istioServiceEntryGVK()
	return schema.GroupVersionResource{
		Group:    gvk.Group,
		Version:  gvk.Version,
		Resource: "serviceentries",
	}
}
