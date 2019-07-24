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
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type createRoleRequest struct {
	BoundServiceAccountNames      string   `json:"bound_service_account_names"`
	BoundServiceAccountNamespaces string   `json:"bound_service_account_namespaces"`
	MaxTTL                        int64    `json:"max_ttl"`
	Policies                      []string `json:"policies"`
}

func (vc *vaultClient) setK8SRole(roleName, saName, saNamespace string, policies ...string) error {
	url := fmt.Sprintf("%s/v1/auth/kubernetes/role/%s", vc.addr, roleName)

	b, err := json.Marshal(createRoleRequest{
		BoundServiceAccountNames:      saName,
		BoundServiceAccountNamespaces: saNamespace,
		MaxTTL:                        3600,
		Policies:                      policies,
	})
	if err != nil {
		return err
	}

	resp, err := vc.Post(url, "application/json", bytes.NewBuffer(b))
	if err != nil {
		return fmt.Errorf("connection to vault: %s", err)
	}

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("applying role: %s", resp.Status)
	}

	log.Info("Patched Vault", "k8s auth role", roleName)

	return nil
}
