.PHONY: build fmt vet test clean install acctest local-dev-install

all: build

fmt:
	@echo " -> checking code style"
	@! gofmt -d $(shell find . -path ./vendor -prune -o -name '*.go' -print) | grep '^'

vet:
	@echo " -> vetting code"
	@go vet ./...

test:
	@echo " -> testing code"
	@go test -v ./...

build: clean
	@echo " -> Building"
	mkdir -p bin
	CGO_ENABLED=0 go build  -o bin/terraform-provider-proxmox cmd/terraform-provider-proxmox/*
	@echo "Built terraform-provider-proxmox"

acctest: build
	# to run only certain tests, run something of the form:  make acctest TESTARGS='-run=TestAccProxmoxVmQemu_DiskSlot'
	TF_ACC=1 go test ./proxmox $(TESTARGS)

install: build
	cp bin/terraform-provider-proxmox $$GOPATH/bin/terraform-provider-proxmox

KERNEL=$(shell $(uname -s | tr '[:upper:]' '[:lower:]'))
ARCH=$(shell if [ "$$(uname -m)" == "x86_64" ]; then echo amd64; fi)
local-dev-install:
	mkdir -p ~/.terraform.d/plugins/registry.example.com/telmate/proxmox/$(KERNEL)_$(ARCH)/
	cp bin/terraform-provider-proxmox ~/.terraform.d/plugins/registry.example.com/telmate/proxmox/$(KERNEL)_$(ARCH)/

clean:
	@git clean -f -d -X
