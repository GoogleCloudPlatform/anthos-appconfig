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
```bash
python3 -m venv venv
source venv/bin/activate
python3 -m pip install -r requirements.txt

```

```bash
gcloud builds submit \
  --config=examples/hello-app-sm-py/build/cloudbuild.yaml  \
 examples/hello-app-sm-py \
  --project anthos-crd-v1-dev-t2 --substitutions="_APP_VERSION=v3.0.20"
```

https://chromedriver.storage.googleapis.com/76.0.3809.68/chromedriver_linux64.zip