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

import httplib2
import logging
import json
import os
import urllib



import requests

class RestHelper(object):

  def __init__(self, url, timeout=30, is_public=True):
    self.__log = 1
    self.__debug = True
    self.__is_public = is_public
    self.__service_url = url
    self.__timeout = timeout

    http = httplib2.Http(timeout=timeout)

  def get_text(self, jwt=None, headers={}):
    print('get_text')
    if jwt:
      headers['Authorization'] = 'Bearer ' + jwt

    # print('headers', headers)
    result = requests.get(self.__service_url, timeout=self.__timeout, headers=headers)
    # logging.info('process_file_nlp:results-[{}]-[{}]'.format(result, len(result.text)))
    if result.status_code and str(result.status_code) in ["200"]:
      # print('get_test', result.text)
      return result.text
    raise Exception("Respose Failure for HTTP - {} - {}".format(result.status_code, result.text))

  @staticmethod
  def call_with_sequence(url_base, collection, jwt=None, headers={}, timeout=30):
    logging.info('call_with_sequence:start')
    get_params={}

    for index,c in enumerate(collection):
      get_params.update(c)

    logging.info('call_with_sequence:base[{}]:params[{}]'.format(url_base, get_params))
    if jwt:
      headers['Authorization'] = 'Bearer ' + jwt
    # print('headers', headers)
    result = requests.get(url_base, params=get_params, timeout=timeout, headers = headers)
    # logging.info('process_file_nlp:results-[{}]-[{}]'.format(result, len(result.text)))
    if result.status_code and str(result.status_code) in ["200"]:
      # print('get_test', result.text)
      return result.text
    logging.error('HTTP Error')
    raise Exception("Respose Failure for HTTP - {} - {}".format(result.status_code, result.text))

    return None

