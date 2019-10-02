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
import time

from kubernetes import client, config
from pprint import pprint

config.load_kube_config()
core_v1 = client.CoreV1Api()
namespace = "uc-opa"

class OpaTestCase(unittest.TestCase):

  def test_pod_creation(self):
    # Assert that gatekeeper has blocked the creation of pods
    # that are missing required labels.
    should_exist = core_v1.list_namespaced_pod(namespace, label_selector="app=satisfies-labels")
    should_not_exist = core_v1.list_namespaced_pod(namespace, label_selector="app=missing-version-label-on-pods")
    self.assertEqual(len(should_exist.items), 3)
    self.assertEqual(len(should_not_exist.items), 0)

  def test_appconfig_ns_limit(self):
    # Assert that no more than one app config can be created in a single
    # namespace.
    time.sleep(300)

    # Should succeed.
    self.kubectl_apply("opa-appconfig-1.yaml")

    time.sleep(300)
    # Only one appconfig per namespaces should be allowed, so this should fail.
    with self.assertRaises(subprocess.CalledProcessError):
      self.kubectl_apply("opa-appconfig-2.yaml")

  def kubectl_apply(self, name):
    subprocess.check_call(["kubectl", "apply", "-f", path.join(path.dirname(__file__), "config", name)])

if __name__ == '__main__':
  unittest.main()

