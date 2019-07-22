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
	"testing"

	"github.com/gogo/protobuf/proto"
	"github.com/onsi/gomega"
	"github.com/stretchr/testify/require"
	istionet "istio.io/api/networking/v1alpha3"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestUnstructuredFromProto(t *testing.T) {
	cases := []struct {
		name                 string
		gvk                  schema.GroupVersionKind
		meta                 map[string]interface{}
		spec                 proto.Message
		expectedUnstructured *unstructured.Unstructured
	}{
		{
			name: "Istio VirtualService",
			gvk: schema.GroupVersionKind{
				Group:   "mygroup",
				Version: "v1",
				Kind:    "MyKind",
			},
			meta: map[string]interface{}{"name": "myname"},
			spec: &istionet.VirtualService{
				Hosts:    []string{"host-a", "host-b"},
				Gateways: []string{"gateway-a", "gateway-b"},
				Http: []*istionet.HTTPRoute{{
					Match: []*istionet.HTTPMatchRequest{{
						Uri: &istionet.StringMatch{
							MatchType: &istionet.StringMatch_Exact{
								Exact: "/abc",
							},
						},
					}},
				}},
				//Tls:      []*istionet.TLSRoute{},
				//Tcp:      []*istionet.TCPRoute{},
				ExportTo: []string{"a", "b", "c"},
			},
			expectedUnstructured: &unstructured.Unstructured{

				Object: map[string]interface{}{
					"apiVersion": "mygroup/v1",
					"kind":       "MyKind",
					"metadata": map[string]interface{}{
						"name": "myname",
					},
					"spec": map[string]interface{}{
						"exportTo": []interface{}{"a", "b", "c"},
						"gateways": []interface{}{"gateway-a", "gateway-b"},
						"hosts":    []interface{}{"host-a", "host-b"},
						"http": []interface{}{map[string]interface{}{
							"match": []interface{}{map[string]interface{}{
								"uri": map[string]interface{}{
									"exact": "/abc"}}}}}},
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			g := gomega.NewWithT(t)

			// Debugging type-related issues with gomega.Equal is difficult,
			// use the following to get type info:
			//	u, _ := unstructuredFromProto(c.message)
			//	fmt.Printf("%#v", u)

			g.Expect(unstructuredFromProto(c.gvk, c.meta, c.spec)).
				To(gomega.Equal(c.expectedUnstructured))
		})
	}
}

func TestParseAllowedClient(t *testing.T) {
	const defaultNamespace = "default-namespace"
	cases := []struct {
		name               string
		inputAllowedClient string
		expectedNamespace  string
		expectedAppLabel   string
		expectError        bool
	}{
		{
			name:               "Empty",
			inputAllowedClient: "",
			expectError:        true,
		},
		{
			name:               "NonNamespaced",
			inputAllowedClient: "my-app",
			expectedNamespace:  "default-namespace",
			expectedAppLabel:   "my-app",
		},
		{
			name:               "Namespaced",
			inputAllowedClient: "my-ns/my-app",
			expectedNamespace:  "my-ns",
			expectedAppLabel:   "my-app",
		},
		{
			name:               "TooManySlashes",
			inputAllowedClient: "a/b/c",
			expectError:        true,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ns, app, err := parseAllowedClient(c.inputAllowedClient, defaultNamespace)
			require.Equal(t, c.expectedNamespace, ns)
			require.Equal(t, c.expectedAppLabel, app)
			if c.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
