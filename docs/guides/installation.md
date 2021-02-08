
# Automatic Registry Installation

To install this provider, copy and paste this code into your Terraform configuration (include a version tag). 
```hcl
terraform {
  required_providers {
    proxmox = {
      source = "Telmate/proxmox"
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
terraform init
```


# Manual Build & Install

## How to get terraform to recognize third party provider

Third-party plugins  can be manually installed into the user plugins directory,
located at `%APPDATA%\terraform.d\plugins` on Windows and `~/.terraform.d/plugins` on other systems. Plugins come
with executables that have to be placed in the plugin directory.

## Compile the executables with Go

In order to build the required executables, [install Go](https://golang.org/doc/install) first. Then clone this
repository and run the following commands inside the cloned repository.

```shell
export GO111MODULE=on go install github.com/Telmate/terraform-provider-proxmox/cmd/terraform-provider-proxmox
```

Then create the executables. They are placed in the `bin` folder inside the repository.

```shell
make
```

## Copy executables to plugin directory (Terraform <0.13)

You need to copy these executables to the ~/.terraform.d directory which will also need to have a plugins directory
created.

```shell
mkdir ~/.terraform.d/plugins
cp bin/terraform-provider-proxmox_v2.0.0 ~/.terraform.d/plugins
```

## Copy executables to plugin directory (Terraform >=0.13)

As of Terraform v0.13, locally-installed, third-party plugins must [conform to a new filesystem layout](https://github.com/hashicorp/terraform/blob/guide-v0.13-beta/draft-upgrade-guide.md#new-filesystem-layout-for-local-copies-of-providers).

>Terraform assumes that a provider without an explicit source address belongs to the "hashicorp" namespace on registry.terraform.io, which is not true for your in-house provider. Instead, you can use any domain name under your control to establish a virtual source registry to serve as a separate namespace for your local use.

Use the format: [host.domain]/telmate/proxmox/[version]/[arch].

In our case, we will use `registry.example.com` as our virtual source registry in the following examples.

```shell
# Uncomment for macOS
# PLUGIN_ARCH=darwin_amd64
PLUGIN_ARCH=linux_amd64

# Create the directory holding the newly built Terraform plugins
mkdir -p ~/.terraform.d/plugins/registry.example.com/telmate/proxmox/1.0.0/$PLUGIN_ARCH
```
Then, copy the executables to the directory you just created.

```shell
cp bin/terraform-provider-proxmox ~/.terraform.d/plugins/registry.example.com/telmate/proxmox/1.0.0/$PLUGIN_ARCH/
```

Add the source to `main.tf` `required_providers` section like so:

```
terraform {
  required_providers {
    proxmox = {
      source  = "registry.example.com/telmate/proxmox"
      version = ">=1.0.0"
    }
  }
  required_version = ">= 0.13"
}
```

## Initialize Terraform

Initialize Terraform so that it installs the new plugins:

```
terraform init
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
