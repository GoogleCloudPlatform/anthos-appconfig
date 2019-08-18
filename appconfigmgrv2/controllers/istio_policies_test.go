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

	appconfig "github.com/GoogleCloudPlatform/anthos-appconfig/appconfigmgrv2/api/v1alpha1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIstioPolicies(t *testing.T) {
	r, stop := startTestReconciler(t)
	defer stop()
	in, cleanup := createTestInstance(t, testFeatureFlags{istio: true})
	defer cleanup()

	list, err := istioPolicies(in)
	require.NoError(t, err)
	require.Len(t, list, len(in.Spec.Services))

	gvr := istioPolicyGVR()

	for i, p := range list {
		unstructuredShouldExist(t, r.Dynamic, gvr, p)
		removeServiceFromSpec(t, r.Client, in, i)
		unstructuredShouldNotExist(t, r.Dynamic, gvr, p)
	}
}

func TestResolveJWTIssuerJWKS(t *testing.T) {
	cases := []struct {
		name            string
		spec            *appconfig.AppEnvConfigTemplateJWT
		expectedIssuer  string
		expectedJwksUri string
		expectedError   bool
	}{
		{
			name: "ValidGoogle",
			spec: &appconfig.AppEnvConfigTemplateJWT{
				Type: "google",
			},
			expectedIssuer:  "https://accounts.google.com",
			expectedJwksUri: "https://www.googleapis.com/oauth2/v3/certs",
		},
		{
			name: "ValidFirebase",
			spec: &appconfig.AppEnvConfigTemplateJWT{
				Type: "firebase",
				Params: map[string]string{
					"project": "my-firebase-project",
				},
			},
			expectedIssuer:  "https://securetoken.google.com/my-firebase-project",
			expectedJwksUri: "https://www.googleapis.com/service_accounts/v1/jwk/securetoken@system.gserviceaccount.com",
		},
		{
			name: "InvalidFirebaseMissingParams",
			spec: &appconfig.AppEnvConfigTemplateJWT{
				Type: "firebase",
			},
			expectedError: true,
		},
		{
			name: "InvalidFirebaseMissingProjectParam",
			spec: &appconfig.AppEnvConfigTemplateJWT{
				Type:   "firebase",
				Params: map[string]string{},
			},
			expectedError: true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			issuer, jwksUri, err := resolveJWTIssuerJWKS(c.spec)

			assert.Equal(t, c.expectedIssuer, issuer)
			assert.Equal(t, c.expectedJwksUri, jwksUri)
			if c.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
