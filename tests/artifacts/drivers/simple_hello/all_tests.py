import os
import sys

sys.path.append(os.path.abspath('..'))

from unittest import TestLoader, TestSuite
import HtmlTestRunner
from test.test_simple_hello import SimpleHelloTestCase

the_testing_list = []

the_testing_list.append(TestLoader().loadTestsFromTestCase(SimpleHelloTestCase))


suite = TestSuite(the_testing_list)

runner = HtmlTestRunner.HTMLTestRunner(combine_reports=True, output="reports/temp", report_name="all_tests", add_timestamp=False)

runner.run(suite)