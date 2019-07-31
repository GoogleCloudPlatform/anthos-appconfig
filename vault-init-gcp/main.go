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
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
)

var (
	version   = "0.1"
	build     = "dev"
	userAgent = fmt.Sprintf("vault-gcp-init/%s (%s)", version, runtime.Version())

	K8S_ROOT     string
	K8S_ROLENAME string

	KSA_JWT                        = mustGetenv("KSA_JWT")
	VAULT_ADDR                     = mustGetenv("VAULT_ADDR")
	VAULT_CAPATH                   = mustGetenv("VAULT_CAPATH")
	INIT_K8S_KEYPATH               = mustGetenv("INIT_K8S_KEYPATH")
	INIT_GCP_KEYPATH               = mustGetenv("INIT_GCP_KEYPATH")
	GOOGLE_APPLICATION_CREDENTIALS = mustGetenv("GOOGLE_APPLICATION_CREDENTIALS")
)

func mustGetenv(k string) string {
	v := os.Getenv(k)
	if v == "" {
		panic(fmt.Sprintf("%s undefined", k))
	}
	return v
}

func parseK8S() {
	re := regexp.MustCompile("(auth/.*)/role/(.*)")

	mg := re.FindStringSubmatch(INIT_K8S_KEYPATH)
	if len(mg) != 3 {
		panic(fmt.Sprintf("invalid value for INIT_K8S_KEYPATH: \"%s\"", INIT_K8S_KEYPATH))
	}

	K8S_ROOT = mg[1]
	K8S_ROLENAME = mg[2]
}

func main() {
	log.Printf("vault-gcp-init v%s-%s starting", version, build)
	parseK8S()

	c := newVaultClient()
	if err := c.login(); err != nil {
		panic(err)
	}

	res, err := c.readGCPKey()
	if err != nil {
		panic(err)
	}

	b, err := base64.StdEncoding.DecodeString(res.Data.PrivateKeyData)
	if err != nil {
		panic(err)
	}

	if err := ioutil.WriteFile(GOOGLE_APPLICATION_CREDENTIALS, b, 0644); err != nil {
		panic(err)
	}

	log.Printf("wrote credentials to %s", GOOGLE_APPLICATION_CREDENTIALS)
}
