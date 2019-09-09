#!/usr/bin/env bash
gcloud alpha builds triggers create github \
 --repo-name=anthos-appconfig \
 --repo-owner=GoogleCloudPlatform \
 --build-config=builder/appconfig-crd/cloudbuild.yaml \
 --substitutions='_BRANCH_NAME=*AUTO*,_BUILDER_VERSION=v2.0.1,_REF_PATH=BRANCH_AUTO,_APP_VERSION=*AUTO*' \
 --branch-pattern='.*' \
 --project appconfig-crd-env-bld
