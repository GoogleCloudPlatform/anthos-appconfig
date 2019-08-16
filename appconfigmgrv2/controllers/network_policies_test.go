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
	"testing"

	appconfig "github.com/GoogleCloudPlatform/anthos-appconfig/appconfigmgrv2/api/v1alpha1"

	"github.com/stretchr/testify/require"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func TestReconcileNetworkPolicies(t *testing.T) {
	r, stop := startTestReconciler(t)
	defer stop()
	in, cleanup := createTestInstance(t, testFeatureFlags{istio: false})
	defer cleanup()

	var n int
	for _, s := range in.Spec.Services {
		if len(s.AllowedClients) > 0 {
			n++
		}
	}
	nps, err := networkPolicies(in)
	require.NoError(t, err)
	require.Len(t, nps, n)

	for i, np := range nps {
		key := types.NamespacedName{
			Name:      np.Name,
			Namespace: in.Namespace,
		}
		obj := &netv1.NetworkPolicy{}

		ctx := context.Background()
		retryTest(t, func() error { return r.Client.Get(ctx, key, obj) })

		// Test garbage collection by removing service.
		removeServiceFromSpec(t, r.Client, in, i)
		retryTest(t, func() error { return shouldBeNotFound(r.Client.Get(ctx, key, obj)) })
	}
}

func TestNewNetworkPolicies(t *testing.T) {
	template := &appconfig.AppEnvConfigTemplateV2{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-meta-name",
			Namespace: "my-meta-namespace",
		},
		Spec: appconfig.AppEnvConfigTemplateV2Spec{
			Services: []appconfig.AppEnvConfigTemplateServiceInfo{
				{
					Name:          "my-service-name",
					DeploymentApp: "my-deployment-app",
					AllowedClients: []appconfig.AppEnvConfigTemplateRelatedClientInfo{
						{Name: "my-allowed-service-name-0"},
						{Name: "my-allowed-service-name-1"},
					},
				},
			},
		},
	}

	expectedNetworkPolicies := []*netv1.NetworkPolicy{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-meta-name-service--my-service-name",
				Namespace: "my-meta-namespace",
			},
			Spec: netv1.NetworkPolicySpec{
				PodSelector: metav1.LabelSelector{
					MatchLabels: map[string]string{
						"app": "my-deployment-app",
					},
				},
				Ingress: []netv1.NetworkPolicyIngressRule{
					{
						From: []netv1.NetworkPolicyPeer{
							{
								PodSelector: &metav1.LabelSelector{
									MatchLabels: map[string]string{
										"app": "my-allowed-service-name-0",
									},
								},
							},
							{
								PodSelector: &metav1.LabelSelector{
									MatchLabels: map[string]string{
										"app": "my-allowed-service-name-1",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	ps, err := networkPolicies(template)
	require.NoError(t, err)
	require.EqualValues(t, expectedNetworkPolicies, ps)
}
