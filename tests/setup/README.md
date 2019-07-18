# Testing


## Build Environment

```bash
  gcloud builds submit --config=tests/setup/cloudbuild.yaml \
   tests/setup --project anthos-appconfig \
   --substitutions='_BRANCH_NAME=master'

```

       #UC3aTEST=$(curl "http://$$INGRESS_ISTIO_HOST/testcallseq?call1=http://app-allowed-jwt-istio-appconfigv2-service-sm-2.uc-allowed-jwt-istio/testcallseq&call2=http://app-allowed-jwt-istio-appconfigv2-service-sm-3/testcallseq")
        #UC3bTEST=$(docker run -it --rm -v ${KEY_PATH}:/my/credentials -e INGRESS_ISTIO_HOST=$$INGRESS_ISTIO_HOST -e GOOGLE_APPLICATION_CREDENTIALS=/my/credentials/appconfigcrd-demo-sa1.json us.gcr.io/anthos-crd-v1-dev-t2/hello-app-sm-py:v3.0.20 python hello_app_ext_client_py.py --host='devtest.anthos-crd-demo.example.com' --service_name='app-allowed-jwt-istio-appconfigv2-service-sm-' --namespace_name='uc-allowed-jwt-istio' --nested_calls='2,3')
        #UC3cTEST=$(docker run -it --rm -v ${KEY_PATH}:/my/credentials -e INGRESS_ISTIO_HOST=$$INGRESS_ISTIO_HOST -e GOOGLE_APPLICATION_CREDENTIALS=/my/credentials/appconfigcrd-demo-sa1.json us.gcr.io/anthos-crd-v1-dev-t2/hello-app-sm-py:v3.0.20 python hello_app_ext_client_py.py --host='devtest.anthos-crd-demo.example.com' --service_name='app-allowed-jwt-istio-appconfigv2-service-sm-' --namespace_name='uc-allowed-jwt-istio' --nested_calls='2,3' --skip_jwt=X)
        #UC4aTEST=$(curl "http://$$INGRESS_NO_ISTIO_HOST/testcallseq?call1=http://app-secrets-k8s-appconfigv2-service-sm-2.uc-secrets-k8s/testcallseq&call2=http://app-secrets-k8s-appconfigv2-service-sm-1/testcallseq&call3=https://httpbin.org/get")
        #UC4bTEST=$(curl "http://$$INGRESS_NO_ISTIO_HOST/testcallseq?call1=http://app-secrets-k8s-appconfigv2-service-sm-2.uc-secrets-k8s/testcallseq&call2=http://app-secrets-k8s-appconfigv2-service-sm-1/testcallseq&call3=http://app-secrets-k8s-appconfigv2-service-pubsub?gcpProjectID=$$PARM_PROJ&topic=appconfigcrd-demo-topic1&message=hello1")
        #UC4cTEST=$(gcloud pubsub subscriptions pull --auto-ack  appconfigcrd-demo-topic1 --project $$PARM_PROJ)
        #UC4dTEST=$(curl "http://$$INGRESS_NO_ISTIO_HOST/testcallseq?call1=http://app-secrets-k8s-appconfigv2-service-sm-2.uc-secrets-k8s/testcallseq&call2=http://app-secrets-k8s-appconfigv2-service-sm-1/testcallseq&call3=http://app-secrets-k8s-appconfigv2-service-pubsub?gcpProjectID=$$PARM_PROJ&topic=appconfigcrd-demo-topic2&message=hello1")
        #UC5aTEST=$(curl --header 'Host: devtest.anthos-crd-demo.example.com' "http://$$INGRESS_ISTIO_HOST/testcallseq?call1=http://app-secrets-istio-appconfigv2-service-sm-2.uc-secrets-istio/testcallseq&call2=http://app-secrets-istio-appconfigv2-service-sm-3/testcallseq")
        #UC5bTEST=$(curl "http://$$INGRESS_ISTIO_HOST/testcallseq?call1=http://app-secrets-istio-appconfigv2-service-sm-2.uc-secrets-istio/testcallseq&call2=http://app-secrets-istio-appconfigv2-service-sm-3/testcallseq&call3=http://app-secrets-istio-appconfigv2-service-pubsub?gcpProjectID=$$PARM_PROJ&topic=appconfigcrd-demo-topic2&message=hello1")
        #UC5cTEST=$(gcloud pubsub subscriptions pull --auto-ack  appconfigcrd-demo-topic2 --project $$PARM_PROJ)
        #UC5dTEST=$(curl "http://$$INGRESS_ISTIO_HOST/testcallseq?call1=http://app-secrets-istio-appconfigv2-service-sm-2.uc-secrets-istio/testcallseq&call2=http://app-secrets-istio-appconfigv2-service-sm-3/testcallseq&call3=http://app-secrets-istio-appconfigv2-service-pubsub?gcpProjectID=${PROJECT_NAME}&topic=appconfigcrd-demo-topic1&message=hello2")
        #UC6aTEST=$(curl "http://$$INGRESS_ISTIO_HOST/app/tasks.html")
