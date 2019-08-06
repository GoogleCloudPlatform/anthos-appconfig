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
  "flag"
  "io/ioutil"

  "encoding/json"
  //"encoding/base64"
  //"errors"
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
    k8sJWT            = mustGetenv("KSA_JWT")
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

  log.WithFields(log.Fields{
    "ca": string(ca),
  }).Info("main:ca")


  client, err := api.NewClient(&api.Config{
    Address: vaultAddr,
  })
  if err != nil {
    panic(err)
  }

  client.SetToken(k8sJWT)

  data, err := getGCPKey(client, gcpRolesetKeyPath)
  if err != nil {
    panic(err)
  }

  err = updateGCPKey(credentialPath, data)
  if err != nil {
    panic(err)
  }
}
