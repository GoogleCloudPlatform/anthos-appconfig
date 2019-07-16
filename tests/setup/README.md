# Testing


## Build Environment

```bash
  gcloud builds submit --config=tests/setup/cloudbuild.yaml \
   tests/setup --project anthos-appconfig \
   --substitutions='_BRANCH_NAME=feat_test_end_to_end_137548002'

```