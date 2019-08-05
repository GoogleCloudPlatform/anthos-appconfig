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
from os import path
import sys
import unittest
import subprocess

# sys.path.append(os.path.abspath('../simple_hello'))
# from auth_helper import GCPAuthHelper

from kubernetes import client, config
from pprint import pprint

config.load_kube_config()
core_v1 = client.CoreV1Api()
namespace = "uc-opa"

class OpaTestCase(unittest.TestCase):

  def test_pod_creation(self):
    should_exist = core_v1.list_namespaced_pod(namespace, label_selector="app=satisfies-labels")
    should_not_exist = core_v1.list_namespaced_pod(namespace, label_selector="app=missing-version-label-on-pods")
    self.assertEqual(len(should_exist.items), 3)
    self.assertEqual(len(should_not_exist.items), 0)

  def test_appconfig_ns_limit(self):
    # Should succeed.
    self.kubectl_apply("opa-appconfig-1.yaml")
    # Only one appconfig per namespaces should be allowed, so this should fail.
    with self.assertRaises(subprocess.CalledProcessError):
      self.kubectl_apply("opa-appconfig-2.yaml")

  def kubectl_apply(self, name):
    subprocess.check_call(["kubectl", "apply", "-f", path.join(path.dirname(__file__), "config", name)])

if __name__ == '__main__':
  unittest.main()

