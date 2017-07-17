FROM docker-registry.zoona.io/ops-engineering/docker-alpine-java8:e8fec096bd03dd932b58731b1878806a1ed8d459


RUN apk update && apk-install openssh-client && apk-install bash

RUN echo -e 'Host *\nUseRoaming no' >> /etc/ssh/ssh_config


ENV NOMAD_VERSION 0.5.6
ENV NOMAD_SHA256 3f5210f0bcddf04e2cc04b14a866df1614b71028863fe17bcdc8585488f8cb0c

ADD https://releases.hashicorp.com/nomad/${NOMAD_VERSION}/nomad_${NOMAD_VERSION}_linux_amd64.zip /tmp/nomad.zip
RUN echo "${NOMAD_SHA256}  /tmp/nomad.zip" > /tmp/nomad.sha256 \
  && sha256sum -c /tmp/nomad.sha256 \
  && cd /bin \
  && unzip /tmp/nomad.zip \
  && chmod +x /bin/nomad \
  && rm /tmp/nomad.zip


ENV ENVCONSUL_VERSION 0.6.2
ENV ENVCONSUL_SHA256 c86ecd5b1cac5b6d59326e495809ce4778ecca0bf2e41b2a650613b865f2565b

ADD https://releases.hashicorp.com/envconsul/${ENVCONSUL_VERSION}/envconsul_${ENVCONSUL_VERSION}_linux_amd64.zip /tmp/envconsul.zip
RUN echo "${ENVCONSUL_SHA256}  /tmp/envconsul.zip" > /tmp/envconsul.sha256 \
  && sha256sum -c /tmp/envconsul.sha256 \
  && cd /bin \
  && unzip /tmp/envconsul.zip \
  && chmod +x /bin/envconsul \
  && rm /tmp/envconsul.zip

RUN apk add --update bash parallel git jq curl && rm -rf /var/cache/apk/*

COPY nomad-eval-monitor /usr/local/bin
