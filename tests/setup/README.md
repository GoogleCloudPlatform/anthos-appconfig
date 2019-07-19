# Testing


## Build Environment

```bash
 gcloud builds submit --config=tests/setup/cloudbuild.yaml \
    tests/setup --project anthos-appconfig  \
    --substitutions='_BRANCH_NAME=master,_STEPS_X=CLUSTER1_CRD1,_REF_PATH=refs/heads/feat_end_to_end_2_137548002'
```

