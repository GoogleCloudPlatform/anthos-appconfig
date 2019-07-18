#!/usr/bin/env bash
        export INGRESS_ISTIO_HOST=$(kubectl -n istio-system get service istio-ingressgateway -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
        export INGRESS_NO_ISTIO_HOST=$(kubectl -n devtest get service test-service-external -o jsonpath='{.status.loadBalancer.ingress[0].ip}')

        UC1aTEST=$(curl "http://$INGRESS_NO_ISTIO_HOST/testcallseq?call1=http://app-allowed-k8s-appconfigv2-service-sm-2.uc-allowed-services-k8s/testcallseq&call2=http://app-allowed-k8s-appconfigv2-service-sm-1/testcallseq&call3=https://httpbin.org/get")
        UC1bTEST=$(curl "http://$INGRESS_NO_ISTIO_HOST/testcallseq?call1=http://app-allowed-k8s-appconfigv2-service-sm-2.uc-allowed-services-k8s/testcallseq&call2=http://app-allowed-k8s-appconfigv2-service-sm-3/testcallseq&call3=https://httpbin.org/get")
        UC2aTEST=$(curl "http://$INGRESS_ISTIO_HOST/testcallseq?call1=http://app-allowed-istio-appconfigv2-service-sm-2.uc-allowed-services-istio/testcallseq&call2=http://app-allowed-istio-appconfigv2-service-sm-3/testcallseq")
        UC2bTEST=$(curl "http://$INGRESS_ISTIO_HOST/testcallseq?call1=http://app-allowed-istio-appconfigv2-service-sm-1.uc-allowed-services-istio/testcallseq&call2=http://app-allowed-istio-appconfigv2-service-sm-3/testcallseq")
        UC3aTEST=$(curl "http://$INGRESS_ISTIO_HOST/testcallseq?call1=http://app-allowed-jwt-istio-appconfigv2-service-sm-2.uc-allowed-jwt-istio/testcallseq&call2=http://app-allowed-jwt-istio-appconfigv2-service-sm-3/testcallseq")

        UC3bTEST=$(python artifacts/drivers/simple-hello/hello_app_ext_client_py.py --host='devtest.anthos-crd-demo.example.com' --service_name='app-allowed-jwt-istio-appconfigv2-service-sm-' --namespace_name='uc-allowed-jwt-istio' --nested_calls='2,3')
        UC3cTEST=$(python artifacts/drivers/simple-hello/hello_app_ext_client_py.py --host='devtest.anthos-crd-demo.example.com' --service_name='app-allowed-jwt-istio-appconfigv2-service-sm-' --namespace_name='uc-allowed-jwt-istio' --nested_calls='2,3' --skip_jwt='X')
        UC4aTEST=$(curl "http://$INGRESS_NO_ISTIO_HOST/testcallseq?call1=http://app-secrets-k8s-appconfigv2-service-sm-2.uc-secrets-k8s/testcallseq&call2=http://app-secrets-k8s-appconfigv2-service-sm-1/testcallseq&call3=https://httpbin.org/get")
        UC4bTEST=$(curl "http://$INGRESS_NO_ISTIO_HOST/testcallseq?call1=http://app-secrets-k8s-appconfigv2-service-sm-2.uc-secrets-k8s/testcallseq&call2=http://app-secrets-k8s-appconfigv2-service-sm-1/testcallseq&call3=http://app-secrets-k8s-appconfigv2-service-pubsub?gcpProjectID=$PARM_PROJ&topic=appconfigcrd-demo-topic1&message=hello1")
        UC4cTEST=$(gcloud pubsub subscriptions pull --auto-ack  appconfigcrd-demo-topic1 --project $PARM_PROJ)
        UC4dTEST=$(curl "http://$INGRESS_NO_ISTIO_HOST/testcallseq?call1=http://app-secrets-k8s-appconfigv2-service-sm-2.uc-secrets-k8s/testcallseq&call2=http://app-secrets-k8s-appconfigv2-service-sm-1/testcallseq&call3=http://app-secrets-k8s-appconfigv2-service-pubsub?gcpProjectID=$PARM_PROJ&topic=appconfigcrd-demo-topic2&message=hello1")
        UC5aTEST=$(curl --header 'Host: devtest.anthos-crd-demo.example.com' "http://$INGRESS_ISTIO_HOST/testcallseq?call1=http://app-secrets-istio-appconfigv2-service-sm-2.uc-secrets-istio/testcallseq&call2=http://app-secrets-istio-appconfigv2-service-sm-3/testcallseq")
        UC5bTEST=$(curl "http://$INGRESS_ISTIO_HOST/testcallseq?call1=http://app-secrets-istio-appconfigv2-service-sm-2.uc-secrets-istio/testcallseq&call2=http://app-secrets-istio-appconfigv2-service-sm-3/testcallseq&call3=http://app-secrets-istio-appconfigv2-service-pubsub?gcpProjectID=$PARM_PROJ&topic=appconfigcrd-demo-topic2&message=hello1")
        UC5cTEST=$(gcloud pubsub subscriptions pull --auto-ack  appconfigcrd-demo-topic2 --project $PARM_PROJ)
        UC5dTEST=$(curl "http://$INGRESS_ISTIO_HOST/testcallseq?call1=http://app-secrets-istio-appconfigv2-service-sm-2.uc-secrets-istio/testcallseq&call2=http://app-secrets-istio-appconfigv2-service-sm-3/testcallseq&call3=http://app-secrets-istio-appconfigv2-service-pubsub?gcpProjectID=${PROJECT_NAME}&topic=appconfigcrd-demo-topic1&message=hello2")
        UC6aTEST=$(curl "http://$INGRESS_ISTIO_HOST/app/tasks.html")



        usecase=()
        usecase_output=()
        usecase_error=()
        usecase+=("UC1aTEST")
        usecase_output+=($UC1aTEST)
        use_case_error+=("N")
        usecase+=("UC1bTEST")
        usecase_output+=($UC1bTEST)
        use_case_error+=("Y")
        usecase+=("UC2aTEST")
        usecase_output+=($UC2aTEST)
        use_case_error+=("N")
        usecase+=("UC2bTEST")
        usecase_output+=($UC2bTEST)
        use_case_error+=("Y")

        for i in {0..3}; do
          if [ "${usecase_error[i]}"  == "Y" ]; then
            echo "ERROR-${usecase[i]}"
          else
            echo "NO-ERROR-${usecase[i]}"
          fi
        done