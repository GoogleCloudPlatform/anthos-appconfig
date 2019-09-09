#!/usr/bin/env bash

# -e TF_LOG=TRACE \
docker run -it --rm --network host  \
  -e TF_VAR_billing_id="$(gsutil cat gs://anthos-appconfig_build/billing/billing_acount_testing.txt)"  \
  -e TF_VAR_org_id=553683458383 \
  -e TF_VAR_folder_id=3925445538   \
  -v /Users/joseret/.config/gcloud:/root/.config/gcloud  \
  -v $(pwd):/root/tf \
  -v /Users/joseret/.ssh:/root/.ssh  -w /root/tf \
  gcr.io/joseret-bootstrap-sce-trax-1/sce-trax-terraform:v11 $1 $2 $3