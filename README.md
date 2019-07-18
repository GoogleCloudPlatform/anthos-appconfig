# Anthos Application Configuration - Custom Resource Definition/Controller

This project is about application configuration for deployment.
It focuses on two different user personas:
 
* the developer team 
* the platform team

The Platform team approves an Application Configuration CRD and accompanying
 webhooks (mutation/validation) admission controllers which all together 
 set up the application operation environment definition.   
 
The CRD builds the guardrails and allows integration with other pods and services.
In our proposed environments, the guardrails include “least privileged” 
for namespace both Network ACL and RBAC.  CRD and webhooks are built 
using kubebuilder v2 (v2.0.0-alpha 4) (https://github.com/kubernetes-sigs/kubebuilder)
which leverages the k8s controller framework.


## High Level Diagram

[[https://github.com/GoogleCloudPlatform/anthos-appconfig/wiki/images/global/ApplicationConfigTemplate.png|ApplicatinConfigTemplate High Level View]]


### Source Code Headers

Every file containing source code must include copyright and license
information. This includes any JS/CSS files that you might be serving out to
browsers. (This is to help well-intentioned people avoid accidental copying that
doesn't comply with the license.)

Apache header:

    Copyright 2019 Google LLC

    Licensed under the Apache License, Version 2.0 (the "License");
    you may not use this file except in compliance with the License.
    You may obtain a copy of the License at

        https://www.apache.org/licenses/LICENSE-2.0

    Unless required by applicable law or agreed to in writing, software
    distributed under the License is distributed on an "AS IS" BASIS,
    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
    See the License for the specific language governing permissions and
    limitations under the License.
