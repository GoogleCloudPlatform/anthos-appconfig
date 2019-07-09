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

/*

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package controllers

import (
	"testing"

	"github.com/gogo/protobuf/proto"
	"github.com/onsi/gomega"
	istiov1a3 "istio.io/api/networking/v1alpha3"
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
			spec: &istiov1a3.VirtualService{
				Hosts:    []string{"host-a", "host-b"},
				Gateways: []string{"gateway-a", "gateway-b"},
				Http: []*istiov1a3.HTTPRoute{{
					Match: []*istiov1a3.HTTPMatchRequest{{
						Uri: &istiov1a3.StringMatch{
							MatchType: &istiov1a3.StringMatch_Exact{
								Exact: "/abc",
							},
						},
					}},
				}},
				//Tls:      []*istiov1a3.TLSRoute{},
				//Tcp:      []*istiov1a3.TCPRoute{},
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
