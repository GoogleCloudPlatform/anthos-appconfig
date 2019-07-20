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

RUN pip install -r "requirements.txt"

CMD [python3]