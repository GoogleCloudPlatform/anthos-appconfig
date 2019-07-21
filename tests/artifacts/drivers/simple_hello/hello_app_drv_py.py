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
# 
# Copyright 2019 Google LLC. This software is provided as-is, 
# without warranty or representation for any use or purpose.#
#



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
# 
# Copyright 2019 Google LLC. This software is provided as-is, 
# without warranty or representation for any use or purpose.#
#



import argparse
import logging
import os
import traceback

from collections import OrderedDict
from flask import Flask, request

# from threading import Timer
from http_rest_helper import RestHelper

APP_LOG_PREFIX = '[PERF]'
APP_LOG_PREFIX_SUCCESS = APP_LOG_PREFIX + '[OK]'
APP_LOG_PREFIX_ERROR = APP_LOG_PREFIX + '[ERROR]'
APP_LOG_PREFIX_EXCEPTION = APP_LOG_PREFIX + '[EXCEPTION]'
APP_LOG_PREFIX_STATUS = APP_LOG_PREFIX + '[STATUS]'

K8S_HOST_NAME='?'
if 'HOSTNAME' in os.environ:
  K8S_HOST_NAME= os.environ['HOSTNAME']


app = Flask(__name__)

def get_collect(k, v, collection):
  if v:
    collection.append({k: v})
    return True
  return False

def get_next_call(d, collection):
  found_call = False
  kv_entry = None
  d_sorted = OrderedDict(sorted(d.items()))
  for k,v in d_sorted.items():
    if not kv_entry:
      if k.startswith('call'):
        print('get_next_call',k,v)
        kv_entry = v
    else:
      collection.append({k:v})
  return kv_entry


def get_headers_to_include(d, o_dict):
  for k,v in d.items():
    if k in ['Authorization']:
      o_dict[k] = v
  return o_dict


def manual_hello():
  collection = [
    {'call4': 'http://localhost:8080/testcallseq'},
    {'call3': 'http://httpbin.org/get'},
  ]

  text = RestHelper.call_with_sequence('http://localhost:8080/testcallseq', collection)

  print(APP_LOG_PREFIX_SUCCESS + ':Call:Result[{}]'.format(text))
  return text

@app.route('/hello', methods=['GET'])
def hello():
  """Return a friendly HTTP greeting."""
  collection = [
    {'call': 'http://httpbin.org/get'},
  ]
  next_call = collection.pop()['call']
  print('next_call-[{}]-[{}]'.format(next_call, collection))
  text = RestHelper.call_with_sequence(next_call, collection)

  print(APP_LOG_PREFIX_SUCCESS + ':Call:Result[{}]'.format(text))
  return text


@app.route('/testcallseq')
def testcallseq():
  """Return a friendly HTTP greeting."""
  result_prefix = APP_LOG_PREFIX_ERROR
  result_text = "Unexpected"
  try:
    print('testcallseq','start', request.url)
    collection = []
    headers_dict = {}
    next_call = get_next_call(request.args.to_dict(), collection)
    headers_dict = get_headers_to_include(request.headers, headers_dict)
    if next_call:
      try:
        print('next_call-[{}]-[{}]-[{}]'.format(next_call, collection, headers_dict))
        result_text = RestHelper.call_with_sequence(next_call, collection, headers=headers_dict)

        if not result_text.startswith('*Error*'):
          result_prefix = APP_LOG_PREFIX_SUCCESS
      except Exception as e:
        result_text = '*Error*-Happened - Making the request-url[{}]'.format(request.url)
    else:
      result_text="Last Call Successful"
  except Exception as e:
    result_text = '*Error* - Unexpected Error Happened - Probably Parsing'
    traceback.print_exc()

  print(result_prefix + ':Call:Result[{}]'.format(result_text))
  return "host:" + K8S_HOST_NAME + "\n" + result_text


@app.errorhandler(500)
def server_error(e):
  logging.exception('An error occurred during a request.')
  return """
    An internal error occurred: <pre>{}</pre>
    See logs for full stacktrace.
    """.format(e), 500


if __name__ == '__main__':

  parser = argparse.ArgumentParser()

  parser.add_argument('--debug')
  args = parser.parse_args()
  # RestHelper('http://httpbin.org/get').get_text()
  if args.debug:
    manual_hello()
  else:
    app.run(host='0.0.0.0', port=8080, debug=True)
