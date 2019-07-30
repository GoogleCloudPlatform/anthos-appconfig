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

// reconcileIstioRules reconciles istio Rule instances.
func (r *AppEnvConfigTemplateV2Reconciler) reconcileIstioRules(
	ctx context.Context,
	cfg Config,
	in *appconfig.AppEnvConfigTemplateV2,
) error {
	list, err := istioRules(cfg, in)
	if err != nil {
		return fmt.Errorf("building: %v", err)
	}

	gvr := istioRuleGVR()

	for _, h := range list {
		if err := controllerutil.SetControllerReference(in, h, r.Scheme); err != nil {
			return err
		}

		if err := r.upsertUnstructured(ctx, h, gvr, true); err != nil {
			return fmt.Errorf("reconciling: %v", err)
		}
	}

	if err := r.garbageCollect(in, unstructuredNames(list), gvr); err != nil {
		return fmt.Errorf("garbage collecting: %v", err)
	}

	return nil
}

func istioRules(cfg Config, t *appconfig.AppEnvConfigTemplateV2) ([]*unstructured.Unstructured, error) {
	list := make([]*unstructured.Unstructured, 0, len(t.Spec.Services))
	gvk := istioRuleGVK()

	for i := range t.Spec.Services {
		var allowedClients types.ListValue
		for _, allowed := range t.Spec.Services[i].AllowedClients {
			allowedClients.Values = append(allowedClients.Values, &types.Value{Kind: &types.Value_StringValue{StringValue: allowed.Name}})
		}

		meta := map[string]interface{}{
			"name":      istioRuleName(t, i),
			"namespace": t.Namespace,
		}
		spec := &v1beta1.Rule{
			Match: fmt.Sprintf(`destination.labels["app"] == "%v"`, t.Spec.Services[i].Name),
			Actions: []*v1beta1.Action{
				{Handler: istioWhitelistHandlerName(t, i), Instances: []string{istioAppLabelInstanceName(t)}},
			},
		}

		unst, err := unstructuredFromProto(gvk, meta, spec)
		if err != nil {
			return nil, fmt.Errorf("unstructured from proto: %v", err)
		}
		list = append(list, unst)
	}

	return list, nil
}

func istioRuleName(t *appconfig.AppEnvConfigTemplateV2, i int) string {
	return fmt.Sprintf("%v-checkdestination--%v",
		t.Name,
		t.Spec.Services[i].Name,
	)
}

func istioRuleGVK() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   "config.istio.io",
		Version: "v1alpha2",
		Kind:    "rule",
	}
}

func istioRuleGVR() schema.GroupVersionResource {
	gvk := istioRuleGVK()
	return schema.GroupVersionResource{
		Group:    gvk.Group,
		Version:  gvk.Version,
		Resource: "rules",
	}
}
