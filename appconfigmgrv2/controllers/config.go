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
	istiov1a3 "istio.io/api/networking/v1alpha3"
)

// Config for the controller. This encompasses all knobs that control controller
// behavior outside of app configs.
type Config struct {
	// PolicyCachingInterval determines how long caches should be valid for
	// istio policy decisions.
	PolicyCachingInterval string
	EgressTypes           map[string][]*istiov1a3.Port
}

var defaultConfig = Config{
	// TODO: Update this to be longer for production.
	PolicyCachingInterval: "10s",
	EgressTypes: map[string][]*istiov1a3.Port{
		"https": {
			{
				Name:     "https",
				Number:   443,
				Protocol: "HTTPS",
			},
		},
		"http": {
			{
				Name:     "http",
				Number:   80,
				Protocol: "HTTP",
			},
		},
		"kafka": {
			{
				Name:     "kafka",
				Number:   9092,
				Protocol: "TCP",
			},
			{
				Name:     "kafka-rest",
				Number:   8082,
				Protocol: "HTTP",
			},
			{
				Name:     "kafka-zookeeper",
				Number:   2181,
				Protocol: "TCP",
			},
		},
	},
}
