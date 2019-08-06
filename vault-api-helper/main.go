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

//  Based on this example https://gist.github.com/jun06t/c5a628abae1cb1562d16f369ca31b22a

package main

import (
  "bytes"
  "crypto/tls"
  "crypto/x509"
  "flag"
  "io"
  "io/ioutil"
  "encoding/json"
  "net/http"
  "strings"

  //"encoding/base64"
  "errors"
  "fmt"
  "github.com/hashicorp/vault/api"
  //"io/ioutil"
  log "github.com/sirupsen/logrus"
  "os"

  //"time"
)

func mustGetenv(k string) string {
  v := os.Getenv(k)
  if v == "" {
    panic(fmt.Sprintf("%s undefined", k))
  }
  return v
}

func rootCAs() (*x509.CertPool, error) {
  switch {
  case vaultCaPem != "":
    pool := x509.NewCertPool()
    if err := loadCert(pool, []byte(vaultCaPem)); err != nil {
      return nil, err
    }
    return pool, nil
  case vaultCaCert != "":
    pool := x509.NewCertPool()
    if err := loadCertFile(pool, vaultCaCert); err != nil {
      return nil, err
    }
    return pool, nil
  case vaultCaPath != "":
    pool := x509.NewCertPool()
    if err := loadCertFolder(pool, vaultCaPath); err != nil {
      return nil, err
    }
    return pool, nil
  default:
    pool, err := x509.SystemCertPool()
    if err != nil {
      return nil, errors.Wrap(err, "failed to load system certs")
    }
    return pool, err
  }
}

func authenticate(role, jwt string) (string, string, error) {
  // Setup the TLS (especially required for custom CAs)
  rootCAs, err := rootCAs()
  if err != nil {
    return "", "", err
  }

  tlsClientConfig := &tls.Config{
    MinVersion: tls.VersionTLS12,
    RootCAs:    rootCAs,
  }

  if vaultSkipVerify {
    tlsClientConfig.InsecureSkipVerify = true
  }

  if vaultServerName != "" {
    tlsClientConfig.ServerName = vaultServerName
  }

  transport := &http.Transport{
    TLSClientConfig: tlsClientConfig,
  }

  if err := http2.ConfigureTransport(transport); err != nil {
    return "", "", errors.New("failed to configure http2")
  }

  client := &http.Client{
    Transport: transport,
  }

  transport.Proxy = http.ProxyFromEnvironment

  addr := vaultAddr + "/v1/auth/" + vaultK8SMountPath + "/login"
  body := fmt.Sprintf(`{"role": "%s", "jwt": "%s"}`, role, jwt)

  req, err := http.NewRequest(http.MethodPost, addr, strings.NewReader(body))
  req.Header.Set("Content-Type", "application/json")
  if err != nil {
    return "", "", errors.Wrap(err, "failed to create request")
  }

  resp, err := client.Do(req)
  if err != nil {
    return "", "", errors.Wrap(err, "failed to login")
  }
  defer resp.Body.Close()

  if resp.StatusCode != 200 {
    var b bytes.Buffer
    io.Copy(&b, resp.Body)
    return "", "", fmt.Errorf("failed to get successful response: %#v, %s",
      resp, b.String())
  }

  var s struct {
    Auth struct {
      ClientToken    string `json:"client_token"`
      ClientAccessor string `json:"accessor"`
    } `json:"auth"`
  }

  if err := json.NewDecoder(resp.Body).Decode(&s); err != nil {
    return "", "", errors.Wrap(err, "failed to read body")
  }

  return s.Auth.ClientToken, s.Auth.ClientAccessor, nil
}

func getGCPKey(c *api.Client, keyRolesetPath string) (string, error) {
  res, err := c.Logical().Read(keyRolesetPath)
  if err != nil {
    return "", err
  }

  resAsJSON, _ := json.Marshal(res.Data)
  return string(resAsJSON), nil
}

func updateGCPKey(credentialPath string, key string) (error) {
  return ioutil.WriteFile(credentialPath, []byte(key), 0644)
}

func main() {
  initMode := flag.String("mode", "GCP-KSA", "a string")
  flag.Parse()

  log.WithFields(log.Fields{
    "initMode": *initMode,
  }).Info("main:start")

  var (
    vaultAddr = mustGetenv("VAULT_ADDR")
    vaultCAPath = mustGetenv("VAULT_CAPATH")
    gcpRolesetKeyPath = mustGetenv("INIT_GCP_KEYPATH")
    k8sJWTPath            = "/var/run/secrets/tokens/vault-token"
    credentialPath    = mustGetenv("GOOGLE_APPLICATION_CREDENTIALS")
  )

  log.WithFields(log.Fields{
    "vaultAddr": vaultAddr,
    "vaultCAPath": vaultCAPath,

  }).Info("main:Parms")

  ca, err := ioutil.ReadFile(vaultCAPath)
  if err != nil {
    panic(err)
  }

  k8sJWT, err := ioutil.ReadFile(k8sJWTPath)
  if err != nil {
    panic(err)
  }

  log.Infoln("read jwt")
  log.WithFields(log.Fields{
    "ca": string(ca),
  }).Info("main:ca")

  client, err := api.NewClient(&api.Config{
    Address: vaultAddr,
  })
  if err != nil {
    panic(err)
  }

  //Auth with K8s vault
  vaultK8sInfo := map[string]interface{}{"jwt": string(k8sJWT), "role": "app-crd-vault-test-role"}
  secret, err := client.Logical().Write(fmt.Sprintf("auth/%s/login",
    "k8s-app-crd-vault-test-appcrd-cicenas-20190806-c-b-bcicen-uc-secrets-vault-13"), vaultK8sInfo)
  if err != nil {
    panic(err)
  }

  client.SetToken(string(secret.Auth.ClientToken))


  log.Infoln("getGCPKey")

  data, err := getGCPKey(client, gcpRolesetKeyPath)
  if err != nil {
    panic(err)
  }

  err = updateGCPKey(credentialPath, data)
  if err != nil {
    panic(err)
  }
}
