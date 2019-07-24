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
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	appconfigmgrv1alpha1 "github.com/GoogleCloudPlatform/anthos-appconfig/appconfigmgrv2/api/v1alpha1"
)

const (
	VAULT_CONFIGMAP_NAME = "vault"
	VAULT_CA_SECRET_NAME = "vault-ca"
	VAULT_ACM_ROLE       = "acm-vault"
	TODO_FIND_NAMESPACE  = "appconfigmgrv2-system"
)

func (r *AppEnvConfigTemplateV2Reconciler) reconcileVault(
	ctx context.Context,
	meta metav1.ObjectMeta,
	vaultInfo *appconfigmgrv1alpha1.AppEnvConfigTemplateGCPAccessVaultInfo,
) error {
	log.Info("Starting Vault reconcile")
	defer log.Info("Vault reconcile complete")

	cm := &corev1.ConfigMap{}
	name := types.NamespacedName{
		Name:      VAULT_CONFIGMAP_NAME,
		Namespace: TODO_FIND_NAMESPACE,
	}
	if err := r.Client.Get(ctx, name, cm); err != nil {
		return fmt.Errorf("finding %s ConfigMap: %s", VAULT_CONFIGMAP_NAME, err)
	}

	vaultClient, err := newVaultClient(cm.Data["vault-addr"], cm.Data["serviceaccount-jwt"])
	if err != nil {
		return fmt.Errorf("instantiating vault client: %s", err)
	}

	log.Info("Reconciling", "vault", "policy")

	uid := string(meta.UID)

	hcl := fmt.Sprintf("path \"gcp/key/%s\" { capabilities = [\"read\"] }", vaultInfo.RoleSet)
	if err := vaultClient.setPolicy(uid, hcl); err != nil {
		return err
	}

	log.Info("Reconciling", "vault", "auth roles")

	if err := vaultClient.setK8SRole(uid, meta.Name, meta.Namespace, uid); err != nil {
		return err
	}

	return nil
}
