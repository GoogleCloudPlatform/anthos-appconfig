FROM python:3.7
COPY . /app
WORKDIR /app
RUN \
  apt-get update && \
  apt-get install unzip wget -y && \
  rm -rf /var/lib/apt/lists/*

RUN mkdir -p /app/utils/down && wget -O "/app/utils/down/chromedriver.zip" \
 "https://chromedriver.storage.googleapis.com/76.0.3809.68/chromedriver_linux64.zip"

RUN unzip '/app/utils/down/chromedriver.zip' -d "/app/utils/"

RUN curl -LO https://storage.googleapis.com/kubernetes-release/release/v1.15.0/bin/linux/amd64/kubectl
RUN chmod +x ./kubectl
RUN mv ./kubectl /usr/local/bin/kubectl

RUN pip install -r "requirements.txt"

CMD [python3]
