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
# without warranty or representation for any use or purpose.
#

import os
import sys
import time
import unittest



sys.path.append(os.path.abspath('../simple_hello'))

import HtmlTestRunner
from auth_helper import GCPAuthHelper  # this doesn't work without 'sys.path.append' per step 2 above
from http_rest_helper import RestHelper


class SimpleHelloTestCase(unittest.TestCase):

  def get_url(self, host, service_name, namespace_name, nested_calls):
    full_url = 'http://' + host + '/testcallseq?'
    for k,v in enumerate(nested_calls):
      full_url += 'call' + str(k+1) + '=http://' + service_name + str(v) + '.' + namespace_name + '/testcallseq&'

    return full_url

  def util(self, parm_call, parm_svc, parm_ns):
    full_url = "call" + parm_call + "=http://" + parm_svc
    if parm_ns:
      full_url = full_url + "." + parm_ns
    full_url += "/testcallseq"
    return full_url

  def warmup(self, up_to_number, sleep_seconds, check, url):

    for i in range(1,up_to_number):
      headers = {}
      response = RestHelper(url).get_text(None, headers)
      if len(response) >  0:
        if check in response:
          return
      time.sleep(sleep_seconds)
    return

  def test_simple_hello_uc3_empty_token(self):

    token_gcf = None
    self.assertIn('INGRESS_ISTIO_HOST', os.environ, "INGRESS_ISTIO_HOST environment variable not set")
    self.assertTrue(len(os.environ['INGRESS_ISTIO_HOST']) >  0, "INGRESS_ISTIO_HOST empty - len == 0")
    full_url = self.get_url(os.environ['INGRESS_ISTIO_HOST'],
                            'app-allowed-jwt-istio-appconfigv2-service-sm-', 'uc-allowed-jwt-istio', [ '2', '3' ])
    headers = {"Host": "test-simple-hello.example.com"}
    response = RestHelper(full_url).get_text(token_gcf,headers)
    self.assertTrue(len(response) >  0, "response empty - len == 0")
    self.assertIn("401", response, "Failed Test")

  def test_simple_hello_uc3_with_google_com_token(self):

    self.assertIn('GOOGLE_APPLICATION_CREDENTIALS', os.environ, "INGRESS_ISTIO_HOST environment variable not set")
    self.assertTrue(len(os.environ['INGRESS_ISTIO_HOST']) >  0, "INGRESS_ISTIO_HOST empty - len == 0")
    o_auth_helper = GCPAuthHelper()
    token_gcf = o_auth_helper.get_google_open_id_connect_token("https://site2.ecom1.joecloudy.com")
    full_url = self.get_url(os.environ['INGRESS_ISTIO_HOST'],
                            'app-allowed-jwt-istio-appconfigv2-service-sm-', 'uc-allowed-jwt-istio', [ '2', '3' ])
    headers = {"Host": "test-simple-hello.example.com"}
    response = RestHelper(full_url).get_text(token_gcf,headers)
    self.assertTrue(len(response) >  0, "response empty - len == 0")
    self.assertIn("Last Call Successful", response, "Failed Test")



  def test_simple_hello_uc1_outbound_ok(self):
    uc = "uc-allowed-services-k8s"
    self.assertTrue(len(os.environ['INGRESS_NO_ISTIO_HOST']) >  0, "INGRESS_NO_ISTIO_HOST empty - len == 0")
    full_url = "http://" + os.environ['INGRESS_NO_ISTIO_HOST'] + "/testcallseq?"
    full_url = full_url + self.util("1", "app-allowed-k8s-appconfigv2-service-sm-2", uc)
    full_url = full_url + "&" + self.util("2", "app-allowed-k8s-appconfigv2-service-sm-1", uc)
    full_url = full_url + "&call3=https://httpbin.org/get"


    headers = {"Host": "test-simple-hello.example.com"}
    response = RestHelper(full_url).get_text(None,headers)
    self.assertTrue(len(response) >  0, "response empty - len == 0")
    self.assertIn('"User-Agent": "python-requests/2.22.0"', response, "Failed Test")
    self.assertIn('"Host": "httpbin.org"', response, "Failed Test")

  def test_simple_hello_uc1_service_blocked(self):
    uc = "uc-allowed-services-k8s"
    self.assertTrue(len(os.environ['INGRESS_NO_ISTIO_HOST']) >  0, "INGRESS_ISTIO_HOST empty - len == 0")
    full_url = "http://" + os.environ['INGRESS_NO_ISTIO_HOST'] + "/testcallseq?"
    full_url = full_url + self.util("1", "app-allowed-k8s-appconfigv2-service-sm-2", uc)
    full_url = full_url + "&" + self.util("2", "app-allowed-k8s-appconfigv2-service-sm-3", uc)
    full_url = full_url + "&call3=https://httpbin.org/get"


    headers = {"Host": "test-simple-hello.example.com"}
    try:
      response = RestHelper(full_url).get_text(None,headers)
    except:
      print('exception')
      return

    self.fail('Should fail with Timeout')

  def test_simple_hello_uc2_service_ok(self):
    uc = "uc-allowed-services-istio"
    self.assertTrue(len(os.environ['INGRESS_ISTIO_HOST']) >  0, "INGRESS_ISTIO_HOST empty - len == 0")
    full_url = "http://" + os.environ['INGRESS_ISTIO_HOST'] + "/testcallseq?"
    full_url = full_url + self.util("1", "app-allowed-istio-appconfigv2-service-sm-2", uc)
    full_url = full_url + "&" + self.util("2", "app-allowed-istio-appconfigv2-service-sm-3", uc)

    headers = {}
    response = RestHelper(full_url).get_text(None,headers)
    self.assertTrue(len(response) >  0, "response empty - len == 0")
    self.assertNotIn('*Error*', response, "Failed Test")
    self.assertNotIn('403', response, "Failed Test- 403")
    self.assertNotIn('PERMISSION_DENIED', response, "Failed Test - Denied Text")

  def test_simple_hello_uc2_service_blocked(self):
    uc = "uc-allowed-services-istio"


    self.assertTrue(len(os.environ['INGRESS_ISTIO_HOST']) >  0, "INGRESS_ISTIO_HOST empty - len == 0")
    full_url = "http://" + os.environ['INGRESS_ISTIO_HOST'] + "/testcallseq?"
    full_url = full_url + self.util("1", "app-allowed-istio-appconfigv2-service-sm-1", uc)
    full_url = full_url + "&" + self.util("2", "app-allowed-istio-appconfigv2-service-sm-3", uc)

    # self.warmup(10,30,"403", full_url)

    headers = {}
    response = RestHelper(full_url).get_text(None,headers)
    self.assertTrue(len(response) >  0, "response empty - len == 0")
    self.assertIn('403', response, "Failed Test- 403-[" + full_url)
    self.assertIn('PERMISSION_DENIED', response, "Failed Test - Denied Text-[" + full_url)

  def test_simple_hello_uc4_outbound_ok(self):
    uc = "uc-secrets-k8s"
    self.assertTrue(len(os.environ['INGRESS_NO_ISTIO_HOST']) >  0, "INGRESS_NO_ISTIO_HOST empty - len == 0")
    full_url = "http://" + os.environ['INGRESS_NO_ISTIO_HOST'] + "/testcallseq?"
    full_url = full_url + self.util("1", "app-secrets-k8s-appconfigv2-service-sm-2", uc)
    full_url = full_url + "&" + self.util("2", "app-secrets-k8s-appconfigv2-service-sm-1", uc)
    full_url = full_url + "&call3=https://httpbin.org/get"


    headers = {"Host": "test-simple-hello.example.com"}
    response = RestHelper(full_url).get_text(None,headers)
    self.assertTrue(len(response) >  0, "response empty - len == 0")
    self.assertIn('"User-Agent": "python-requests/2.22.0"', response, "Failed Test")
    self.assertIn('"Host": "httpbin.org"', response, "Failed Test")

  def test_simple_hello_uc4_pubsub_ok(self):
    uc = "uc-secrets-k8s"
    self.assertIn('INGRESS_NO_ISTIO_HOST', os.environ, "INGRESS_NO_ISTIO_HOST environment variable not set")
    self.assertTrue(len(os.environ['INGRESS_NO_ISTIO_HOST']) >  0, "INGRESS_NO_ISTIO_HOST empty - len == 0")
    self.assertIn('PUBSUB_GCP_PROJECT', os.environ, "PUBSUB_GCP_PROJECT environment variable not set")
    self.assertTrue(len(os.environ['PUBSUB_GCP_PROJECT']) >  0, "PUBSUB_GCP_PROJECT empty - len == 0")
    full_url = "http://" + os.environ['INGRESS_NO_ISTIO_HOST'] + "/testcallseq?"
    full_url = full_url + self.util("1", "app-secrets-k8s-appconfigv2-service-sm-2", uc)
    full_url = full_url + "&" + self.util("2", "app-secrets-k8s-appconfigv2-service-sm-1", uc)
    full_url = full_url + "&call3=http://app-secrets-k8s-appconfigv2-service-pubsub?"
    full_url = full_url + "gcpProjectID=" + os.environ['PUBSUB_GCP_PROJECT']
    full_url = full_url + "&topic=appconfigcrd-demo-topic1&message=hello1"


    headers = {"Host": "test-simple-hello.example.com"}
    response = RestHelper(full_url).get_text(None,headers)
    self.assertTrue(len(response) >  0, "response empty - len == 0")
    self.assertIn('Publish Success:', response, "Failed Test - Publish")

  def test_simple_hello_uc4_pubsub_topic_acl_not_allowed(self):
    uc = "uc-secrets-k8s"
    self.assertIn('INGRESS_NO_ISTIO_HOST', os.environ, "INGRESS_NO_ISTIO_HOST environment variable not set")
    self.assertTrue(len(os.environ['INGRESS_NO_ISTIO_HOST']) >  0, "INGRESS_NO_ISTIO_HOST empty - len == 0")
    self.assertIn('PUBSUB_GCP_PROJECT', os.environ, "PUBSUB_GCP_PROJECT environment variable not set")
    self.assertTrue(len(os.environ['PUBSUB_GCP_PROJECT']) >  0, "PUBSUB_GCP_PROJECT empty - len == 0")
    full_url = "http://" + os.environ['INGRESS_NO_ISTIO_HOST'] + "/testcallseq?"
    full_url = full_url + self.util("1", "app-secrets-k8s-appconfigv2-service-sm-2", uc)
    full_url = full_url + "&" + self.util("2", "app-secrets-k8s-appconfigv2-service-sm-1", uc)
    full_url = full_url + "&call3=http://app-secrets-k8s-appconfigv2-service-pubsub?"
    full_url = full_url + "gcpProjectID=" + os.environ['PUBSUB_GCP_PROJECT']
    full_url = full_url + "&topic=appconfigcrd-demo-topic2&message=hello1"


    headers = {"Host": "test-simple-hello.example.com"}
    response = RestHelper(full_url).get_text(None,headers)
    self.assertTrue(len(response) >  0, "response empty - len == 0")
    self.assertIn('PermissionDenied', response, "Failed Test - Publish")

  def test_simple_hello_uc5_simple_call_ok(self):
    uc = "uc-secrets-istio"
    self.assertIn('INGRESS_ISTIO_HOST', os.environ, "INGRESS_ISTIO_HOST environment variable not set")
    self.assertTrue(len(os.environ['INGRESS_ISTIO_HOST']) >  0, "INGRESS_ISTIO_HOST empty - len == 0")
    full_url = "http://" + os.environ['INGRESS_ISTIO_HOST'] + "/testcallseq?"
    full_url = full_url + self.util("1", "app-secrets-istio-appconfigv2-service-sm-2", uc)
    full_url = full_url + "&" + self.util("2", "app-secrets-istio-appconfigv2-service-sm-3", None)


    headers = {"Host": "test-simple-hello.example.com"}
    response = RestHelper(full_url).get_text(None,headers)
    self.assertTrue(len(response) >  0, "response empty - len == 0")
    self.assertIn('appconfigv2-service-sm-3', response, "Failed Test - Last Call")
    self.assertNotIn('*Error*', response, "Failed Test - Error?")

  def test_simple_hello_uc5_simple_external_not_google_apis(self):
    uc = "uc-secrets-istio"
    self.assertIn('INGRESS_ISTIO_HOST', os.environ, "INGRESS_ISTIO_HOST environment variable not set")
    self.assertTrue(len(os.environ['INGRESS_ISTIO_HOST']) >  0, "INGRESS_ISTIO_HOST empty - len == 0")
    full_url = "http://" + os.environ['INGRESS_ISTIO_HOST'] + "/testcallseq?"
    full_url = full_url + self.util("1", "app-secrets-istio-appconfigv2-service-sm-2", uc)
    full_url = full_url + "&" + self.util("2", "app-secrets-istio-appconfigv2-service-sm-3", None)
    full_url = full_url + "&call3=https://httpbin.org/get"

    headers = {"Host": "test-simple-hello.example.com"}
    response = RestHelper(full_url).get_text(None,headers)
    self.assertTrue(len(response) >  0, "response empty - len == 0")
    self.assertIn('*Error*', response, "Failed Test - Last Call")

  def test_simple_hello_uc5_pubsub_ok(self):
    uc = "uc-secrets-istio"
    self.assertIn('INGRESS_ISTIO_HOST', os.environ, "INGRESS_ISTIO_HOST environment variable not set")
    self.assertTrue(len(os.environ['INGRESS_ISTIO_HOST']) >  0, "INGRESS_ISTIO_HOST empty - len == 0")
    self.assertIn('PUBSUB_GCP_PROJECT', os.environ, "PUBSUB_GCP_PROJECT environment variable not set")
    self.assertTrue(len(os.environ['PUBSUB_GCP_PROJECT']) >  0, "PUBSUB_GCP_PROJECT empty - len == 0")
    full_url = "http://" + os.environ['INGRESS_ISTIO_HOST'] + "/testcallseq?"
    full_url = full_url + self.util("1", "app-secrets-istio-appconfigv2-service-sm-2", uc)
    full_url = full_url + "&" + self.util("2", "app-secrets-istio-appconfigv2-service-sm-3", uc)
    full_url = full_url + "&call3=http://app-secrets-istio-appconfigv2-service-pubsub?"
    full_url = full_url + "gcpProjectID=" + os.environ['PUBSUB_GCP_PROJECT']
    full_url = full_url + "&topic=appconfigcrd-demo-topic2&message=hello2"

    headers = {"Host": "test-simple-hello.example.com"}
    response = RestHelper(full_url).get_text(None,headers)
    self.assertTrue(len(response) >  0, "response empty - len == 0")
    self.assertIn('Publish Success:', response, "Failed Test - Publish")

  def test_simple_hello_uc5_pubsub_topic_acl_not_allowed(self):
    uc = "uc-secrets-istio"
    self.assertIn('INGRESS_ISTIO_HOST', os.environ, "INGRESS_ISTIO_HOST environment variable not set")
    self.assertTrue(len(os.environ['INGRESS_ISTIO_HOST']) >  0, "INGRESS_ISTIO_HOST empty - len == 0")
    self.assertIn('PUBSUB_GCP_PROJECT', os.environ, "PUBSUB_GCP_PROJECT environment variable not set")
    self.assertTrue(len(os.environ['PUBSUB_GCP_PROJECT']) >  0, "PUBSUB_GCP_PROJECT empty - len == 0")
    full_url = "http://" + os.environ['INGRESS_ISTIO_HOST'] + "/testcallseq?"
    full_url = full_url + self.util("1", "app-secrets-istio-appconfigv2-service-sm-2", uc)
    full_url = full_url + "&" + self.util("2", "app-secrets-istio-appconfigv2-service-sm-3", uc)
    full_url = full_url + "&call3=http://app-secrets-istio-appconfigv2-service-pubsub?"
    full_url = full_url + "gcpProjectID=" + os.environ['PUBSUB_GCP_PROJECT']
    full_url = full_url + "&topic=appconfigcrd-demo-topic1&message=hello2"

    headers = {"Host": "test-simple-hello.example.com"}
    response = RestHelper(full_url).get_text(None,headers)
    self.assertTrue(len(response) >  0, "response empty - len == 0")
    self.assertIn('PermissionDenied', response, "Failed Test - Publish")


  def test_simple_hello_uc7_outbound_ok(self):
    uc = "uc-secrets-vault-k8s"
    self.assertTrue(len(os.environ['INGRESS_NO_ISTIO_HOST']) >  0, "INGRESS_NO_ISTIO_HOST empty - len == 0")
    full_url = "http://" + os.environ['INGRESS_NO_ISTIO_HOST'] + "/testcallseq?"
    full_url = full_url + self.util("1", "app-secrets-vault-k8s-appconfigv2-service-sm-2", uc)
    full_url = full_url + "&" + self.util("2", "app-secrets-vault-k8s-appconfigv2-service-sm-1", uc)
    full_url = full_url + "&call3=https://httpbin.org/get"


    headers = {"Host": "test-simple-hello.example.com"}
    response = RestHelper(full_url).get_text(None,headers)
    self.assertTrue(len(response) >  0, "response empty - len == 0")
    self.assertIn('"User-Agent": "python-requests/2.22.0"', response, "Failed Test")
    self.assertIn('"Host": "httpbin.org"', response, "Failed Test")

  def test_simple_hello_uc7_pubsub_ok_topic1(self):
    uc = "uc-secrets-vault-k8s"
    self.assertIn('INGRESS_NO_ISTIO_HOST', os.environ, "INGRESS_NO_ISTIO_HOST environment variable not set")
    self.assertTrue(len(os.environ['INGRESS_NO_ISTIO_HOST']) >  0, "INGRESS_NO_ISTIO_HOST empty - len == 0")
    self.assertIn('PUBSUB_GCP_PROJECT', os.environ, "PUBSUB_GCP_PROJECT environment variable not set")
    self.assertTrue(len(os.environ['PUBSUB_GCP_PROJECT']) >  0, "PUBSUB_GCP_PROJECT empty - len == 0")
    full_url = "http://" + os.environ['INGRESS_NO_ISTIO_HOST'] + "/testcallseq?"
    full_url = full_url + self.util("1", "app-secrets-vault-k8s-appconfigv2-service-sm-2", uc)
    full_url = full_url + "&" + self.util("2", "app-secrets-vault-k8s-appconfigv2-service-sm-1", uc)
    full_url = full_url + "&call3=http://app-secrets-vault-k8s-appconfigv2-service-pubsub?"
    full_url = full_url + "gcpProjectID=" + os.environ['PUBSUB_GCP_PROJECT']
    full_url = full_url + "&topic=appconfigcrd-demo-topic1&message=hello1"



    headers = {"Host": "test-simple-hello.example.com"}
    response = RestHelper(full_url).get_text(None,headers)
    self.assertTrue(len(response) >  0, "response empty - len == 0")
    self.assertIn('Publish Success:', response, "Failed Test - Publish")

  def test_simple_hello_uc7_pubsub_ok_topic2(self):
    uc = "uc-secrets-vault-k8s"
    self.assertIn('INGRESS_NO_ISTIO_HOST', os.environ, "INGRESS_NO_ISTIO_HOST environment variable not set")
    self.assertTrue(len(os.environ['INGRESS_NO_ISTIO_HOST']) >  0, "INGRESS_NO_ISTIO_HOST empty - len == 0")
    self.assertIn('PUBSUB_GCP_PROJECT', os.environ, "PUBSUB_GCP_PROJECT environment variable not set")
    self.assertTrue(len(os.environ['PUBSUB_GCP_PROJECT']) >  0, "PUBSUB_GCP_PROJECT empty - len == 0")
    full_url = "http://" + os.environ['INGRESS_NO_ISTIO_HOST'] + "/testcallseq?"
    full_url = full_url + self.util("1", "app-secrets-vault-k8s-appconfigv2-service-sm-2", uc)
    full_url = full_url + "&" + self.util("2", "app-secrets-vault-k8s-appconfigv2-service-sm-1", uc)
    full_url = full_url + "&call3=http://app-secrets-vault-k8s-appconfigv2-service-pubsub?"
    full_url = full_url + "gcpProjectID=" + os.environ['PUBSUB_GCP_PROJECT']
    full_url = full_url + "&topic=appconfigcrd-demo-topic2&message=hello1"


    headers = {"Host": "test-simple-hello.example.com"}
    response = RestHelper(full_url).get_text(None,headers)
    self.assertTrue(len(response) >  0, "response empty - len == 0")
    self.assertIn('Publish Success:', response, "Failed Test - Publish")

  def test_workload_identity_pubsub_ok(self):
    # Assert that the pod with Google Cloud Workload Identity configured
    # can access a pubsub topic that it would otherwise not be able to.
    retries = 5
    while retries > 0:
      try:
        self.workload_identity_pubsub_ok()
        time.sleep(5)
        break
      except:
        retries -= 1
    if retries == 0:
      self.workload_identity_pubsub_ok()

  def workload_identity_pubsub_ok(self):
    uc = "uc-workload-identity"
    self.assertIn('INGRESS_NO_ISTIO_HOST', os.environ, "INGRESS_NO_ISTIO_HOST environment variable not set")
    self.assertTrue(len(os.environ['INGRESS_NO_ISTIO_HOST']) >  0, "INGRESS_NO_ISTIO_HOST empty - len == 0")
    self.assertIn('PUBSUB_GCP_PROJECT', os.environ, "PUBSUB_GCP_PROJECT environment variable not set")
    self.assertTrue(len(os.environ['PUBSUB_GCP_PROJECT']) >  0, "PUBSUB_GCP_PROJECT empty - len == 0")
    full_url = "http://" + os.environ['INGRESS_NO_ISTIO_HOST'] + "/testcallseq?"
    full_url = full_url + "call1=http://workload-identity-pubsub-app." + uc + ":8000"
    full_url = full_url + "?gcpProjectID=" + os.environ['PUBSUB_GCP_PROJECT']
    full_url = full_url + "&topic=workload-identity-topic&message=hello"

    headers = {"Host": "test-simple-hello.example.com"}
    response = RestHelper(full_url).get_text(None,headers)
    self.assertTrue(len(response) >  0, "response empty - len == 0")
    self.assertIn('Publish Success:', response, "Failed Test - Publish")

# if __name__ == '__main__':
#   h = HtmlTestRunner.HTMLTestRunner(combine_reports=True, report_name="MyReport", add_timestamp=False).run(suite)
