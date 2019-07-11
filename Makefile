
.PHONY:  build clean install

all: build

setup:
	go get github.com/Telmate/proxmox-api-go
	go get github.com/hashicorp/terraform/plugin
	go get github.com/hashicorp/terraform/terraform
	go get github.com/Telmate/terraform-provider-proxmox/cmd/terraform-provider-proxmox
	go get github.com/Telmate/terraform-provider-proxmox/cmd/terraform-provisioner-proxmox

build: clean
	@echo " -> Building"
	@cd cmd/terraform-provider-proxmox && go build
	@echo "Built terraform-provider-proxmox"
	@cd cmd/terraform-provisioner-proxmox && go build
	@echo "Built terraform-provisioner-proxmox"


install: clean
	@echo " -> Installing"
	go install github.com/Telmate/terraform-provider-proxmox/cmd/terraform-provider-proxmox
	go install github.com/Telmate/terraform-provider-proxmox/cmd/terraform-provisioner-proxmox

clean:
	@git clean -f -d -X
