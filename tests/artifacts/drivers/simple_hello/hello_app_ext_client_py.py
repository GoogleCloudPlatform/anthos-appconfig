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

import sys


def main(argv):
    pass


if __name__ == '__main__':
    main(sys.argv)
# Copyright 2015 Google Inc. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

import argparse
import logging
import os
import traceback

from auth_helper import GCPAuthHelper

from http_rest_helper import RestHelper


def get_url(host, service_name, namespace_name, nested_calls):
  full_url = 'http://' + host + '/testcallseq?'
  for k,v in enumerate(nested_calls):
    full_url += 'call' + str(k+1) + '=http://' + service_name + str(v) + '.' + namespace_name + '/testcallseq&'

  return full_url

if __name__ == '__main__':

  parser = argparse.ArgumentParser()
  parser.add_argument('--debug')
  parser.add_argument('--host', required=True)
  parser.add_argument('--service_name', required=True)
  parser.add_argument('--namespace_name', required=True)
  parser.add_argument('--nested_calls', required=True)
  parser.add_argument('--skip_jwt')
  args = parser.parse_args()
  port = '80'

  host = args.host
  service_name = args.service_name
  namespace_name = args.namespace_name
  nested_calls = args.nested_calls
  skip_jwt = args.skip_jwt

  if 'INGRESS_ISTIO_HOST' in os.environ and os.environ['INGRESS_ISTIO_HOST']:
    o_auth_helper = GCPAuthHelper()
    token_gcf = o_auth_helper.get_google_open_id_connect_token("https://site2.ecom1.joecloudy.com")
    if skip_jwt:
      print('Empty JWT')
      token_gcf = None
    # token_gcf = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c'
    headers = { 'Host': host}
    full_url = get_url(os.environ['INGRESS_ISTIO_HOST'], service_name, namespace_name, nested_calls.split(','))
    print('full_url', full_url, 'host', host)

    print(RestHelper(full_url).get_text(token_gcf,headers))
  else:
    print('Need to export INGRESS_ISTIO_HOST IP')