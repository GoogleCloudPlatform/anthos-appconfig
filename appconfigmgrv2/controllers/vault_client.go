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
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
)

type kubernetesAuthRequest struct {
	Jwt  string `json:"jwt"`
	Role string `json:"role"`
}

type kubernetesAuthResponse struct {
	Auth struct {
		Accessor      string `json:"accessor"`
		ClientToken   string `json:"client_token"`
		EntityID      string `json:"entity_id"`
		LeaseDuration int64  `json:"lease_duration"`
		Metadata      struct {
			Role                     string `json:"role"`
			ServiceAccountName       string `json:"service_account_name"`
			ServiceAccountNamespace  string `json:"service_account_namespace"`
			ServiceAccountSecretName string `json:"service_account_secret_name"`
			ServiceAccountUID        string `json:"service_account_uid"`
		} `json:"metadata"`
		Policies      []string `json:"policies"`
		Renewable     bool     `json:"renewable"`
		TokenPolicies []string `json:"token_policies"`
	} `json:"auth"`
}

type vaultClient struct {
	http.Client
	addr  string
	token string
}

func newVaultClient(addr, jwtToken string) (*vaultClient, error) {
	vc := &vaultClient{
		Client: http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					// TODO: add CA certificate loading and validation
					InsecureSkipVerify: true,
				},
			},
		},
		addr: addr,
	}

	return vc, vc.init(jwtToken)
}

func (vc *vaultClient) init(jwtToken string) error {
	url := fmt.Sprintf("%s/v1/auth/kubernetes/login", vc.addr)

	b, err := json.Marshal(kubernetesAuthRequest{jwtToken, VAULT_ACM_ROLE})
	if err != nil {
		return err
	}

	resp, err := vc.Post(url, "application/json", bytes.NewBuffer(b))
	if err != nil {
		return fmt.Errorf("connection to vault: %s", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("authenticating with vault: %s", resp.Status)
	}
	defer resp.Body.Close()

	var authRes kubernetesAuthResponse
	err = json.NewDecoder(resp.Body).Decode(&authRes)
	if err != nil {
		return fmt.Errorf("authenticating with vault: %s", err)
	}

	vc.token = authRes.Auth.ClientToken
	return nil
}
