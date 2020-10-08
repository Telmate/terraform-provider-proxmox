# How to get terraform to recognize third party provider

Third-party plugins (both providers and provisioners) can be manually installed into the user plugins directory,
located at `%APPDATA%\terraform.d\plugins` on Windows and `~/.terraform.d/plugins` on other systems. Plugins come
with executables that have to be placed in the plugin directory.

## Compile the executables with Go

In order to build the required executables, [install Go](https://golang.org/doc/install) first. Then clone this
repository and run the following commands inside the cloned repository. Since this plugin is both a provider and
provisioner in one, there are two install commands.

```
go install github.com/Telmate/terraform-provider-proxmox/cmd/terraform-provider-proxmox
go install github.com/Telmate/terraform-provider-proxmox/cmd/terraform-provisioner-proxmox
```

Then create the executables. They are placed in the `bin` folder inside the repository.

```
make
```

## Copy executables to plugin directory (Terraform <0.13)

You need to copy these executables to the ~/.terraform.d directory which will also need to have a plugins directory
created.

```shell
mkdir ~/.terraform.d/plugins
cp bin/terraform-provider-proxmox ~/.terraform.d/plugins
cp bin/terraform-provisioner-proxmox ~/.terraform.d/plugins
```

## Copy executables to plugin directory (Terraform >=0.13)

As of Terraform v0.13, locally-installed, third-party plugins must [conform to a new filesystem layout](https://github.com/hashicorp/terraform/blob/guide-v0.13-beta/draft-upgrade-guide.md#new-filesystem-layout-for-local-copies-of-providers).

>Terraform assumes that a provider without an explicit source address belongs to the "hashicorp" namespace on registry.terraform.io, which is not true for your in-house provider. Instead, you can use any domain name under your control to establish a virtual source registry to serve as a separate namespace for your local use.

Use the format: [host.domain]/telmate/proxmox/[version]/[arch].

Examples:  

macOS  
`mkdir -p ~/.terraform.d/plugins/terraform.example.com/telmate/proxmox/1.0.0/darwin_amd64/`

Linux   
``~/.terraform.d/plugins/terraform.example.com/telmate/proxmox/1.0.0/linux_amd64/``

Then, copy the executables to the directory you just created.

```shell
mkdir ~/.terraform.d/plugins/registry.example.com/
cp bin/terraform-provider-proxmox ~/.terraform.d/plugins/terraform.example.com/telmate/proxmox/1.0.0/darwin_amd64/
cp bin/terraform-provisioner-proxmox ~/.terraform.d/plugins/terraform.example.com/telmate/proxmox/1.0.0/darwin_amd64/
```

Add the source to required_providers like so:

```
terraform {
  required_providers {
    proxmox = {
      source  = "terraform.example.com/telmate/proxmox"
      version = ">=1.0.0"
    }
  }
  required_version = ">= 0.13"
}
```

## Initialize Terraform

Now the plugin is installed, you can simply create a new terraform directory and do usual terraforming.

```
terraform init
```
