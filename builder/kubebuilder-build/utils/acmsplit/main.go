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
	"bufio"
	"bytes"
	b64 "encoding/base64"
	"flag"
	"fmt"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"k8s.io/api/admissionregistration/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	utilyaml "k8s.io/apimachinery/pkg/util/yaml"
	"log"
	//metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	//"sigs.k8s.io/controller-runtime/pkg/webhook"
	// "k8s.io/client-go/pkg/api"
	// _ "k8s.io/client-go/pkg/api/install"
	// _ "k8s.io/client-go/pkg/apis/extensions/install"
	// _ "k8s.io/client-go/pkg/apis/rbac/install"
	// //...
	// // "k8s.io/client-go/pkg/api/v1"
	// "k8s.io/client-go/pkg/apis/rbac/v1beta1"
	// make sure to import all client-go/pkg/api(s) you need to cover with your code
)

// createConfigMapWithYAML creates a ConfigMap with the YAML of the webhook, to be able to generate
// at install time with the key information and caBundle
func createConfigMapWithYAML(b []byte) ([] byte, error) {
	decode := scheme.Codecs.UniversalDeserializer().Decode
	obj, groupVersionKind, err := decode(b, nil, nil)
	if err != nil {
		return nil, errors.New("Invalid Object")
	}

	name := ""
	namespace := ""
	if groupVersionKind.Kind == "MutatingWebhookConfiguration" {
		name = obj.(*v1beta1.MutatingWebhookConfiguration).Name
		namespace = obj.(*v1beta1.MutatingWebhookConfiguration).Namespace
	}
	if groupVersionKind.Kind == "ValidatingWebhookConfiguration" {
		name = obj.(*v1beta1.ValidatingWebhookConfiguration).Name
		namespace = obj.(*v1beta1.ValidatingWebhookConfiguration).Namespace
	}

	if len(namespace) > 0 {
		namespace = "namespace: " + namespace + "\n"
	}
	//fmt.Println(string(b))
	sEnc := b64.StdEncoding.EncodeToString(b)

	b = append([]byte("---\napiVersion: v1\nkind: ConfigMap\nmetadata:\n" +
		"  name: " + name + "\n" + namespace +
		"data:\n" +
		"  webhook.yaml: " + sEnc + "\n"))

	return b, nil

}


// SplitYAMLDocuments reads the YAML bytes per-document, unmarshals the TypeMeta information from each document
// and returns a map between the GroupVersionKind of the document and the document bytes
func SplitYAMLDocuments(yamlBytes []byte) ([]byte, []byte, error) {
	gvkListCluster := make([]byte, 0)
	gvkListOther := make([]byte, 0)
	errs := []error{}
	buf := bytes.NewBuffer(yamlBytes)
	reader := utilyaml.NewYAMLReader(bufio.NewReader(buf))
	for {
		typeMetaInfo := runtime.TypeMeta{}
		typeObjectInfo := corev1.ObjectReference{}
		// Read one YAML document at a time, until io.EOF is returned
		b, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, nil, err
		}
		if len(b) == 0 {
			break
		}
		// Deserialize the TypeMeta information of this byte slice
		if err := yaml.Unmarshal(b, &typeMetaInfo); err != nil {
			return nil, nil, err
		}
		if err := yaml.Unmarshal(b, &typeObjectInfo); err != nil {
			return nil, nil, err
		}
		// Require TypeMeta information to be present
		if len(typeMetaInfo.APIVersion) == 0 || len(typeMetaInfo.Kind) == 0 {
			errs = append(errs, errors.New("invalid configuration: kind and apiVersion is mandatory information that needs to be specified in all YAML documents"))
			continue
		}

		fmt.Println("kind", typeMetaInfo.Kind, len(b))
		// Save the mapping between the gvk and the bytes that object consists of
		switch typeMetaInfo.Kind {
		case "Namespace":

		case "ClusterRole":
			fallthrough
		case "ClusterRoleBinding":
			fallthrough
		case "CustomResourceDefinition":
			b = append([]byte("---\n"), b...)
			gvkListCluster = append(gvkListCluster, b...)
		case "MutatingWebhookConfiguration":
			fallthrough
		case "ValidatingWebhookConfiguration":
			configMapBytes, err :=   createConfigMapWithYAML(b)
			if err != nil {
				return nil, nil, err
			}
			gvkListOther = append(gvkListOther, configMapBytes...)
		default:
			b = append([]byte("---\n"), b...)
			gvkListOther = append(gvkListOther, b...)

		}
	}

	return gvkListCluster, gvkListOther, nil
}

//TODO - This could use a shared utitlity
func generateOSSHeader() []byte {
  return []byte(
`# Copyright 2019 Google LLC
 #
 # Licensed under the Apache License, Version 2.0 (the "License");
 # you may not use this file except in compliance with the License.
 # You may obtain a copy of the License at
 #
 #      http://www.apache.org/licenses/LICENSE-2.0
 #
 # Unless required by applicable law or agreed to in writing, software
 # distributed under the License is distributed on an "AS IS" BASIS,
 # WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 # See the License for the specific language governing permissions and
 # limitations under the License.`)
# 
# Copyright 2019 Google LLC. This software is provided as-is, 
# without warranty or representation for any use or purpose.#
#



}

// Utility to split yaml files for Anthos Config Management to separate cluster commands from Other
func main() {
	readFrom := flag.String("read-from", "./config/generated/all.yaml", "a string")
	splitToDir := flag.String("split-to-dir", "./config/generated/", "a string")
	flag.Parse()

	yamlFile, err := ioutil.ReadFile(*readFrom)
	if err != nil {
		log.Println(fmt.Sprintf("Error while decoding YAML object. Err was: %s", err))
	}
	// parseK8sYaml(yamlFile)
	gvkCluster, gvkOther, err := SplitYAMLDocuments(yamlFile)
	if err != nil {
		log.Println(err, "errors-gvkmap")
	}

  gvkCluster = append(generateOSSHeader(), gvkCluster...)
  gvkOther = append(generateOSSHeader(), gvkOther...)


	fmt.Println("Total - Len", len(gvkCluster)+len(gvkOther))
	ioutil.WriteFile(*splitToDir + "all-cluster.yaml", gvkCluster, 0644)
	ioutil.WriteFile(*splitToDir + "all-other.yaml", gvkOther, 0644)
}
