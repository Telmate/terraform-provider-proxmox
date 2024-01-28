ARG GO_VERSION=1.21.6
ARG TERRAFORM_VERSION=1.6
ARG COMMAND=build
ARG VERSION

FROM docker.io/golang:${GO_VERSION}

ARG COMMAND=build

CMD mkdir /app
COPY .. /app
WORKDIR /app
CMD mkdir -p bin
RUN make $COMMAND

FROM docker.io/hashicorp/terraform:${TERRAFORM_VERSION}

CMD mkdir -p /root/.terraform.d/plugins/registry.terraform.io/telmate/proxmox/$VERSION/linux_amd64
COPY --from=0 /app/bin/terraform-provider-proxmox /root/.terraform.d/plugins/registry.terraform.io/telmate/proxmox/$VERSION/linux_amd64/terraform-provider-proxmox

RUN apk add py3-pip
RUN apk add gcc musl-dev python3-dev libffi-dev openssl-dev cargo make

RUN python3 -m venv .venv
RUN .venv/bin/pip install --no-cache-dir -U pip setuptools azure-cli
ENV PATH="/app/.venv/bin:$PATH"

ENTRYPOINT ["terraform"]