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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func TestReconcileServices(t *testing.T) {
	r, stop := startTestReconciler(t)
	defer stop()
	in, cleanup := createTestInstance(t, testFeatureFlags{})
	defer cleanup()

	// Get a list of expected services.
	svcs := services(in)
	require.Len(t, svcs, len(in.Spec.Services))

	for i, s := range svcs {
		key := types.NamespacedName{
			Name:      s.Name,
			Namespace: in.Namespace,
		}
		obj := &corev1.Service{}

		ctx := context.Background()
		retryTest(t, func() error { return r.Client.Get(ctx, key, obj) })

		// Test garbage collection by removing service from the spec..
		removeServiceFromSpec(t, r.Client, in, i)
		retryTest(t, func() error { return shouldBeNotFound(r.Client.Get(ctx, key, obj)) })
	}
}

// TestNewServices tests the generation of service specs from a given
// app config spec.
func TestNewServices(t *testing.T) {
	cases := []struct {
		name             string
		template         *appconfig.AppEnvConfigTemplateV2
		expectedServices []*corev1.Service
	}{
		{
			name: "basic",
			template: &appconfig.AppEnvConfigTemplateV2{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-meta-name",
					Namespace: "my-meta-namespace",
				},
				Spec: appconfig.AppEnvConfigTemplateV2Spec{
					Services: []appconfig.AppEnvConfigTemplateServiceInfo{
						{
							Name:                   "my-service-name",
							DeploymentApp:          "my-deployment-app",
							DeploymentVersion:      "my-deployment-version",
							DeploymentPort:         1234,
							ServicePort:            5678,
							DeploymentPortProtocol: corev1.ProtocolTCP,
						},
					},
				},
			},
			expectedServices: []*corev1.Service{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "my-meta-name-my-service-name",
						Namespace: "my-meta-namespace",
					},
					Spec: corev1.ServiceSpec{
						Type: corev1.ServiceTypeClusterIP,
						Selector: map[string]string{
							"app": "my-deployment-app",
						},
						Ports: []corev1.ServicePort{
							{
								Name:     "http-default",
								Protocol: corev1.ProtocolTCP,
								Port:     5678,
								TargetPort: intstr.IntOrString{
									IntVal: 1234,
								},
							},
						},
					},
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			svcs := services(c.template)
			require.EqualValues(t, c.expectedServices, svcs)
		})
	}
}
