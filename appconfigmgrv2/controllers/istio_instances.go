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

	"github.com/gogo/protobuf/types"
	"istio.io/api/policy/v1beta1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// reconcileIstioHandlers reconciles istio Instance resources to support allowedClients
// functionality.
func (r *AppEnvConfigTemplateV2Reconciler) reconcileIstioInstances(
	ctx context.Context,
	in *appconfig.AppEnvConfigTemplateV2,
) error {
	appLabelInst, err := istioAppLabelInstance(in)
	if err != nil {
		return fmt.Errorf("building: %v", err)
	}
	if err := controllerutil.SetControllerReference(in, appLabelInst, r.Scheme); err != nil {
		return err
	}

	gvr := istioInstanceGVR()

	if err := r.upsertUnstructured(ctx, appLabelInst, gvr, true); err != nil {
		return fmt.Errorf("reconciling app label instance: %v", err)
	}

	return nil
}

func istioAppLabelInstance(t *appconfig.AppEnvConfigTemplateV2) (*unstructured.Unstructured, error) {
	var (
		gvk  = istioInstanceGVK()
		meta = map[string]interface{}{
			"name":      istioAppLabelInstanceName(t),
			"namespace": t.Namespace,
		}
		spec = &v1beta1.Instance{
			CompiledTemplate: "listentry",
			Params: &types.Struct{
				Fields: map[string]*types.Value{
					"value": {Kind: &types.Value_StringValue{StringValue: `source.namespace + "/" + source.labels["app"]`}},
				},
			},
		}
	)

	return unstructuredFromProto(gvk, meta, spec)
}

func istioAppLabelInstanceName(t *appconfig.AppEnvConfigTemplateV2) string {
	return fmt.Sprintf("%v-applabel", t.Name)
}

func istioInstanceGVK() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   "config.istio.io",
		Version: "v1alpha2",
		Kind:    "instance",
	}
}

func istioInstanceGVR() schema.GroupVersionResource {
	gvk := istioInstanceGVK()
	return schema.GroupVersionResource{
		Group:    gvk.Group,
		Version:  gvk.Version,
		Resource: "instances",
	}
}
