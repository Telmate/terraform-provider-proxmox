FROM debian:stable-slim

ARG TERRAFORM_VERSION=0.13.5

RUN \
  apt-get update -y && \
  apt-get install unzip wget -y


RUN wget --progress=dot:mega https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_amd64.zip

RUN \
  unzip terraform_${TERRAFORM_VERSION}_linux_amd64.zip && \
  mv terraform /usr/local/bin/ && \
  chmod +x /usr/local/bin/terraform && \
  terraform --version
