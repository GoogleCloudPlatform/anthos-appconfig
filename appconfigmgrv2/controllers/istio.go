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

import "k8s.io/apimachinery/pkg/runtime/schema"

var istioTypes = []struct {
	Resource schema.GroupVersionResource
	Kind     schema.GroupVersionKind
}{
	{
		Resource: istioHandlerGVR(),
		Kind:     istioHandlerGVK(),
	},
	{
		Resource: istioInstanceGVR(),
		Kind:     istioInstanceGVK(),
	},
	{
		Resource: istioPolicyGVR(),
		Kind:     istioPolicyGVK(),
	},
	{
		Resource: istioRuleGVR(),
		Kind:     istioRuleGVK(),
	},
	{
		Resource: istioServiceEntryGVR(),
		Kind:     istioServiceEntryGVK(),
	},
	{
		Resource: istioVirtualServiceGVR(),
		Kind:     istioVirtualServiceGVK(),
	},
}
