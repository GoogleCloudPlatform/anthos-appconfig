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
package controllers

import (
	"context"
	"fmt"

	appconfig "github.com/GoogleCloudPlatform/anthos-appconfig/appconfigmgrv2/api/v1alpha1"

	istiopolicy "istio.io/api/policy/v1beta1"

	"github.com/gogo/protobuf/types"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func (r *AppEnvConfigTemplateV2Reconciler) reconcileIstioHandlers(
	ctx context.Context,
	cfg Config,
	in *appconfig.AppEnvConfigTemplateV2,
) error {
	list, err := istioHandlers(cfg, in)
	if err != nil {
		return fmt.Errorf("building: %v", err)
	}

	gvr := istioHandlerGVR()

	for _, h := range list {
		if err := controllerutil.SetControllerReference(in, h, r.Scheme); err != nil {
			return err
		}

		if err := r.reconcileUnstructured(ctx, h, gvr); err != nil {
			return fmt.Errorf("reconciling: %v", err)
		}
	}

	if err := r.garbageCollect(in, unstructuredNames(list), gvr); err != nil {
		return fmt.Errorf("garbage collecting: %v", err)
	}

	return nil
}

func istioHandlers(cfg Config, t *appconfig.AppEnvConfigTemplateV2) ([]*unstructured.Unstructured, error) {
	list := make([]*unstructured.Unstructured, 0, len(t.Spec.Services))
	for i := range t.Spec.Services {
		appWhitelist, err := istioAppWhitelistHandler(cfg, t, i)
		if err != nil {
			return nil, fmt.Errorf("new app whitelist: %v", err)
		}
		nsWhitelist, err := istioNamespaceWhitelistHandler(cfg, t, i)
		if err != nil {
			return nil, fmt.Errorf("new namespace whitelist: %v", err)
		}

		list = append(list, appWhitelist, nsWhitelist)
	}

	return list, nil
}

func istioAppWhitelistHandler(
	cfg Config,
	t *appconfig.AppEnvConfigTemplateV2,
	i int,
) (*unstructured.Unstructured, error) {
	var allowedClients types.ListValue
	for _, allowed := range t.Spec.Services[i].AllowedClients {
		allowedClients.Values = append(allowedClients.Values, &types.Value{Kind: &types.Value_StringValue{StringValue: allowed.Name}})
	}

	meta := map[string]interface{}{
		"name":      istioAppWhitelistHandlerName(t, i),
		"namespace": t.Namespace,
	}
	spec := &istiopolicy.Handler{
		CompiledAdapter: "listchecker",
		Params: &types.Struct{
			Fields: map[string]*types.Value{
				"overrides":       {Kind: &types.Value_ListValue{ListValue: &allowedClients}},
				"blacklist":       {Kind: &types.Value_BoolValue{BoolValue: false}},
				"cachingInterval": {Kind: &types.Value_StringValue{StringValue: cfg.PolicyCachingInterval}},
			},
		},
	}

	return unstructuredFromProto(istioHandlerGVK(), meta, spec)
}

func istioNamespaceWhitelistHandler(
	cfg Config,
	t *appconfig.AppEnvConfigTemplateV2,
	i int,
) (*unstructured.Unstructured, error) {
	allowedNamespaces := &types.ListValue{
		Values: []*types.Value{&types.Value{Kind: &types.Value_StringValue{StringValue: t.Namespace}}},
	}

	meta := map[string]interface{}{
		"name":      istioNamespaceWhitelistHandlerName(t, i),
		"namespace": t.Namespace,
	}
	spec := &istiopolicy.Handler{
		CompiledAdapter: "listchecker",
		Params: &types.Struct{
			Fields: map[string]*types.Value{
				"overrides":       {Kind: &types.Value_ListValue{ListValue: allowedNamespaces}},
				"blacklist":       {Kind: &types.Value_BoolValue{BoolValue: false}},
				"cachingInterval": {Kind: &types.Value_StringValue{StringValue: cfg.PolicyCachingInterval}},
			},
		},
	}

	return unstructuredFromProto(istioHandlerGVK(), meta, spec)
}

func istioAppWhitelistHandlerName(t *appconfig.AppEnvConfigTemplateV2, i int) string {
	return fmt.Sprintf("%v-app-whitelist--%v",
		t.Name,
		t.Spec.Services[i].Name,
	)
}

func istioNamespaceWhitelistHandlerName(t *appconfig.AppEnvConfigTemplateV2, i int) string {
	return fmt.Sprintf("%v-namespace-whitelist--%v",
		t.Name,
		t.Spec.Services[i].Name,
	)
}

func istioHandlerGVK() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   "config.istio.io",
		Version: "v1alpha2",
		Kind:    "handler",
	}
}

func istioHandlerGVR() schema.GroupVersionResource {
	gvk := istioHandlerGVK()
	return schema.GroupVersionResource{
		Group:    gvk.Group,
		Version:  gvk.Version,
		Resource: "handlers",
	}
}
