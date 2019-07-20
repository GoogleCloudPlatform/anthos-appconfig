# Testing


## Build Environment

```bash
 gcloud builds submit --config=tests/setup/cloudbuild.yaml \
    tests/setup --project anthos-appconfig  \
    --substitutions='_BRANCH_NAME=feat_end_to_end_2_137548002,_STEPS_X=CRD1,_REF_PATH=*BRANCH-MANUAL*'
```

