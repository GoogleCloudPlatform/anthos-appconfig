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
EXAMPLES_BUCKET="gs://anthos-appconfig_public/acm/anthos-config-management/build-script-2019-07-10/acm-crd-examples/config-management-root"
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

usage() {
  echo "usage: $(basename ${BASH_SOURCE[0]}) <action>"
  echo -e "\nactions:"
  echo "  help       display this usage dialog"
  echo "  status     show install and repo sync status"
  echo "  install    install config management operator and dependencies to active k8s cluster"
  echo "  init-repo  initialize a new config root for active k8s cluster"
}

load-ctxvars() {
  # set cluster variables
  export K8S_CONTEXT=$(kubectl config current-context)
  export ACM_CLUSTER_REGISTRY_NAME=${K8S_CONTEXT//_/-}
  export ACM_ENV_ROOT=./env/${ACM_CLUSTER_REGISTRY_NAME}
}

load-repovars() {
  REPO_PATH=${@:-$(pwd)}
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

init-repo() {
  load-ctxvars
  load-repovars $@

  [[ -a $ACM_ENV_ROOT ]] && _errexit "config root exists: $ACM_ENV_ROOT"

  echo; for v in REPO_REMOTE REPO_BRANCH REPO_URL K8S_CONTEXT ACM_CLUSTER_REGISTRY_NAME; do
    echo -e "\033[32m${v}\033[0m\t| ${!v}"
  done | column -t
  _confirm "\nproceed with above configuration?" || exit 0

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

  echo; for v in K8S_CONTEXT ACM_CLUSTER_REGISTRY_NAME; do
    echo -e "\033[32m${v}\033[0m\t| ${!v}"
  done | column -t
  _confirm "\nproceed with above configuration?" || exit 0

  local n=1
  (kubectl get crds | grep -q configmanagements.addons.sigs.k8s.io) || n=0
  (kubectl get deployments config-management-operator --namespace=kube-system &> /dev/null) || n=0

  [[ "$n" -eq 1 ]] || {
    _output "installing config management operator to cluster"
    gsutil cp ${CM_OPERATOR_BUCKET} - | kubectl apply -f -
  } && _output "config management operator install OK"

  (kubectl get namespaces istio-system &> /dev/null) || {
    _output "installing istio to cluster"
    install_istio
  } && _output "istio install OK"

  _output "done"
}

install_istio() {
  # TODO - Specify Istio Release (istio fails with mount), check istio version no blank don't create stuff until
#  istio_version=$(curl -L -s https://github.com/istio/istio/releases/tag/1.1.9 | \
#                    grep tag_name | sed "s/ *\"tag_name\": *\"\\(.*\\)\",*/\\1/")
   istio_version=1.1.9
#  tmpdir=$(mktemp -d -t acm-init.XXXXX)
  tmpdir="/Users/$USER/tmp/acm-init/downloads/$(date +%Y%m%d-%H%M%S)"
  mkdir -p $tmpdir

  istio_dir="${tmpdir}/istio-${istio_version}"

  _output "fetching istio v${istio_version}"
  URL="https://github.com/istio/istio/releases/download/${istio_version}/istio-${istio_version}-linux.tar.gz"
  curl -sL $URL | tar -C $tmpdir -xz

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
}

acm-init() {
  [[ $# -eq 0 ]] && { usage; exit 0; }
  case $1 in
    help) usage ;;
    status) status ;;
    install) install ;;
    init-repo) shift; init-repo $@;;
    *) _errexit "invalid action: $1\n\n$($0 help)" ;;
  esac
}

[[ "${BASH_SOURCE[0]}" != "${0}" ]] || acm-init $@