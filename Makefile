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
	CGO_ENABLED=0 go build  -o bin/terraform-provider-proxmox_v2.0.0 cmd/terraform-provider-proxmox/* 
	@echo "Built terraform-provider-proxmox"

acctest:
	TF_ACC=1 go test ./proxmox

install: build 
	cp bin/terraform-provider-proxmox_v2.0.0 $$GOPATH/bin/terraform-provider-proxmox

clean:
	@git clean -f -d -X
