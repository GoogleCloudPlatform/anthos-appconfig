
import os
import sys
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

  def util(self, parm_call, parm_svc, parm_ns):
    full_url = "call" + parm_call + "=http://" + parm_svc
    if parm_ns:
      full_url = full_url + "." + parm_ns
    full_url += "/testcallseq"
    return full_url

  def test_simple_hello_uc1_outbound_ok(self):
    uc = "uc-allowed-services-k8s"
    self.assertTrue(len(os.environ['INGRESS_NO_ISTIO_HOST']) >  0, "INGRESS_ISTIO_HOST empty - len == 0")
    full_url = "http://" + os.environ['INGRESS_NO_ISTIO_HOST'] + "/testcallseq?"
    full_url = full_url + self.util("1", "app-allowed-k8s-appconfigv2-service-sm-2", uc)
    full_url = full_url + "&" + self.util("2", "app-allowed-k8s-appconfigv2-service-sm-1", uc)
    full_url = full_url + "&call3=https://httpbin.org/get"


    headers = {"Host": "test-simple-hello.example.com"}
    response = RestHelper(full_url).get_text(None,headers)
    self.assertTrue(len(response) >  0, "response empty - len == 0")
    self.assertIn('"User-Agent": "python-requests/2.22.0"', response, "Failed Test")
    self.assertIn('"Host": "httpbin.org"', response, "Failed Test")
# if __name__ == '__main__':
#   h = HtmlTestRunner.HTMLTestRunner(combine_reports=True, report_name="MyReport", add_timestamp=False).run(suite)
