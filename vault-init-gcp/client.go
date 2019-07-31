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

package main

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"runtime"
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

type gcpReadKeyResponse struct {
	Auth interface{} `json:"auth"`
	Data struct {
		KeyAlgorithm   string `json:"key_algorithm"`
		KeyType        string `json:"key_type"`
		PrivateKeyData string `json:"private_key_data"`
	} `json:"data"`
	LeaseDuration int64       `json:"lease_duration"`
	LeaseID       string      `json:"lease_id"`
	Renewable     bool        `json:"renewable"`
	RequestID     string      `json:"request_id"`
	Warnings      interface{} `json:"warnings"`
	WrapInfo      interface{} `json:"wrap_info"`
}

type vaultClient struct {
	http.Client
	addr  string
	token string
}

func newVaultClient() *vaultClient {
	ca, err := ioutil.ReadFile(VAULT_CAPATH)
	if err != nil {
		panic(err)
	}
	caPool := x509.NewCertPool()
	caPool.AppendCertsFromPEM(ca)

	return &vaultClient{
		Client: http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					RootCAs: caPool,
				},
			},
		},
		addr: VAULT_ADDR,
	}
}

func (vc *vaultClient) login() error {
	url := fmt.Sprintf("%s/v1/%s/login", vc.addr, K8S_ROOT)

	b, err := json.Marshal(kubernetesAuthRequest{KSA_JWT, K8S_ROLENAME})
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

	log.Printf("vault authentication successful")
	vc.token = authRes.Auth.ClientToken
	return nil
}

func (vc *vaultClient) doGet(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("User-Agent", userAgent)
	req.Header.Add("X-Vault-Token", vc.token)
	return vc.Do(req)
}

func (vc *vaultClient) readGCPKey() (gcpRes gcpReadKeyResponse, err error) {
	url := fmt.Sprintf("%s/v1/%s", vc.addr, INIT_GCP_KEYPATH)

	resp, err := vc.doGet(url)
	if err != nil {
		return gcpRes, fmt.Errorf("connection to vault: %s", err)
	}
	if resp.StatusCode != http.StatusOK {
		return gcpRes, fmt.Errorf("read key failed: %s", resp.Status)
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&gcpRes)
	if err != nil {
		return gcpRes, fmt.Errorf("read key failed: %s", err)
	}

	log.Printf("gcp key read successful")
	return gcpRes, nil
}
