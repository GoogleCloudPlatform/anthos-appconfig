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
  "context"
  "crypto/tls"
  "crypto/x509"
  "encoding/json"
  "flag"
  "io"
  "io/ioutil"
  "net/http"
  "path/filepath"
  "strings"
  "time"

  "fmt"
  "github.com/hashicorp/vault/api"
  //"encoding/base64"
  "github.com/pkg/errors"
  //"io/ioutil"
  log "github.com/sirupsen/logrus"
  "golang.org/x/net/http2"

  "os"

  corev1 "k8s.io/api/core/v1"
  metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
  "k8s.io/client-go/kubernetes"
  "k8s.io/client-go/tools/clientcmd"

  //"time"
)

func mustGetenv(k string) string {
  v := os.Getenv(k)
  if v == "" {
    panic(fmt.Sprintf("%s undefined", k))
  }
  return v
}

// loadCert loads a single pem-encoded certificate into the given pool.
func loadCert(pool *x509.CertPool, pem []byte) error {
  if ok := pool.AppendCertsFromPEM(pem); !ok {
    return fmt.Errorf("failed to parse PEM")
  }
  return nil
}

func rootCAs(vaultCaPem string) (*x509.CertPool, error) {
  //switch {
  //case vaultCaPem != "":
  pool := x509.NewCertPool()
  if err := loadCert(pool, []byte(vaultCaPem)); err != nil {
    return nil, err
  }
  return pool, nil
  //case vaultCaCert != "":
  //  pool := x509.NewCertPool()
  //  if err := loadCertFile(pool, vaultCaCert); err != nil {
  //    return nil, err
  //  }
  //  return pool, nil
  //case vaultCaPath != "":
  //  pool := x509.NewCertPool()
  //  if err := loadCertFolder(pool, vaultCaPath); err != nil {
  //    return nil, err
  //  }
  //  return pool, nil
  //default:
  //  pool, err := x509.SystemCertPool()
  //  if err != nil {
  //    return nil, errors.Wrap(err, "failed to load system certs")
  //  }
  //  return pool, err
  //}
}

// svcAcctJWT looks up the stored JWT secret token for a given service account
func svcAcctJWT(ctx context.Context, name, namespace string) (string, error) {
  log.Info("common:svcAcctJWT")

  var (
    err error

    secret     = &corev1.Secret{}
    svcAccount = &corev1.ServiceAccount{}
  )

  log.Info("common:svcAcctJWT:secret", "name", name, "namespace", namespace)

  config, err := clientcmd.BuildConfigFromFlags("", "")
  if err != nil {
    panic(err)
  }
  clientset, err := kubernetes.NewForConfig(config)
  if err != nil {
    panic(err)
  }

  // get service account
  sa, err := clientset.CoreV1().ServiceAccounts(namespace).Get(name, metav1.GetOptions{})
  if err != nil {
    log.Error(err, "get ServiceAccount")
    return "", fmt.Errorf("%s serviceAccount not found in %s namespace", name, namespace)
  }

  if len(sa.Secrets) == 0 {
    return "", fmt.Errorf("%s serviceAccount token not found", name)
  }

  log.Info("common:svcAcctJWT:secret:value", "name", name, "namespace", namespace)

  ref := svcAccount.Secrets[0]

  // get service account token secret
  secret, err = clientset.CoreV1().Secrets(namespace).Get(ref.Name, metav1.GetOptions{})
  if err != nil {
    return "", fmt.Errorf("%s serviceAccount token not found: %s", name, err)
  }

  b := string(secret.Data["token"])
  //b, err := base64.StdEncoding.DecodeString(string(secret.Data["token"]))
  //if err != nil {
  //	return "", err
  //}

  return string(b), nil
}

//// getApplicationsSecrets looks up the stored JWT secret token for a given service account
//func getApplicationsSecrets(ctx context.Context, name string, namespace string) (*map[string]string, error) {
//  log.Info("common:getApplicationsSecrets")
//
//  var (
//    appSecretInfo = map[string]string{}
//  )
//
//  log.Info("common:getApplicationsSecrets:secret:", "name-", name, "-namespace-", namespace)
//
//  config, err := rest.InClusterConfig()
//  if err != nil {
//    panic(err.Error())
//  }
//
//  clientset, err := kubernetes.NewForConfig(config)
//  if err != nil {
//    panic(err)
//  }
//
//  // get service account
//  secret, err := clientset.CoreV1().Secrets(namespace).Get(name, metav1.GetOptions{})
//  if err != nil {
//    return &appSecretInfo, fmt.Errorf("%s serviceAccount token not found: %s", name, err)
//  }
//
//  for k, v := range secret.StringData {
//    appSecretInfo[k] = v
//  }
//
//  resAsJSON, _ := json.Marshal(appSecretInfo)
//  log.Println("common:getApplicationsSecrets:JSON:", string(resAsJSON))
//  return &appSecretInfo, nil
//}

func authenticate(role string, jwt string, vaultCaPem string, vaultAddr string, vaultK8SMountPath string) (string, string, error) {
  // Setup the TLS (especially required for custom CAs)


  log.WithFields(log.Fields{
    "role": role,
    "jwt": jwt,
    "vaultAddr": vaultAddr,
    "vaultK8SMountPath": vaultK8SMountPath,
  }).Info("authenticate:start")



  rootCAs, err := rootCAs(vaultCaPem)
  if err != nil {
    return "", "", err
  }

  tlsClientConfig := &tls.Config{
    MinVersion: tls.VersionTLS12,
    RootCAs:    rootCAs,
  }

  //if vaultSkipVerify {
  //  tlsClientConfig.InsecureSkipVerify = true
  //}

  //if vaultServerName != "" {
  //  tlsClientConfig.ServerName = vaultServerName
  //}

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
  log.WithFields(log.Fields{
    "path": credentialPath,
    "dir": filepath.Dir(credentialPath),
  })
  err := os.MkdirAll(filepath.Dir(credentialPath), os.ModePerm)
  if  err != nil {
    return err
  }
  err = ioutil.WriteFile(credentialPath, []byte(key), 0644)
  if  err != nil {
    return err
  }
  return nil
}

const (
  version = "0.1"
)

func watch(dur time.Duration) {
  log.Printf("vault-init-gcp v%s starting watcher", version)

  log.Printf("next cycle in %ds", dur)
  time.Sleep(time.Duration(dur))
  log.Printf("cycling")
}


func main() {
  initMode := flag.String("mode", "GCP-KSA", "a string")
  flag.Parse()

  if *initMode == "GCP-RECYCLE" {
    dur, _ := time.ParseDuration("5m")
    watch(dur)
    os.Exit(0)

  }

  log.WithFields(log.Fields{
    "initMode": *initMode,
  }).Info("main:start")

  var (
    vaultAddr             = mustGetenv("VAULT_ADDR")
    vaultCAPath           = mustGetenv("VAULT_CAPATH")
    gcpRolesetKeyPath     = mustGetenv("INIT_GCP_KEYPATH")
    k8sTokenPath     = mustGetenv("INIT_K8S_TOKEN_KEYPATH")
    k8sPath = mustGetenv("INIT_K8S_KEYPATH")
    k8sRole =mustGetenv("INIT_K8S_ROLE")
    credentialPath        = mustGetenv("GOOGLE_APPLICATION_CREDENTIALS")
    //k8sServiceAccountName = mustGetenv("MY_POD_SERVICE_ACCOUNT")
    k8sNamespace          = mustGetenv("MY_POD_NAMESPACE")
  )

  log.WithFields(log.Fields{
    "vaultAddr":   vaultAddr,
    "vaultCAPath": vaultCAPath,
  }).Info("main:Parms")

  ca, err := ioutil.ReadFile(vaultCAPath)
  if err != nil {
    panic(err)
  }

  log.Infoln("read jwt-ns", k8sNamespace)

  k8sJWT, err := ioutil.ReadFile(k8sTokenPath)
  if err != nil {
   panic(err)
  }

  //log.WithFields(log.Fields{
  //  "ca": string(ca),
  //}).Info("main:ca")
  //
  //VAULT_ADDITIONAL_SECRET := "vault-helper-info"
  //secretsAuthInfo, err := getApplicationsSecrets(context.TODO(),VAULT_ADDITIONAL_SECRET, k8sNamespace )
  //k8sJWT := (*secretsAuthInfo)["ksa.token"]
  //k8sJWT, err := svcAcctJWT(context.TODO(), k8sServiceAccountName, k8sNamespace)
  //if err != nil {
  //  panic(err)
  //}
  //k8sJWT, err := ioutil.ReadFile(k8sTokenPath)
  //if err != nil {
  //  panic(err)
  //}

  log.Infoln("authenticate", string(k8sJWT))

  //err = updateKSAToken(k8sTokenPath, k8sJWT)
  //if err != nil {
  //  panic(err)
  //}

  log.Infoln("authenticate", string(k8sJWT))

  token, accessor, err := authenticate(k8sRole, string(k8sJWT),
    string(ca), vaultAddr, k8sPath)
  if err != nil {
    panic(err)
  }

  log.WithFields(log.Fields{
    "token":    token,
    "accessor": accessor,
  }).Info("authenticate")

  log.Infoln("client")
  client, err := api.NewClient(&api.Config{
    Address: vaultAddr,
  })
  if err != nil {
    panic(err)
  }

  //Auth with K8s vault
  vaultK8sInfo := map[string]interface{}{"jwt": string(k8sJWT), "role": k8sRole}
  secret, err := client.Logical().Write(fmt.Sprintf("auth/%s/login",
    k8sPath), vaultK8sInfo)
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
