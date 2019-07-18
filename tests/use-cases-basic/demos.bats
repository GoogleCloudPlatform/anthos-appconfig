#!/usr/bin/env bats
# docker run -it -v "/Users/joseret/go112/src/github.com/GoogleCloudPlatform/anthos-appconfig/tests:/code" bats/bats:latest

@test "uc-1" {
  COMMAND="curl \""
  COMMAND="$COMMANDhttp://${INGRESS_NO_ISTIO_HOST}/testcallseq?"
  COMMAND="$COMMANDcall1=http://app-allowed-k8s-appconfigv2-service-sm-2.uc-allowed-services-k8s/testcallseq&"
  COMMAND="$COMMANDcall2=http://app-allowed-k8s-appconfigv2-service-sm-1/testcallseq&"
  COMMAND="$COMMANDccall3=https://httpbin.org/get\""
  run "curl https://www.google.com"
  [ "$status" -eq 127 ]
}

