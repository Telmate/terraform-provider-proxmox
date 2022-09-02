SHELL := /bin/bash

DESCRIBE           := $(shell git describe --match "v*" --always --tags)
DESCRIBE_PARTS     := $(subst -, ,$(DESCRIBE))

VERSION_TAG        := $(word 1,$(DESCRIBE_PARTS))
COMMITS_SINCE_TAG  := $(word 2,$(DESCRIBE_PARTS))

VERSION            := $(subst v,,$(VERSION_TAG))
VERSION_PARTS      := $(subst ., ,$(VERSION))

MAJOR              := $(word 1,$(VERSION_PARTS))
MINOR              := $(word 2,$(VERSION_PARTS))
MICRO              := $(word 3,$(VERSION_PARTS))

NEXT_MAJOR         := $(shell echo $$(($(MAJOR)+1)))
NEXT_MINOR         := $(shell echo $$(($(MINOR)+1)))
NEXT_MICRO          = $(shell echo $$(($(MICRO)+1)))

ifeq ($(strip $(COMMITS_SINCE_TAG)),)
CURRENT_VERSION_MICRO := $(MAJOR).$(MINOR).$(MICRO)
CURRENT_VERSION_MINOR := $(CURRENT_VERSION_MICRO)
CURRENT_VERSION_MAJOR := $(CURRENT_VERSION_MICRO)
else
CURRENT_VERSION_MICRO := $(MAJOR).$(MINOR).$(NEXT_MICRO)
CURRENT_VERSION_MINOR := $(MAJOR).$(NEXT_MINOR).0
CURRENT_VERSION_MAJOR := $(NEXT_MAJOR).0.0
endif

DATE                = $(shell date +'%d.%m.%Y')
TIME                = $(shell date +'%H:%M:%S')
COMMIT             := $(shell git rev-parse HEAD)
AUTHOR             := $(firstword $(subst @, ,$(shell git show --format="%aE" $(COMMIT))))
BRANCH_NAME        := $(shell git rev-parse --abbrev-ref HEAD)

TAG_MESSAGE         = "$(TIME) $(DATE) $(AUTHOR) $(BRANCH_NAME)"
COMMIT_MESSAGE     := $(shell git log --format=%B -n 1 $(COMMIT))

CURRENT_TAG_MICRO  := "v$(CURRENT_VERSION_MICRO)"
CURRENT_TAG_MINOR  := "v$(CURRENT_VERSION_MINOR)"
CURRENT_TAG_MAJOR  := "v$(CURRENT_VERSION_MAJOR)"

# Determine KERNEL and ARCH
GOOS := $(shell go tool dist env | grep 'GOOS' | sed -e 's|.*"\(.*\)"|\1|')
GOARCH := $(shell go tool dist env | grep 'GOARCH' | sed -e 's|.*"\(.*\)"|\1|')

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
	CGO_ENABLED=0 go build -trimpath -o bin/terraform-provider-proxmox
	@echo "Built terraform-provider-proxmox"

acctest: build
	# to run only certain tests, run something of the form:  make acctest TESTARGS='-run=TestAccProxmoxVmQemu_DiskSlot'
	TF_ACC=1 go test ./proxmox $(TESTARGS)

install: build
	cp bin/terraform-provider-proxmox $$GOPATH/bin/terraform-provider-proxmox

local-dev-install: build
	@echo "$(CURRENT_VERSION_MICRO)"
	@echo "GOOS=$(GOOS)"
	@echo "GOARCH=$(GOARCH)"
	mkdir -p ~/.terraform.d/plugins/localhost/telmate/proxmox/$(MAJOR).$(MINOR).$(NEXT_MICRO)/$(GOOS)_$(GOARCH)/
	cp bin/terraform-provider-proxmox ~/.terraform.d/plugins/localhost/telmate/proxmox/$(MAJOR).$(MINOR).$(NEXT_MICRO)/$(GOOS)_$(GOARCH)/

clean:
	@git clean -f -d
