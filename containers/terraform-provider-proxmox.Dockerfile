FROM docker.io/golang:1.21.6

ARG COMMAND=build

CMD mkdir /app
COPY .. /app
WORKDIR /app
CMD mkdir -p bin
RUN make $COMMAND

FROM docker.io/hashicorp/terraform:1.6

ARG VERSION

CMD mkdir -p /root/.terraform.d/plugins/registry.terraform.io/telmate/proxmox/$VERSION/linux_amd64
COPY --from=0 /app/bin/terraform-provider-proxmox /root/.terraform.d/plugins/registry.terraform.io/telmate/proxmox/$VERSION/linux_amd64/terraform-provider-proxmox

ENTRYPOINT ["terraform"]