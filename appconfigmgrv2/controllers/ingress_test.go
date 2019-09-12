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

	"github.com/stretchr/testify/require"
	"k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/types"
)

func TestReconcileIngress(t *testing.T) {
	r, stop := startTestReconciler(t)
	defer stop()
	in, cleanup := createTestInstance(t, testFeatureFlags{})
	defer cleanup()

	ing := ingress(in)
	require.NotNil(t, ing)

	key := types.NamespacedName{
		Name:      ing.Name,
		Namespace: in.Namespace,
	}
	obj := &v1beta1.Ingress{}

	ctx := context.Background()
	retryTest(t, func() error { return r.Client.Get(ctx, key, obj) })

	// Clear the ingress spec and expect the ingress to be garbage collected.
	noIng := in.DeepCopy()
	for i := range noIng.Spec.Services {
		noIng.Spec.Services[i].Ingress = nil
	}
	require.NoError(t, r.Client.Update(ctx, noIng))

	retryTest(t, func() error { return shouldBeNotFound(r.Client.Get(ctx, key, obj)) })
}
