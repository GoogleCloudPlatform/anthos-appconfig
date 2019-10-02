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

sys.path.append(os.path.abspath('..'))

from unittest import TestLoader, TestSuite
import HtmlTestRunner
from test.test_simple_hello import SimpleHelloTestCase
from test.test_opa import OpaTestCase
from test.test_ingress import IngressTestCase

the_testing_list = []

the_testing_list.append(TestLoader().loadTestsFromTestCase(SimpleHelloTestCase))
the_testing_list.append(TestLoader().loadTestsFromTestCase(OpaTestCase))
the_testing_list.append(TestLoader().loadTestsFromTestCase(IngressTestCase))


suite = TestSuite(the_testing_list)

runner = HtmlTestRunner.HTMLTestRunner(combine_reports=True, output="reports/temp", report_name="all_tests", add_timestamp=False)

results = runner.run(suite)

print('results', results)


if len(results.errors) > 0 or len(results.failures) > 0:
  sys.exit(1)
