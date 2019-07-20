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
# Copyright 2018 Google LLC. All rights reserved. Licensed under the Apache License, Version 2.0 (the “License”);
#  you may not use this file except in compliance with the License. You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an “AS IS” BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions
# and limitations under the License.
#
# Any software provided by Google hereunder is distributed “AS IS”, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, and is not intended for production use.
#
# using information from:
#   https://github.com/istio/istio/tree/master/security/tools/jwt

import google.auth
from google.auth import jwt
from google.auth.transport.requests import Request
from google.oauth2 import service_account

import json
import requests
import os



OAUTH_TOKEN_URI = 'https://www.googleapis.com/oauth2/v4/token'
SA_SCOPES =  ['https://www.googleapis.com/auth/cloud-platform']

class GCPAuthHelper(object):

  def __get_key_from_file(self, absolute_creds_path=None):
    if not absolute_creds_path :
      if 'GOOGLE_APPLICATION_CREDENTIALS' in os.environ:
        absolute_creds_path = os.environ.get('GOOGLE_APPLICATION_CREDENTIALS')

    if absolute_creds_path :

      with open(absolute_creds_path, "r") as f:
        data = f.read()
        udata = json.loads(data)
        return udata

    raise Exception("service account info file required")

  def __init__(self, service_account_absolute_path=None):
    self.__service_account_info =  self.__get_key_from_file(service_account_absolute_path)
    # print( self.__service_account_info)
    self.__credentials_jwt = jwt.Credentials.from_service_account_info(
      self.__service_account_info,
      audience=OAUTH_TOKEN_URI)
    self.__credentials = service_account.Credentials.from_service_account_info(
      self.__service_account_info,
      scopes=SA_SCOPES
    )
    self.__credentials.refresh(Request())

  def get_email(self):
    return self.__credentials.signer_email

  def get_google_open_id_connect_token(self, url):
    """Get an OpenID Connect token issued by Google for the service account.

    This function:

      1. Generates a JWT signed with the service account's private key
         containing a special "target_audience" claim.

      2. Sends it to the OAUTH_TOKEN_URI endpoint. Because the JWT in #1
         has a target_audience claim, that endpoint will respond with
         an OpenID Connect token for the service account -- in other words,
         a JWT signed by *Google*. The aud claim in this JWT will be
         set to the value from the target_audience claim in #1.

    For more information, see
    https://developers.google.com/identity/protocols/OAuth2ServiceAccount .
    The HTTP/REST example on that page describes the JWT structure and
    demonstrates how to call the token endpoint. (The example on that page
    shows how to get an OAuth2 access token; this code is using a
    modified version of it to get an OpenID Connect token.)
    """
    self.__service_account_credentials = google.oauth2.service_account.Credentials(
      self.__credentials_jwt.signer,
      self.__credentials_jwt.signer_email,
      token_uri=OAUTH_TOKEN_URI, #additional_claims={}
      additional_claims={
        'target_audience': url,
        # 'target_audience': "https://w608e489101319183-tp.appspot.com"
      }
    )

    service_account_jwt = (
      self.__service_account_credentials._make_authorization_grant_assertion())

    request = google.auth.transport.requests.Request()

    body = {
      'assertion': service_account_jwt,
      'grant_type': google.oauth2._client._JWT_GRANT_TYPE,
    }

    token_response = google.oauth2._client._token_endpoint_request(
      request, OAUTH_TOKEN_URI, body)

    return token_response['id_token']

