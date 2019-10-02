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
# 
# Copyright 2019 Google LLC. This software is provided as-is, 
# without warranty or representation for any use or purpose.#
#



set -e

####
# GLOBAL CONFIG VARS
####

# TODO - Add Branch Name
GATEKEEPER_BUCKET="gs://anthos-appconfig_public/acm/anthos-config-management/${RELEASE_NAME}/gatekeeper-config"
TEMPLATE_BUCKET="gs://anthos-appconfig_public/acm/anthos-config-management/${RELEASE_NAME}/acm-crd/config-management-root"
EXAMPLES_BUCKET="gs://anthos-appconfig_public/acm/anthos-config-management/${RELEASE_NAME}/acm-crd-examples/config-management-root"
CM_OPERATOR_BUCKET="gs://config-management-release/released/1.1.0/config-management-operator.yaml"
HELM_IMAGE="alpine/helm:2.13.1"
CM_CRD_COUNT=8

####

# common helper funcs
_red() { echo -ne "\033[31m$@\033[0m"; }
_grn() { echo -ne "\033[32m$@\033[0m"; }
_ylw() { echo -ne "\033[33m$@\033[0m"; }
_output() { echo -e "\n[$(_grn crd-setup-helper)] $@"; }
_errexit() { echo -e "[$(_red error)] $@"; exit 1; }
_installed() { command -v "$1" >/dev/null 2>&1; }
_ensure_path() { mkdir -p $(dirname $1); echo "creating $1"; }
_confirm() {
  local x
  echo "reading"
  prompt=$(echo -ne "$@ \033[32m(y/N)\033[0m")
  read -n1 -p "$prompt" x; echo
  echo "read-$x"
  [ "$x" == "y" ] && return 0
  echo "No!"
  return 1
}

load-gcpvars() {
  export GCP_ACCOUNT=$(gcloud config get-value core/account 2> /dev/null)
#  export PROJECT_NAME=$(gcloud config get-value core/project 2> /dev/null)
  [[ -z "$PROJECT_NAME" ]] || [[ -z "$GCP_ACCOUNT" ]] && \
    _errexit "missing gcloud configuration, run 'gcloud init' to create"
  return 0
}

load-ctxvars() {
  # set cluster variables
  [[ -z $PROJECT_NAME ]] && _errexit "missing Project Name context"
  [[ -z $RELEASE_NAME ]] && _errexit "missing Release Name (use master) context"
  if [[ "${ACTION}" == 'pre-install' ]]; then
    export K8S_CONTEXT="dummy"
  else
    export K8S_CONTEXT=$(kubectl config current-context)
    export ACM_CLUSTER_REGISTRY_NAME=${K8S_CONTEXT//_/-}
    export ACM_ENV_ROOT=./env/${ACM_CLUSTER_REGISTRY_NAME}
    [[ -z $K8S_CONTEXT ]] ||  [[ -z $K8S_CONTEXT ]] || [[ -z $K8S_CONTEXT ]] || return 0
     _errexit "missing k8s context"
  fi
}

load-repovars() {
  echo "load-repovars - args - ${ARGS[@]} - opts -${OPTS[@]}"
  REPO_PATH=${ARGS[0]:-$(pwd)}
  REPO_USER_GCP_CSR="${GCP_ACCOUNT}"
  [ -z "${ARGS[1]}" ] || REPO_USER_GCP_CSR=${ARGS[1]}
  [ -d $REPO_PATH ] || echo "creating directory for repo"; mkdir -p $REPO_PATH
  echo "REPO-PATH-BEFORE-$(pwd)"
  pushd $REPO_PATH
  echo "REPO-PATH-AFTER-$(pwd)"

#  export REPO_PATH="$(dirname ${ARGS[0]})/$(basename ${ARGS[0]})"
  export REPO_NAME=$(basename $REPO_PATH)
  # set repo variables
  export REPO_REMOTE=$(git remote | head -1)
  [[ -z "$REPO_REMOTE" ]] && _errexit "repo missing remote upstream url"
  export REPO_URL=$(git config --get remote.${REPO_REMOTE}.url 2> /dev/null) || _errexit "repo missing remote upstream url"
  export REPO_URL="ssh://${REPO_USER_GCP_CSR}@source.developers.google.com:2022/p/${PROJECT_NAME}/r/${REPO_NAME}"

  # default to master branch on repos with no commit index
  REPO_BRANCH=$(git rev-parse --abbrev-ref HEAD 2> /dev/null) || REPO_BRANCH="master"
  export REPO_BRANCH
}

load-vaultvars() {
  export GCP_ACCOUNT=$(gcloud config get-value core/account 2> /dev/null)
#  export PROJECT_NAME=$(gcloud config get-value core/project 2> /dev/null)
  [[ -z "$PROJECT_NAME" ]] || [[ -z "$GCP_ACCOUNT" ]] && \
    _errexit "missing gcloud configuration, run 'gcloud init' to create"
  return 0
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
  echo "create-repo - args - ${ARGS}"
  [ -e "${ARGS[0]}" ] && _errexit "create-repo:check-path:path exists: $REPO_PATH"
#  [ -d "${ARGS[0]}" ] || mkdir -p "${ARGS[0]}"
#  [ -d "${ARGS[0]}" ] || _errexit "create-repo:check-path:invalid path - ${ARGS[0]}"

  export REPO_PATH="$(dirname ${ARGS[0]})/$(basename ${ARGS[0]})"
  export REPO_NAME=$(basename $REPO_PATH)
  export REPO_URL="ssh://${GCP_ACCOUNT}@source.developers.google.com:2022/p/${PROJECT_NAME}/r/${REPO_NAME}"
  export REPO_REMOTE=origin
  export REPO_BRANCH=master



  _echo_vars PROJECT_NAME REPO_NAME REPO_PATH REPO_URL
  _confirm "\ncreate repo with above configuration?" || exit 0

  _output "creating cloud source repo"
  gcloud source repos create --project $PROJECT_NAME $REPO_NAME

  _output "cloning new repo"
  gcloud source repos clone --project $PROJECT_NAME $REPO_NAME $REPO_PATH

}

init-repo() {
  local create
  echo "init-repo - args - ${ARGS} - opts - ${OPTS} - ${OPTS[@]}"
  for opt in ${OPTS[@]}; do
    case $opt in
      -c) create=1 ;;
      -external) echo "option-external";;
      *) _errexit "unknown install option \"$opt\"";;
    esac
  done
  echo "init-repo-create - create - ${create}"
  load-gcpvars
  load-ctxvars
  [[ -z "$create" ]] || create-repo

  load-repovars

  echo; for v in GCP_ACCOUNT ACM_ENV_ROOT REPO_PATH REPO_REMOTE REPO_BRANCH REPO_URL K8S_CONTEXT ACM_CLUSTER_REGISTRY_NAME PROJECT_NAME; do
    echo -e "\033[32m${v}\033[0m\t| ${!v}"
  done | column -t
  _confirm "\ninitialize repo with above configuration?" || { echo "x" ; popd ; exit 0; }

  [[ -a $ACM_ENV_ROOT ]] && echo "WARN:config root exists: $ACM_ENV_ROOT"

  _output "initializing cluster config root-$(pwd)"

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
  labels:
    appconfigmgr.cft.dev/trusted: "true"
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

apiVersion: configmanagement.gke.io/v1
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

  echo "BEFORE NEXT QUESTION"
  if _confirm "\npush new config?"; then

    git add * || echo "git add empty- might be all checked in - should be ok"
    git status
    echo "git commit"
    git commit -am "auto-initialize $ACM_CLUSTER_REGISTRY_NAME" || echo "git commit - might be empty - should be ok"
    echo "git push"
    git push -q --set-upstream ${REPO_REMOTE} ${REPO_BRANCH}
    echo "kubectl apply $(pwd)"
    kubectl apply -f ${ACM_ENV_ROOT}/config-management.yaml
    echo "Instructions for keys"

    cat <<EOM

# Complete Setup

1. Run the following commands to complete setup:

Create private key in secrets to access repostory

Example:

ssh-keygen -t rsa -b 4096 -N '' -q \
  -C "${ACM_CLUSTER_REGISTRY_NAME}" \
  -f "private key path"

2.  Create secret

kubectl create secret generic git-creds \
--namespace=config-management-system \
--from-file=ssh="private key path"

**************IMPORTANT**************
Delete the private key from the local disk or otherwise protect it.
**************IMPORTANT**************

2. Add the below public SSH key to your repo provider - "public key path + .pub"
$(git-key-url $REPO_URL)

EOM
    popd
    exit 0
  fi

}

init-demos() {
  echo "init-demos - args - ${ARGS[@]} - opts - ${OPTS[@]}"

  load-ctxvars
  load-gcpvars
  load-repovars


  local x app_iters="1 2" app_name="appconfigcrd-demo"

  echo; for v in REPO_REMOTE REPO_BRANCH REPO_URL PROJECT_NAME K8S_CONTEXT ACM_CLUSTER_REGISTRY_NAME; do
    echo -e "\033[32m${v}\033[0m\t| ${!v}"
  done | column -t
  _confirm "\nproceed with above configuration?" || exit 0

  prompt=$(echo -ne "please provide an app name prefix to use (\033[32m$app_name\033[0m) > ")
  read -p "$prompt" x
  [[ -z "$x" ]] || app_name=$x

  _output "creating pubsub topics and subscriptions"
  for i in $app_iters; do
    topic="${app_name}-topic${i}"
    if (gcloud pubsub topics describe $topic --project $PROJECT_NAME &> /dev/null); then
      echo "$topic topic exists, skipping"
    else
      gcloud pubsub topics create $topic --project $PROJECT_NAME
    fi

    if (gcloud pubsub subscriptions describe $topic --project $PROJECT_NAME &> /dev/null); then
      echo "$topic subscription exists, skipping"
    else
      gcloud pubsub subscriptions create $topic --project $PROJECT_NAME\
        --topic ${topic} --topic-project $PROJECT_NAME
    fi
  done

  DEMO_COMMAND=""
  for i in $app_iters; do
    topic="${app_name}-topic${i}"
    iam_name=${app_name}-sa${i}
    iam_account="${iam_name}@${PROJECT_NAME}.iam.gserviceaccount.com"
    DEMO_COMMAND="${DEMO_COMMAND}gcloud iam service-accounts create ${iam_name} --display-name=${iam_name} --project $PROJECT_NAME\n"
    DEMO_COMMAND="${DEMO_COMMAND}gcloud iam service-accounts keys create \"sa${i}.json\" --project $PROJECT_NAME \\\\\n"
    DEMO_COMMAND="${DEMO_COMMAND}  --iam-account=${iam_account}\n\n"
    DEMO_COMMAND="${DEMO_COMMAND}gcloud beta pubsub topics add-iam-policy-binding ${topic} --project $PROJECT_NAME \\\\\n"
    DEMO_COMMAND="${DEMO_COMMAND}  --member=serviceAccount:${iam_account} \\\\\n"
    DEMO_COMMAND="${DEMO_COMMAND}  --role=\"roles/pubsub.publisher\"\n\n"
    DEMO_COMMAND="${DEMO_COMMAND}\nkubectl create secret generic ${iam_name}-secret \\\\\n"
    DEMO_COMMAND="${DEMO_COMMAND}  -n appconfigmgrv2-system \\\\\n"
    DEMO_COMMAND="${DEMO_COMMAND}  --from-file=key.json=sa${i}.json\n\n"

  done


  [[ -d "${ACM_ENV_ROOT}/namespaces/use-cases" ]] && {
    _output "WARN: ${ACM_ENV_ROOT}/namespaces/use-cases already exists"
  }

  _output "adding demo apps to policy config repo"
  echo "Current Dir:["$(pwd)
  gsutil -m cp -R "${EXAMPLES_BUCKET}/*" ${ACM_ENV_ROOT}/


  git add ${ACM_ENV_ROOT} || echo "1"

  git commit -am "initialize $ACM_CLUSTER_REGISTRY_NAME demo apps" && git push

  cat <<EOM

# Complete Setup

Run the following commands to complete setup:

 - Create two service accounts and JSON Keys and corresponding subscriptions and secrets (to test pubsub ACL)

EOM

  echo -e "${DEMO_COMMAND}"
  echo -e '\n\n - (OPTIONAL) Enable and configure ACM Vault integration to your existing Vault Server'
  cat <<EOM

export VAULT_ADDR=<vault_addr>
export VAULT_CACERT=</path/to/vault/ca.pem>
export VAULT_TOKEN="<vault token, to execute vault commands>"

kubectl create secret generic vault-ca \
  --namespace=appconfigmgrv2-system \
  --from-file=${VAULT_CACERT}

bash vault-setup-helper-vault-gcp-sa.sh \
  -p ${PROJECT_NAME} --app-prefix app-crd-vault --cluster ${ACM-CLUSTER-REGISTRY} \
  --key-path ./key.json --role uc-secrets-vault-k8s \
  --role-script-path vault-roles-policy.sh  \
  --ns uc-secrets-vault-k8s

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
  echo "config-management-crds(status): $n"
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
  echo "config-management-operator(status): $n"
  if [[ -z "$n" ]]; then
    _red "MISSING\n"
  elif [[ "$n" -ge 1 ]]; then
    _grn "OK\n"
  else
    _ylw "PENDING\n"
  fi
}

_sync_status() {
  echo -e 'COMPONENT,LAST_UPDATE,TOKEN\n'
  kubectl get repos.configmanagement.gke.io repo \
    -o='go-template' \
    --template='go-template' --template='source,-,{{ .status.source.token }}
git_importer,{{ .status.import.lastUpdate }},{{ .status.import.token }}{{printf "\n"}}
git_syncer,{{ .status.sync.lastUpdate }},{{ .status.sync.latestToken }}{{printf "\n"}}'
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
      -external) echo "option-external";;
      *) _errexit "unknown install option \"$opt\"";;
    esac
  done
  CRD_SETUP_ISTIO_PREINSTALL_DIR="${ARGS[0]}"

  echo; for v in K8S_CONTEXT ACM_CLUSTER_REGISTRY_NAME PROJECT_NAME; do
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

  # gatekeeper install
  n=1
  if [[ -n "$force" ]]; then
    n=0
  else
    (kubectl get namespaces gatekeeper-system &> /dev/null) || n=0
  fi

  [[ "$n" -eq 1 ]] || install_gatekeeper && _output "gatekeeper install OK"

  _output "done"
}

pre-install() {
  load-ctxvars

  CRD_SETUP_ISTIO_PREINSTALL_DIR="${ARGS[0]}"
  install_istio && _output "istio manual install OK ${CRD_SETUP_ISTIO_PREINSTALL_DIR}"
#  install_gatekeeper && _output "gatekeeper manual install OK"

  _output "done"
}

install_operator() {
  _output "installing config management operator to cluster"
  gsutil ls ${CM_OPERATOR_BUCKET} && gsutil cat ${CM_OPERATOR_BUCKET} | kubectl apply -f -
}

install_gatekeeper() {
  kubectl apply -f https://raw.githubusercontent.com/open-policy-agent/gatekeeper/v3.0.4-beta.1/deploy/gatekeeper.yaml
  gsutil cat ${GATEKEEPER_BUCKET}/config.yaml | kubectl apply -f -
  sleep 10
  gsutil cat ${GATEKEEPER_BUCKET}/constraint-templates.yaml | kubectl apply -f -

  n=0
  until [ $n -ge 50 ]
  do
    echo "attempting to install constraints (attempt $n)"
    gsutil cat ${GATEKEEPER_BUCKET}/constraints.yaml | kubectl apply -f - && break
    n=$[$n+1]
    sleep 20
  done
}

install_istio() {
  local x tmpdir shacmd forcedir genexternal
  _output "installing istio to cluster"
  forcedir=0
  genexternal=0
  for opt in ${OPTS[@]}; do
    case $opt in
      -external) forcedir=1; _output "force install istio from dir current dir $(pwd)/istio-external" ;;
      -genexternal) genexternal=1; _output "general install istio from dir current dir $(pwd)/istio-external" ;;
    esac
  done

  if [ "${genexternal}" == "1" ]; then
    tmpdir="${CRD_SETUP_ISTIO_PREINSTALL_DIR}"
    rm -rf "$(pwd)/istio-external"
    mkdir -p ${tmpdir}
    istio_dir="${tmpdir}/istio-generated"
  fi


  if [ "${forcedir}" == "1" ] ; then
    (kubectl get namespaces | grep -q istio-system) || {
      kubectl create namespace istio-system
      kubectl label namespace istio-system appconfigmgr.cft.dev/trusted="true"
    }
    cat ${CRD_SETUP_ISTIO_PREINSTALL_DIR}/istio-init.yaml | kubectl apply -f -
    sleep 30
    cat ${CRD_SETUP_ISTIO_PREINSTALL_DIR}/istio-init.yaml | kubectl apply -f -
    sleep 30
    cat ${CRD_SETUP_ISTIO_PREINSTALL_DIR}/istio.yaml | kubectl apply -f -
  else
    istio_version=$(curl -L -s https://api.github.com/repos/istio/istio/releases | \
      grep -E 'tag_name.*1\.1\.' | grep -vE '(\-rc|\-snapshot)' | \
      sed "s/ *\"tag_name\": *\"\\(.*\\)\",*/\\1/" | head -1)
    prompt=$(echo -ne "istio version to install? \033[32m($istio_version)\033[0m ")
    read -p "$prompt" x
    [[ -z "$x" ]] || istio_version=$x

    if [ -z "${tmpdir}" ]; then
     tmpdir=$(mktemp -d $(pwd)/.acm-init-XXXXXX)
     istio_dir="${tmpdir}/istio-${istio_version}"
    fi


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

    if [ "${genexternal}" == "1" ]; then
      echo "skipping-due to external generation of yaml"
      mv ${tmpdir}/istio-${istio_version} ${tmpdir}/istio-build
#      ls -lR ${tmpdir}
    else
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
    fi
  fi
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
    case "$a" in
      -*)
        opts+=("$a")
        ;;
      *)
        args+=("$a")
       ;;
    esac
  done

  OPTS=(${opts[@]})
  ARGS=(${args[@]})

}

_parseopts $@

echo "main - args - ${ARGS[@]} - opts -${OPTS[@]}"
#set -x

case $ACTION in
  help) usage ;;
  status) status ;;
  pre-install) pre-install ;;
  install) install ;;
  init-repo) init-repo ;;
  init-demos) init-demos ;;
  *) _errexit "invalid action: $1\n\n$($0 help)" ;;
esac
