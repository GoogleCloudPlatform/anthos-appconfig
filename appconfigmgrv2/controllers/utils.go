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
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/proto"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

// unstructuredFromProto converts a proto message into object spec.
// NOTE:
// We are currently using the istio proto (k8s spec) to build an unstructured object.
// Eventually we should be able to use an official istio k8s client:
// https://github.com/istio/istio/issues/8772
func unstructuredFromProto(
	gvk schema.GroupVersionKind,
	meta map[string]interface{},
	spec proto.Message,
) (*unstructured.Unstructured, error) {
	jsn, err := (&jsonpb.Marshaler{}).MarshalToString(spec)
	if err != nil {
		return nil, err
	}

	var specMap map[string]interface{}
	if err := json.Unmarshal([]byte(jsn), &specMap); err != nil {
		return nil, err
	}

	u := unstructured.Unstructured{
		Object: map[string]interface{}{
			"metadata": meta,
			"spec":     specMap,
		},
	}
	u.SetGroupVersionKind(gvk)

	return &u, nil
}

// unstructuredNames returns a map of NamespacedName's for a list of unstructured
// objects.
func unstructuredNames(list []*unstructured.Unstructured) map[types.NamespacedName]bool {
	names := make(map[types.NamespacedName]bool)
	for _, obj := range list {
		names[types.NamespacedName{Name: obj.GetName(), Namespace: obj.GetNamespace()}] = true
	}
	return names
}

// parseAllowedClient splits an allowedClient into namespace, app-label
// returned in that order.
func parseAllowedClient(allowedClient, defaultNamespace string) (string, string, error) {
	if allowedClient == "" {
		return "", "", errors.New("empty allowed client")
	}

	split := strings.Split(allowedClient, "/")
	switch n := len(split); n {
	case 2:
		return split[0], split[1], nil
	case 1:
		return defaultNamespace, allowedClient, nil
	default:
		return "", "", fmt.Errorf("expected format <namespace>/<app> or <app>, got: %s", allowedClient)
	}
}
