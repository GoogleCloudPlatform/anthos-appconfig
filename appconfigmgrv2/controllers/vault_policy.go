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
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
)

type policyUpdateRequest struct {
	Policy string `json:"policy"`
}

func (vc *vaultClient) setPolicy(name, hcl string) error {
	url := fmt.Sprintf("%s/v1/sys/policies/acl/%s", vc.addr, name)

	b64 := base64.StdEncoding.EncodeToString([]byte(hcl))
	b, err := json.Marshal(policyUpdateRequest{b64})
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(b))
	if err != nil {
		return fmt.Errorf("connection to vault: %s", err)
	}

	resp, err := vc.Do(req)
	if err != nil {
		return fmt.Errorf("applying policy: %s", err)
	}

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("applying policy: %s", resp.Status)
	}

	log.Info("Patched Vault", "policy", name)

	return nil
}
