#!/usr/bin/env bash
# Copyright 2019 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -e

####
# GLOBAL CONFIG VARS
####

# TODO - Add Branch Name
TEMPLATE_BUCKET="gs://anthos-appconfig_public/acm/anthos-config-management/build-script-2019-07-10/acm-crd/config-management-root"
EXAMPLES_BUCKET="gs://anthos-appconfig_public/acm/anthos-config-management/build-script-2019-07-10/acm-crd-examples/config-management-root/namespaces"
CM_OPERATOR_BUCKET="gs://config-management-release/released/latest/config-management-operator.yaml"
HELM_IMAGE="alpine/helm:2.13.1"
CM_CRD_COUNT=8

####

# common helper funcs
_red() { echo -ne "\033[31m$@\033[0m"; }
_grn() { echo -ne "\033[32m$@\033[0m"; }
_ylw() { echo -ne "\033[33m$@\033[0m"; }
_output() { echo -e "\n[$(_grn acm-init)] $@"; }
_errexit() { echo -e "[$(_red error)] $@"; exit 1; }
_installed() { command -v "$1" >/dev/null 2>&1; }
_ensure_path() { mkdir -p $(dirname $1); echo "creating $1"; }
_confirm() {
  local x
  prompt=$(echo -ne "$@ \033[32m(y/N)\033[0m")
  read -n1 -p "$prompt" x; echo
  [ "$x" == "y" ] && return 0
  return 1
}

load-gcpvars() {
  export GCP_ACCOUNT=$(gcloud config get-value core/account 2> /dev/null)
  export GCP_PROJECT=$(gcloud config get-value core/project 2> /dev/null)
  [[ -z "$GCP_PROJECT" ]] || [[ -z "$GCP_ACCOUNT" ]] && \
    _errexit "missing gcloud configuration, run 'gcloud init' to create"
  return 0
}

load-ctxvars() {
  # set cluster variables
  export K8S_CONTEXT=$(kubectl config current-context)
  export ACM_CLUSTER_REGISTRY_NAME=${K8S_CONTEXT//_/-}
  export ACM_ENV_ROOT=./env/${ACM_CLUSTER_REGISTRY_NAME}
}

load-repovars() {
  REPO_PATH=${ARGS[0]:-$(pwd)}
  cd $REPO_PATH

  # set repo variables
  export REPO_REMOTE=$(git remote | head -1)
  [[ -z "$REPO_REMOTE" ]] && _errexit "repo missing remote upstream url"
  export REPO_URL=$(git config --get remote.${REPO_REMOTE}.url 2> /dev/null) || _errexit "repo missing remote upstream url"
  # default to master branch on repos with no commit index
  REPO_BRANCH=$(git rev-parse --abbrev-ref HEAD 2> /dev/null) || REPO_BRANCH="master"
  export REPO_BRANCH
}

git-key-url() {
  case $1 in
    *source.*google.com*)
      echo -n "https://source.cloud.google.com/user/ssh_keys" ;;
    *@github.com:*)
      echo -n $1 | sed -E 's#^.*:(.*)/(.*)$#https://github.com/\1/\2/settings/keys#;s#\.git/#/#g' ;;
    *@bitbucket.org:*)
      echo -n $1 | sed -E 's#^.*:(.*)/(.*)$#https://bitbucket.org/\1/\2/admin/access-keys/#;s#\.git/#/#g' ;;
    *@gitlab.com:*)
      echo -n $1 | sed -E 's#^.*:(.*)/(.*)$#https://gitlab.com/\1/\2/-/settings/repository#;s#\.git/#/#g' ;;
    *)
      echo -n "https://cloud.google.com/anthos-config-management/docs/how-to/installing#git-creds-secret" ;;
  esac
}

_echo_vars() {
  echo; for v in $@; do
    echo -e "\033[32m${v}\033[0m\t| ${!v}"
  done | column -t
}

create-repo() {
  load-gcpvars

  export REPO_PATH=$(readlink -f ${ARGS[0]}) || _errexit "invalid path"
  export REPO_NAME=$(basename $REPO_PATH)
  export REPO_URL="ssh://${GCP_ACCOUNT}@google.com@source.developers.google.com:2022/p/${GCP_PROJECT}/r/${REPO_NAME}"
  export REPO_REMOTE=origin
  export REPO_BRANCH=master

  [[ -e "$REPO_PATH" ]] && _errexit "path exists: $REPO_PATH"

  _echo_vars GCP_PROJECT REPO_PATH
  _confirm "\ncreate repo with above configuration?" || exit 0

  _output "creating cloud source repo"
  gcloud source repos create --project $GCP_PROJECT $REPO_NAME

  _output "cloning new repo"
  gcloud source repos clone --project $GCP_PROJECT $REPO_NAME $REPO_PATH
}

init-repo() {
  local create
  for opt in ${OPTS[@]}; do
    case $opt in
      -c) create=1 ;;
      *) _errexit "unknown install option \"$opt\"";;
    esac
  done

  load-ctxvars
  [[ -z "$create" ]] || create-repo
  load-repovars

  echo; for v in REPO_REMOTE REPO_BRANCH REPO_URL K8S_CONTEXT ACM_CLUSTER_REGISTRY_NAME; do
    echo -e "\033[32m${v}\033[0m\t| ${!v}"
  done | column -t
  _confirm "\ninitialize repo with above configuration?" || exit 0

  [[ -a $ACM_ENV_ROOT ]] && _errexit "config root exists: $ACM_ENV_ROOT"

  _output "initializing cluster config root"

  mkdir -p ${ACM_ENV_ROOT}
  gsutil -m cp -R "${TEMPLATE_BUCKET}/*" ${ACM_ENV_ROOT}/

  fpath="${ACM_ENV_ROOT}/clusterregistry/cluster-registry-cluster-info.yaml"
  _ensure_path $fpath
  cat > $fpath <<EOF
kind: Cluster
apiVersion: clusterregistry.k8s.io/v1alpha1
metadata:
  name: ${ACM_CLUSTER_REGISTRY_NAME}
EOF

  fpath="${ACM_ENV_ROOT}/namespaces/istio-system/namespace.yaml";
  _ensure_path $fpath
  cat > $fpath <<EOF
apiVersion: v1
kind: Namespace
metadata:
  name: istio-system
EOF

  fpath="${ACM_ENV_ROOT}/system/README.md"
  _ensure_path $fpath
  cat > $fpath <<EOF
# System

This directory contains system configs such as the repo version and how resources are synced.
EOF

  fpath="${ACM_ENV_ROOT}/system/repo.yaml"
  _ensure_path $fpath
  cat > $fpath <<EOF
apiVersion: configmanagement.gke.io/v1
kind: Repo
metadata:
  creationTimestamp: null
  name: repo
spec:
  version: 0.1.0
status:
  import:
    lastUpdate: null
  source: {}
  sync:
    lastUpdate: null
EOF

  fpath="${ACM_ENV_ROOT}/config-management.yaml"
  _ensure_path $fpath
  cat > $fpath <<EOF
# config-management.yaml

apiVersion: addons.sigs.k8s.io/v1alpha1
kind: ConfigManagement
metadata:
  name: config-management
  namespace: config-management-system
spec:
  # clusterName is required and must be unique among all managed clusters
  clusterName: $ACM_CLUSTER_REGISTRY_NAME
  git:
    syncRepo: $REPO_URL
    syncBranch: $REPO_BRANCH
    secretType: ssh
    policyDir: "${ACM_ENV_ROOT}"
EOF

  # check in new config root
  [[ -f .gitignore ]] || echo 'keys/*' > .gitignore
  git add ${ACM_ENV_ROOT} .gitignore

  mkdir -p ./keys
  key_path="./keys/${ACM_CLUSTER_REGISTRY_NAME}_rsa"
  _output "generating repo ssh key [${key_path}]"
  ssh-keygen -t rsa -b 4096 -N '' -q \
    -C "${ACM_CLUSTER_REGISTRY_NAME}" \
    -f $key_path

  if ! _confirm "\npush new config?"; then
    cat <<EOM

# Complete Setup

1. Run the following commands to complete setup:

git commit -am "auto-initialize $ACM_CLUSTER_REGISTRY_NAME" && git push

kubectl create secret generic git-creds \\
--namespace=config-management-system \\
--from-file=ssh=${key_path}

kubectl apply -f ${ACM_ENV_ROOT}/config-management.yaml


2. Add the below public SSH key to your repo provider
$(git-key-url $REPO_URL)

$(cat ${key_path}.pub)
EOM

    exit 0
  fi

  _output "pushing config repo updates\n"
  git commit -q -am "auto-initialize $ACM_CLUSTER_REGISTRY_NAME"
  git push -q --set-upstream ${REPO_REMOTE} ${REPO_BRANCH}

  _output "configuring operator repo key"
  kubectl create secret generic git-creds \
    --namespace=config-management-system \
    --from-file=ssh=${key_path}
  kubectl apply -f ${ACM_ENV_ROOT}/config-management.yaml

  cat <<EOM

# Complete Setup

Add the below public SSH key to your repo provider to complete setup:
$(git-key-url $REPO_URL)

$(cat ${key_path}.pub)
EOM
}

init-demos() {
  load-ctxvars
  load-gcpvars
  load-repovars
  local x app_iters="1 2" app_name="appconfigcrd-demo"

  echo; for v in REPO_REMOTE REPO_BRANCH REPO_URL GCP_PROJECT K8S_CONTEXT ACM_CLUSTER_REGISTRY_NAME; do
    echo -e "\033[32m${v}\033[0m\t| ${!v}"
  done | column -t
  _confirm "\nproceed with above configuration?" || exit 0

  prompt=$(echo -ne "please provide an app name prefix to use (\033[32m$app_name\033[0m) > ")
  read -p "$prompt" x
  [[ -z "$x" ]] || app_name=$x

  _output "creating pubsub topics and subscriptions"
  for i in $app_iters; do
    topic="${app_name}-topic${i}"
    if (gcloud pubsub topics describe $topic &> /dev/null); then
      echo "$topic topic exists, skipping"
    else
      gcloud pubsub topics create $topic --project $GCP_PROJECT
    fi

    if (gcloud pubsub subscriptions describe $topic &> /dev/null); then
      echo "$topic subscription exists, skipping"
    else
      gcloud pubsub subscriptions create $topic --project $GCP_PROJECT \
        --topic ${app_name}-topic1 --topic-project $GCP_PROJECT
    fi
  done

  _output "adding IAM service account keys to RBAC-protected config management namespace"
  mkdir -p ./keys
  for i in $app_iters; do
    iam_name=${app_name}-sa${i}
    iam_account="${iam_name}@${GCP_PROJECT}.iam.gserviceaccount.com"

    if (gcloud iam service-accounts describe $iam_account &> /dev/null); then
      echo "$iam_account service account exists, skipping creation"
    else
      gcloud iam service-accounts create ${iam_name} --display-name=${iam_name} --project $GCP_PROJECT
    fi

    if (kubectl get secret -n appconfigmgrv2-system ${iam_name}-secret &> /dev/null); then
      echo "${iam_name}-secret exists, skipping"
    else
      gcloud iam service-accounts keys create ./keys/${iam_name}.json --project $GCP_PROJECT \
        --iam-account=${iam_account}
      kubectl create secret generic ${iam_name}-secret \
        -n appconfigmgrv2-system \
        --from-file=key.json=./keys/${iam_name}.json
    fi
  done

  _output "creating service account pubsub ACLs"
  for i in $app_iters; do
    topic="${app_name}-topic${i}"
    iam_name=${app_name}-sa${i}
    iam_account="${iam_name}@${GCP_PROJECT}.iam.gserviceaccount.com"
    gcloud beta pubsub topics add-iam-policy-binding ${topic} --project $GCP_PROJECT \
      --member=serviceAccount:${iam_account} \
      --role=roles/pubsub.publisher || true
    done

  [[ -d "${ACM_ENV_ROOT}/namespaces/use-cases" ]] && {
    _output "${ACM_ENV_ROOT}/namespaces/use-cases already exists, skipping repo update"
    exit 0
  }

  _output "adding demo apps to policy config repo"
  gsutil -m cp -R "${EXAMPLES_BUCKET}/*" ${ACM_ENV_ROOT}/namespaces/
  git add ${ACM_ENV_ROOT}/namespaces/use-cases
  git commit -am "initialize $ACM_CLUSTER_REGISTRY_NAME demo apps" && git push
}

status() {
  _cmo_status
  echo
  _sync_status | column -t -s,
  echo
  #_sync_errors
}

_cmo_status() {
  local n
  echo -ne '\nconfig-management-crds: '
  n=$(kubectl get crds | grep -c configmanagement.gke.io 2> /dev/null) || n=0
  if [[ "$n" -eq 0 ]]; then
    _red "MISSING\n"
  elif [[ "$n" -eq $CM_CRD_COUNT ]]; then
    _grn "OK\n"
  else
    _ylw "PENDING\n"
  fi

  echo -n 'config-management-operator: '
  local n=$(kubectl get deployment config-management-operator \
    --namespace=kube-system \
    -o='go-template' \
    --template='{{.status.readyReplicas}}' 2> /dev/null) || n=""
  if [[ -z "$n" ]]; then
    _red "MISSING\n"
  elif [[ "$n" -ge 1 ]]; then
    _grn "OK\n"
  else
    _ylw "PENDING\n"
  fi
}

_sync_status() {
  echo -e 'COMPONENT,LAST_UPDATE,TOKEN'
  kubectl get repos.configmanagement.gke.io repo \
    -o='go-template' \
    --template='go-template' --template='source,-,{{ .status.source.token }}
git_importer,{{ .status.import.lastUpdate }},{{ .status.import.token }}
git_syncer,{{ .status.sync.lastUpdate }},{{ .status.sync.latestToken }}'
}

_sync_errors() {
  echo "sync errors:"
  kubectl get repos.configmanagement.gke.io repo \
    -o='go-template' \
    --template='{{range $k,$v := .status.source.errors}} {{$k}} | {{$v.errorMessage}}
{{end}}'
}

install() {
  load-ctxvars

  local force
  for opt in ${OPTS[@]}; do
    case $opt in
      -f) force=1; _output "force install enabled" ;;
      *) _errexit "unknown install option \"$opt\"";;
    esac
  done

  echo; for v in K8S_CONTEXT ACM_CLUSTER_REGISTRY_NAME; do
    echo -e "\033[32m${v}\033[0m\t| ${!v}"
  done | column -t
  _confirm "\nproceed with above configuration?" || exit 0

  # operator install
  local n=1
  if [[ -n "$force" ]]; then
    n=0
  else
    (kubectl get crds | grep -q configmanagements.addons.sigs.k8s.io) || n=0
    (kubectl get deployments config-management-operator --namespace=kube-system &> /dev/null) || n=0
  fi

  [[ "$n" -eq 1 ]] || install_operator && _output "config management operator install OK"

  # istio install
  n=1
  if [[ -n "$force" ]]; then
    n=0
  else
    (kubectl get namespaces istio-system &> /dev/null) || n=0
  fi

  [[ "$n" -eq 1 ]] || install_istio && _output "istio install OK"

  _output "done"
}

install_operator() {
  _output "installing config management operator to cluster"
  gsutil cp ${CM_OPERATOR_BUCKET} - | kubectl apply -f -
}

install_istio() {
  local x tmpdir shacmd
  _output "installing istio to cluster"

  istio_version=$(curl -L -s https://api.github.com/repos/istio/istio/releases | \
    grep -E 'tag_name.*1\.1\.' | grep -vE '(\-rc|\-snapshot)' | \
    sed "s/ *\"tag_name\": *\"\\(.*\\)\",*/\\1/" | head -1)
  prompt=$(echo -ne "istio version to install? \033[32m($istio_version)\033[0m ")
  read -p "$prompt" x
  [[ -z "$x" ]] || istio_version=$x

  tmpdir=$(mktemp -d $(pwd)/.acm-init-XXXXXX)
  istio_dir="${tmpdir}/istio-${istio_version}"

  _output "fetching istio v${istio_version}"
  target="istio-${istio_version}-linux.tar.gz"
  url="https://github.com/istio/istio/releases/download/${istio_version}/${target}"
  curl -Lo ${tmpdir}/${target} --progress-bar $url
  curl -Lo ${tmpdir}/${target}.sha256 --progress-bar ${url}.sha256

  # checksum validation
  _installed sha256sum && shacmd="sha256sum"
  _installed gsha256sum && shacmd="gsha256sum"
  if [[ -z "$shacmd" ]]; then
    _confirm "sha256sum command not found; skip checksum validation?" || exit 0
  else
    _output "validating checksums"
    ( cd ${tmpdir} && $shacmd -c ${target}.sha256 || _errexit "checksum validation failed" )
  fi

  ( cd ${tmpdir} && _output "installing" && tar xfz ${target} )

  (kubectl get namespaces | grep -q istio-system) || {
    kubectl create namespace istio-system
    kubectl label namespace istio-system appconfigmgr.cft.dev/trusted="true"
  }

  docker run --rm -ti -v ${istio_dir}:/apps \
    ${HELM_IMAGE} template \
      install/kubernetes/helm/istio-init \
      --name istio-init \
      --namespace istio-system | kubectl apply -f -

  docker run --rm -ti -v ${istio_dir}:/apps \
    ${HELM_IMAGE} template \
      install/kubernetes/helm/istio \
      --name istio \
      --namespace istio-system \
      --set global.mtls.enabled=true \
      --set grafana.enabled=true \
      --set kiali-enabled=true \
      --set tracing.enabled=true \
      --set global.k8sIngress.enableHttps=true  \
      --set global.disablePolicyChecks=false \
      --set global.outboundTrafficPolicy.mode=REGISTRY_ONLY \
      --values install/kubernetes/helm/istio/values-istio-demo-auth.yaml | kubectl apply -f -

  rm -rf $tmpdir
}

usage() {
  echo "usage: $(basename ${BASH_SOURCE[0]}) <action>"
cat <<EOM
actions:
  help                   display this usage dialog
  status                 show install and repo sync status
  install [-f]           install config management operator and dependencies to active k8s cluster
                         use optional -f flag to force install even when components are found
  init-repo [-c] <path>  initialize a new config root for active k8s cluster within the given repo path
                         use optional -c flag to create and clone a new Cloud Source Git repo prior to initializing
  init-demos <path>      initialize and install demo use case apps within the given repo path
EOM
}

_parseopts() {
  local -a opts args

  [[ $# -eq 0 ]] && { usage; exit 0; }
  export ACTION=$1; shift

  for a in $@; do
    case $a in
      -*)
        opts+=($a)
        ;;
      *)
        args+=($a)
        ;;
    esac
  done

  export OPTS=(${opts[@]})
  export ARGS=(${args[@]})
}

_parseopts $@
case $ACTION in
  help) usage ;;
  status) status ;;
  install) install ;;
  init-repo) init-repo ;;
  init-demos) init-demos ;;
  *) _errexit "invalid action: $1\n\n$($0 help)" ;;
esac
