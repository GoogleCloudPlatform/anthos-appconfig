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

	"github.com/stretchr/testify/require"
)

func TestIstioInstalled(t *testing.T) {
	r, stop := startTestReconciler(t)
	defer stop()
	in, cleanup := createTestInstance(t, testFeatureFlags{istio: true})
	defer cleanup()

	cfg, err := r.getConfig()
	require.NoError(t, err)

	list, err := istioHandlers(cfg, in)
	require.NoError(t, err)
	require.Len(t, list, len(in.Spec.Services))

	gvr := istioHandlerGVR()

	for _, h := range list {
		unstructuredShouldExist(t, r.Dynamic, gvr, h)
	}

	for i := range in.Spec.Services {
		removeServiceFromSpec(t, r.Client, in, i)
	}

	for _, h := range list {
		unstructuredShouldNotExist(t, r.Dynamic, gvr, h)
	}
}
