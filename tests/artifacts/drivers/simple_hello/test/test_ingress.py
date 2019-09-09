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
import unittest
import time
import requests

# sys.path.append(os.path.abspath('../simple_hello'))
# from auth_helper import GCPAuthHelper

from kubernetes import client, config
from pprint import pprint

config.load_kube_config()

class IngressTestCase(unittest.TestCase):

  def test_k8s_ingress(self):
    retries = 20
    while retries > 0:
      try:
        self.call_k8s_ingress()
        break
        time.sleep(15)
      except:
        retries -= 1
    if retries == 0:
      self.call_k8s_ingress()

  def call_k8s_ingress(self):
    exts = client.ExtensionsV1beta1Api()
    ig = exts.read_namespaced_ingress("ingress-k8s", "uc-ingress-k8s")
    ip = ig.status.load_balancer.ingress[0].ip
    r = requests.get(url="http://"+ip+"/get", headers={'Host':'my-httpbin-host'})
    self.assertEqual(r.status_code, 200)

if __name__ == '__main__':
  unittest.main()

