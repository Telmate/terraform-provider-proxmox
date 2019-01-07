
.PHONY:  build clean

all: build

build: clean
	@echo " -> Building"
	@cd cmd/terraform-provider-proxmox && go build
	@echo "Built terraform-provider-proxmox"
	@cd cmd/terraform-provisioner-proxmox && go build
	@echo "Built terraform-provisioner-proxmox"

clean:
	@git clean -f -d -X
