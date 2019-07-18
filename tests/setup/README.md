# Testing


## Build Environment

```bash
  gcloud builds submit --config=tests/setup/cloudbuild.yaml \
   tests/setup --project anthos-appconfig \
   --substitutions='_BRANCH_NAME=master'

```

