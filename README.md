# Anthos Application Configuration - Custom Resource Definition/Controller

## Overview

This project is about application configuration for deployment.
It focuses on two different user personas:
 
* the developer team 
* the platform team

The Platform team approves an Application Configuration CRD and accompanying
 webhooks (mutation/validation) admission controllers which all together 
 set up the application operation environment.   
 
The CRD builds the guardrails and allows integration with other pods and services.
In our proposed environments, the guardrails include “least privileged” 
for namespace both Network ACL and RBAC.  CRD and webhooks are built 
using kubebuilder v2 [v2.0.0-alpha 4] (https://github.com/kubernetes-sigs/kubebuilder)
which leverages the k8s controller framework.

## High Level Diagram

![ApplicatinConfigTemplate High Level View](https://github.com/GoogleCloudPlatform/anthos-appconfig/wiki/images/global/ApplicationConfigTemplate.png)


## Documentation / Information (wiki)

[AppConfig CRD Wiki](https://github.com/GoogleCloudPlatform/anthos-appconfig/wiki)

[Releases](https://github.com/GoogleCloudPlatform/anthos-appconfig/releases)

<code>
Copyright 2019 Google LLC. This software is provided as-is, without warranty or representation for any use or purpose. 
</code>
