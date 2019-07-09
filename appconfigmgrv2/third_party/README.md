Copyright 2019 Google LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
# Third Party Files

## Istio

Where CRDs came from:
```sh
curl -L https://git.io/getLatestIstio | ISTIO_VERSION=1.1.7 sh -
./istio-$ISTIO_VERSION/install/kubernetes/helm/istio-init/files/* ./third_party/istio/v$ISTIO_VERSION/original-crds
rm -rf ./istio-$ISTIO_VERSION

# Pull any used CRDs from original-crds/ and place in seperate files in crds/
# because test harness cannot handle multiple documents defined in a single
# .yaml file.
```
