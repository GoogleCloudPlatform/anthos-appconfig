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
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIstioInstances(t *testing.T) {
	r, stop := startTestReconciler(t)
	defer stop()
	in, cleanup := createTestInstance(t, r.Client, true)
	defer cleanup()

	gvr := istioInstanceGVR()

	// App-label instance.
	appLabelInst, err := istioAppLabelInstance(in)
	require.NoError(t, err)

	unstructuredShouldExist(t, r.Dynamic, gvr, appLabelInst)

	// Namespace instance.
	nsInst, err := istioNamespaceInstance(in)
	require.NoError(t, err)

	unstructuredShouldExist(t, r.Dynamic, gvr, nsInst)
}
