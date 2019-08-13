#!/usr/bin/python
#
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


# @parm $APPCONFIG_CRD_PREFIX
# @parm CLUSTER
#
#
get_vault_provider_name() {
  echo "${1}-${2}"

}

# @parm PROJECT_NAME
# @parm APPCONFIG_CRD_PREFIX
#
#
get_vault_service_account_name() {
  echo "${APPCONFIG_CRD_PREFIX}-vault-sa1"
}

# @parm PROJECT_NAME
# @parm APPCONFIG_CRD_PREFIX
#
#


setup_service_account() {
  local PROJECT_NAME=$1
  local APPCONFIG_CRD_PREFIX=$2
  local CLUSTER=$3
  local VAULT_SA_KEY_PATH=$4
  local VAULT_ROLE_NAME=$5
  local VAULT_ROLE_CREATE_SCRIPT=$6
  local VAULT_NS=$7
  local VAULT_KSA="${VAULT_NS}-ksa"
  local KEY_CHECK=$8

  local VAULT_PREFIX="k8s-$(get_vault_provider_name $APPCONFIG_CRD_PREFIX $CLUSTER)"
  local VAULT_SA_EMAIL="$(get_vault_service_account_name $PROJECT_NAME $APPCONFIG_CRD_PREFIX})@${PROJECT_NAME}.iam.gserviceaccount.com"
  local GCP_VAULT_PREFIX="gcp-$APPCONFIG_CRD_PREFIX-$PROJECT_NAME"

  echo; for v in PROJECT_NAME APPCONFIG_CRD_PREFIX CLUSTER VAULT_SA_KEY_PATH VAULT_PREFIX VAULT_SA_EMAIL VAULT_ROLE_NAME VAULT_ROLE_CREATE_SCRIPT VAULT_NS ; do
    echo -e "\033[32m${v}\033[0m\t| ${!v}"
  done | column -t

  kubectl delete configmap vault \
      --namespace=appconfigmgrv2-system

  kubectl create configmap vault \
      --namespace=appconfigmgrv2-system \
      --from-literal vault-addr=${VAULT_ADDR} \
      --from-literal vault-cluster-path=${VAULT_PREFIX} \
      --from-literal gcp-vault-path=${GCP_VAULT_PREFIX}

  CHECK_VAULT_SA1=$(gcloud iam service-accounts describe ${VAULT_SA_EMAIL}  \
    --project ${PROJECT_NAME}  --format="value(name)"   || echo "")
  [ ! -z  $CHECK_VAULT_SA1 ] || gcloud iam service-accounts create $(get_vault_service_account_name $PROJECT_NAME $APPCONFIG_CRD_PREFIX}) \
    --display-name=$(get_vault_service_account_name $PROJECT_NAME $APPCONFIG_CRD_PREFIX}) --project ${PROJECT_NAME}


  if [ ! -f ${VAULT_SA_KEY_PATH} ] ; then
    if [ ! -f ${KEY_CHECK}/vault.json ] ; then
      gcloud iam service-accounts keys create "${KEY_CHECK}/vault.json" --project ${PROJECT_NAME} \
        --iam-account=${VAULT_SA_EMAIL}
    fi
    cp ${KEY_CHECK}/vault.json ${VAULT_SA_KEY_PATH}
  fi

  CHECK_GCP_1=$(vault read "${GCP_VAULT_PREFIX}" || echo "")
  [ ! -z  $CHECK_GCP_1 ] || (vault secrets enable --path="${GCP_VAULT_PREFIX}" gcp || echo "")

  CHECK_K8S_1=$(vault read "${VAULT_PREFIX}" || echo "")
  [ ! -z  $CHECK_K8S_1 ] || (vault auth enable --path="${VAULT_PREFIX}" kubernetes || echo "")

  CHECK_GCP_2=$(vault read "${GCP_VAULT_PREFIX}/config" || echo "")
  [[ ! -z  "$CHECK_GCP_2" ]] || vault write ${GCP_VAULT_PREFIX}/config project=${PROJECT_NAME} \
    ttl=3600 \
    max_ttl=7200 \
    credentials=@${VAULT_SA_KEY_PATH}



  VAULT_SA_SECRET=$(kubectl get -n appconfigmgrv2-system sa vault-auth -o jsonpath="{.secrets[*]['name']}")
  VAULT_SA_JWT_TOKEN=$(kubectl get -n appconfigmgrv2-system secret $VAULT_SA_SECRET -o jsonpath="{.data.token}" | base64 --decode; echo)
  VAULT_SA_CA_CRT=$(kubectl get -n appconfigmgrv2-system secret $VAULT_SA_SECRET -o jsonpath="{.data['ca\.crt']}" | base64 --decode; echo)
  VAULT_REVIEWER_CLUSTER=$(kubectl config current-context)
  VAULT_REVIEWER_CLIENT_API_SERVER=$(kubectl config view -o jsonpath="{.clusters[?(@.name==\"${VAULT_REVIEWER_CLUSTER}\")].cluster.server}")

  CHECK_K8S_2=$(vault read "auth/${VAULT_PREFIX}/config")
  [ ! -z  $CHECK_K8S_2 ] || vault write auth/${VAULT_PREFIX}/config \
    token_reviewer_jwt="${VAULT_SA_JWT_TOKEN}" \
    kubernetes_host="${VAULT_REVIEWER_CLIENT_API_SERVER}" \
    kubernetes_ca_cert="${VAULT_SA_CA_CRT}"



  gcloud projects add-iam-policy-binding ${PROJECT_NAME} \
    --member=serviceAccount:${VAULT_SA_EMAIL} \
    --role roles/pubsub.admin

  gcloud projects add-iam-policy-binding  ${PROJECT_NAME} \
    --member=serviceAccount:${VAULT_SA_EMAIL} \
    --role roles/iam.serviceAccountAdmin

  gcloud projects add-iam-policy-binding  ${PROJECT_NAME} \
    --member=serviceAccount:${VAULT_SA_EMAIL} \
    --role roles/iam.serviceAccountKeyAdmin

  export ROLE_NAME=${VAULT_ROLE_NAME}
  export GCP_VAULT_PREFIX=$GCP_VAULT_PREFIX

  . ${VAULT_ROLE_CREATE_SCRIPT}

  vault policy write ${VAULT_ROLE_NAME} ./${VAULT_ROLE_NAME}-policy.hcl

  vault write ${GCP_VAULT_PREFIX}/roleset/${VAULT_ROLE_NAME} \
    project="${PROJECT_NAME}" \
    secret_type="service_account_key"  \
    bindings=@${VAULT_ROLE_NAME}-gcp.hcl \




  vault write auth/${VAULT_PREFIX}/role/${VAULT_ROLE_NAME} \
    bound_service_account_names="${VAULT_KSA}" \
    bound_service_account_namespaces="${VAULT_NS}" \
    policies=${VAULT_ROLE_NAME} \
    ttl=3600


}





#Check the number of arguments. If none are passed, print help and exit.
NUMARGS=$#
# echo -e \\n"Number of arguments: $NUMARGS"
if [ $NUMARGS -eq 0 ]; then
  HELP
fi

OPTS=`getopt -o vp:h --long "verbose,project:,help,app-prefix:,cluster:,key-path:,role:,role-script-path:,ns:,key-check:" -n "$0" -- "$@"`

if [ $? != 0 ] ; then echo "Failed parsing options." >&2 ; exit 1 ; fi

echo "$OPTS"
eval set -- "$OPTS"
echo "$OPTS"

VERBOSE=false
HELP=false

PROJECT=""
APP_PREFIX=""
APP_CLUSTER=""
KEY_PATH=""
ROLE=""
ROLE_SCRIPT=""
NS=""
KEY_CHECK=""

while true ; do
  case "$1" in
    -v | --verbose ) VERBOSE=true; shift ;;
    -h | --help )    HELP=true; shift ;;
    -p | --project ) PROJECT=$2; shift; shift ;;
    --app-prefix ) APP_PREFIX="$2"; shift; shift ;;
    --cluster ) APP_CLUSTER="$2"; shift; shift ;;
    --key-path ) KEY_PATH="$2"; shift; shift ;;
    --role ) ROLE="$2"; shift; shift ;;
    --role-script-path ) ROLE_SCRIPT="$2"; shift; shift ;;
    --ns ) NS="$2"; shift; shift ;;
    --key-check ) KEY_CHECK="$2"; shift; shift ;;
    -- ) shift; break ;;
    * ) break ;;
  esac
done

echo; for v in PROJECT APP_PREFIX APP_CLUSTER KEY_PATH ROLE ROLE_SCRIPT NS ; do
  echo -e "\033[32m${v}\033[0m\t| ${!v}"
done | column -t

set -x
setup_service_account "$PROJECT" "$APP_PREFIX" "$APP_CLUSTER" "$KEY_PATH" "$ROLE" "$ROLE_SCRIPT" "$NS" "$KEY_CHECK"