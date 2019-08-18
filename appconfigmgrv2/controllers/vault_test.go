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

	"github.com/GoogleCloudPlatform/anthos-appconfig/appconfigmgrv2/api/webhooks/builtins"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func TestReconcileVault(t *testing.T) {
	r, stop := startTestReconciler(t)
	defer stop()
	in, cleanup := createTestInstance(t, testFeatureFlags{vault: true})
	defer cleanup()

	s0 := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      builtins.VAULT_CA_SECRET_NAME,
			Namespace: builtins.TODO_FIND_NAMESPACE,
		},
		StringData: map[string]string{
			"key.json": "abc",
		},
	}
	retryTest(t, func() error { return r.Client.Create(context.Background(), s0) })

	// Assert that the secret gets copied into the instance namespace.
	retryTest(t, func() error {
		return r.Client.Get(context.Background(),
			types.NamespacedName{
				Name:      s0.Name,
				Namespace: in.Namespace,
			}, &corev1.Secret{})
	})
}
