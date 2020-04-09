.PHONY:  build  fmt vet test clean install

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
	CGO_ENABLED=0 go build -v -o bin/terraform-provisioner-proxmox cmd/terraform-provisioner-proxmox/* 
	@echo "Built terraform-provisioner-proxmox"

install: build 
	cp bin/terraform-provider-proxmox $$GOPATH/bin/terraform-provider-proxmox
	cp bin/terraform-provisioner-proxmox $$GOPATH/bin/terraform-provisioner-proxmox

clean:
	@git clean -f -d -X
