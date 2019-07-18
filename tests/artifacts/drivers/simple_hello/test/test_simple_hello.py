
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

  def test_simple_hello_empty_token(self):
    o_auth_helper = GCPAuthHelper()

    token_gcf = o_auth_helper.get_google_open_id_connect_token("https://site2.ecom1.joecloudy.com")
    self.assertIn('INGRESS_ISTIO_HOST', os.environ, "INGRESS_ISTIO_HOST environment variable not set")
    self.assertTrue(len(os.environ['INGRESS_ISTIO_HOST']) >  0, "INGRESS_ISTIO_HOST empty - len == 0")
    full_url = self.get_url(os.environ['INGRESS_ISTIO_HOST'],
                            'app-allowed-jwt-istio-appconfigv2-service-sm-', 'uc-allowed-jwt-istio', [ '2', '3' ])
    headers = {"Host": "test-simple-hello.example.com"}
    response = RestHelper(full_url).get_text(token_gcf,headers)
    self.assertTrue(len(response) >  0, "response empty - len == 0")
    self.assertIn("Last Call Successful", response, "Failed Test")


# if __name__ == '__main__':
#   h = HtmlTestRunner.HTMLTestRunner(combine_reports=True, report_name="MyReport", add_timestamp=False).run(suite)
