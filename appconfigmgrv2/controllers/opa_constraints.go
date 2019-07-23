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
	"context"
	"fmt"

	appconfig "github.com/GoogleCloudPlatform/anthos-appconfig/appconfigmgrv2/api/v1alpha1"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// reconcileOPAContraints reconciles OPA Contraint resources
// which are enforced by Gatekeeper.
func (r *AppEnvConfigTemplateV2Reconciler) reconcileOPAContraints(
	ctx context.Context,
	in *appconfig.AppEnvConfigTemplateV2,
	list *appconfig.AppEnvConfigTemplateV2List,
) error {
	gvr := opaConstraintGVR()

	for _, ct := range []*unstructured.Unstructured{
		opaDeploymentLabelConstraint(list),
	} {
		/*
			TODO: What to do about owner? Constraint is not owned by a single instance.
			if err := controllerutil.SetControllerReference(in, ct, r.Scheme); err != nil {
				return fmt.Errorf("setting controller reference: %v", err)
			}
		*/

		if err := r.upsertUnstructured(ctx, ct, gvr, false); err != nil {
			return fmt.Errorf("reconciling: %v", err)
		}
	}

	return nil
}

func opaDeploymentLabelConstraint(list *appconfig.AppEnvConfigTemplateV2List) *unstructured.Unstructured {
	namespaces := make([]string, 0)
	for _, item := range list.Items {
		namespaces = append(namespaces, item.Namespace)
	}

	u := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"metadata": map[string]interface{}{
				"name": "deployments-must-have-correct-labels",
			},
			"spec": map[string]interface{}{
				"match": map[string]interface{}{
					"kinds": []map[string]interface{}{
						{
							"apiGroups": []string{
								"extensions",
								"apps",
							},
							"kinds": []string{
								"Deployment",
							},
						},
						{
							"apiGroups": []string{""},
							"kinds": []string{
								"Pod",
							},
						},
					},
					"namespaces": namespaces,
				},
				"parameters": map[string]interface{}{
					"labels": []string{
						"app",
						"version",
					},
				},
			},
		},
	}

	u.SetGroupVersionKind(opaConstraintGVK())
	return u
}

func opaConstraintGVK() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   "constraints.gatekeeper.sh",
		Version: "v1alpha1",
		Kind:    "AppConfigRequiredLabels",
	}
}

func opaConstraintGVR() schema.GroupVersionResource {
	gvk := opaConstraintGVK()
	return schema.GroupVersionResource{
		Group:    gvk.Group,
		Version:  gvk.Version,
		Resource: "appconfigrequiredlabels",
	}
}
