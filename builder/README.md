

```bash
gsutil mb -p anthos-appconfig -c multi_regional -l us gs://anthos-appconfig/
gsutil mb -p anthos-appconfig -c multi_regional -l us -b on gs://anthos-appconfig_build/
gsutil bucketpolicyonly set off gs://anthos-appconfig_build/
gsutil defacl set private gs://anthos-appconfig_build/

gsutil mb -p anthos-appconfig -c multi_regional -l us -b on gs://anthos-appconfig_public/
gsutil acl ch -u AllUsers:R gs://anthos-appconfig_public/

```


```bash
PROJECT_ID_NUMBER=20604585440
gcloud iam 
ssh-keygen -t rsa -N '' -b 4096 -C "20604585440@cloudbuild.gserviceaccount.com" \
    -f $HOME/.ssh/id_rsa_anthos-appconfig-repo

gsutil cp $HOME/.ssh/id_rsa_anthos-appconfig-repo*  gs://anthos-appconfig_build/repo/keys/
gsutil acl ch -u 20604585440@cloudbuild.gserviceaccount.com:R  gs://anthos-appconfig_build/repo/keys/*
```

```bash
gcloud builds submit \
  --config=./builder/kubebuilder-build/cloudbuild.yaml  \
    ./builder/kubebuilder-build \
  --project anthos-appconfig --substitutions="_BUILDER_VERSION=v2.0.0" 
  
  gsutil iam ch allUsers:objectViewer gs://artifacts.anthos-appconfig.appspot.com
```

```bash
gcloud builds submit \
  --config=./builder/utils/acmsplit/build/cloudbuild.yaml  \
    ./builder/utils/acmsplit \
  --project anthos-appconfig --substitutions="_BUILDER_VERSION=v2.0.0" 
```

```bash
gcloud builds submit \
  --config=./builder/appconfig-crd/cloudbuild.yaml  \
    ./builder/appconfig-crd \
  --project anthos-appconfig \
  --substitutions="_BRANCH_NAME=build-script-2019-07-10,_BUILDER_VERSION=v2.0.1,_APP_VERSION=v2.0.0" 
```