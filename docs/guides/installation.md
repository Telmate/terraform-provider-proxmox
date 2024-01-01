# Automatic Registry Installation

To install this provider, copy and paste this code into your Terraform configuration (include a version tag).

```hcl
terraform {
  required_providers {
    proxmox = {
      source  = "telmate/proxmox"
      version = "<version tag>"
    }
  }
}

provider "proxmox" {
  # Configuration options
}
```

Then, run

```shell
$ terraform init
```

# Manual Build & Install

When developing this provider, it's useful to bootstrap a development as quick as possible. You can use
the [Proxmox VE vagrant VM](https://github.com/rgl/proxmox-ve) project for instance. Check out
the [examples](../../examples/vagrant_example.tf) for a `main.tf` to use.

## How to get terraform to recognize third party provider

Third-party plugins can be manually installed into the user plugins directory, located
at `%APPDATA%\terraform.d\plugins` on Windows and `~/.terraform.d/plugins` on other systems. Plugins come with
executables that have to be placed in the plugin directory.

## Compile the executables with Go

First, clone this repo and cd into the repo's root.

```shell
git clone https://github.com/Telmate/terraform-provider-proxmox
cd terraform-provider-proxmox
```

In order to build the required executables, [install Go](https://golang.org/doc/install) first. If
you want an automated way to do it, look at go.yml in the root of this repo.

Then to compile the provider:

```shell
make
```

The executable will be in the `./bin` directory.

## Copy executables to plugin directory (Terraform >=0.13)

As of Terraform v0.13, locally-installed, third-party plugins
must [conform to a new filesystem layout](https://github.com/hashicorp/terraform/blob/guide-v0.13-beta/draft-upgrade-guide.md#new-filesystem-layout-for-local-copies-of-providers)
.

> Terraform assumes that a provider without an explicit source address belongs to the "hashicorp" namespace on registry.terraform.io, which is not true for your in-house provider. Instead, you can use any domain name under your control to establish a virtual source registry to serve as a separate namespace for your local use.

Use the format: [host.domain]/telmate/proxmox/[version]/[arch].

In our case, we will use `registry.example.com` as our virtual source registry in the following examples.

```shell
# Uncomment for macOS
# PLUGIN_ARCH=darwin_amd64

$ PLUGIN_ARCH=linux_amd64

# Create the directory holding the newly built Terraform plugins
$ mkdir -p ~/.terraform.d/plugins/registry.example.com/telmate/proxmox/1.0.0/${PLUGIN_ARCH}
```

Then, copy the executables to the directory you just created. You could also use the `make local-dev-install` target.
it's important to note that you aren't required to use a semver, and if you don't, then the path must be altered
accordingly.

```shell
$ cp bin/terraform-provider-proxmox ~/.terraform.d/plugins/registry.example.com/telmate/proxmox/1.0.0/${PLUGIN_ARCH}/
$ ls -al ~/.terraform.d/plugins/registry.example.com/telmate/proxmox/1.0.0/${PLUGIN_ARCH}/
-rwxrwxr-x 1 user user 20352759 Feb 22 21:51 terraform-provider-proxmox_v1.0.0*
```

Add the source to your project's `main.tf` like so:

```
$ cat main.tf
terraform {
  required_providers {
    proxmox = {
      source  = "telmate/proxmox"
      version = ">=1.0.0"
    }
  }
  required_version = ">= 0.14"
}

[...]
```

## Copy executables to plugin directory (Terraform <0.13)

You need to copy these executables to the ~/.terraform.d directory which will also need to have a `plugins` directory
created.

```shell
mkdir -p ~/.terraform.d/plugins
cp -f bin/terraform-provider-proxmox ~/.terraform.d/plugins
```

## Initialize Terraform

Initialize Terraform so that it installs the new plugins:

```
$ terraform init
```

You should see the following marking the successful plugin installation:

```shell
[...]
Initializing provider plugins...
- Finding registry.example.com/telmate/proxmox versions matching ">= 1.0.0"...
- Installing registry.example.com/telmate/proxmox v1.0.0...
- Installed registry.example.com/telmate/proxmox v1.0.0 (unauthenticated)

Terraform has been successfully initialized!
[...]
```

Now that the plugin is installed, you can simply create a new terraform directory and do usual terraforming.
