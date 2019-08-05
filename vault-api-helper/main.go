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
	"fmt"
	"github.com/hashicorp/vault/api"
	//"io/ioutil"
	"log"
	"os"
	"regexp"
	"runtime"
	"encoding/json"
	//"time"
)

const (
	version = "0.1"
	timeFmt = "2006-01-02 15:04:05.999999999 -0700 MST"
)

var (
	userAgent = fmt.Sprintf("vault-init-gcp/%s (%s)", version, runtime.Version())
	//credentialPath = mustGetenv("GOOGLE_APPLICATION_CREDENTIALS")
	//ttlPath        = credentialPath + "_ttl"
)

func mustGetenv(k string) string {
	v := os.Getenv(k)
	if v == "" {
		panic(fmt.Sprintf("%s undefined", k))
	}
	return v
}

func parseK8S() (k8sRoot, k8sRole string) {
	re := regexp.MustCompile("(auth/.*)/role/(.*)")
	keyPath := mustGetenv("INIT_K8S_KEYPATH")

	mg := re.FindStringSubmatch(keyPath)
	if len(mg) != 3 {
		panic(fmt.Sprintf("invalid value for INIT_K8S_KEYPATH: \"%s\"", keyPath))
	}

	return mg[1], mg[2]
}

//func watch() {
//	log.Printf("vault-init-gcp v%s starting watcher", version)
//
//	//log.Printf("reading ttl from %s", ttlPath)
//	b, err := ioutil.ReadFile(ttlPath)
//	if err != nil {
//		panic(err)
//	}
//	expire, err := time.Parse(timeFmt, string(b))
//	if err != nil {
//		panic(err)
//	}
//
//	// set sleep duration to 80% of remaining ttl
//	dur := int64((time.Until(expire).Seconds() * 0.8))
//	log.Printf("next cycle in %ds", dur)
//	time.Sleep(time.Duration(dur) * time.Second)
//	log.Printf("cycling")
//}

func main() {
	if len(os.Args) == 2 && os.Args[1] == "watch" {
		//watch()
		os.Exit(0)
	}

	var (
		//k8sRoot, k8sRole = parseK8S()

		vaultAddr   = mustGetenv("VAULT_ADDR")
		vaultCAPath = mustGetenv("VAULT_CAPATH")
		gcpRolePath = mustGetenv("INIT_GCP_KEYPATH")
		k8sJWT      = mustGetenv("KSA_JWT")
	)

	log.Printf("vault-init-gcp v%s starting", version)

	log.Printf("vault-init-gcp -vaultAddr-%s-vaultCAPath-%s-gcpRolePath-%",
		vaultAddr, vaultCAPath, gcpRolePath)

	client, err := api.NewClient(&api.Config{
		Address: vaultAddr,
	})
	if err != nil {
		panic(err)
	}

	client.SetToken(k8sJWT)

	data, err := client.Logical().Read(gcpRolePath)
	if err != nil {
		panic(err)
	}

	b, _ := json.Marshal(data.Data)
	fmt.Println(string(b))
}
