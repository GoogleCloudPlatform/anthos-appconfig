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
	"testing"

	appconfig "github.com/GoogleCloudPlatform/anthos-appconfig/appconfigmgrv2/api/v1alpha1"
)

func TestOPAConstraints(t *testing.T) {
	r, stop := startTestReconciler(t)
	defer stop()
	instance, cleanup := createTestInstance(t, r.Client, true)
	defer cleanup()

	list := &appconfig.AppEnvConfigTemplateV2List{
		Items: []appconfig.AppEnvConfigTemplateV2{
			*instance,
		},
	}

	gvr := opaConstraintGVR()

	c := opaDeploymentLabelConstraint(list)

	_, _ = gvr, c
	/*
		TODO: Test existance of constraint. Requires dynamically generated CRD
		to exist, something that a running Gatekeeper controller does.
		unstructuredShouldExist(t, r.Dynamic, gvr, c)
	*/
}
