# How to get terraform to recognize third party provider

The default installation of terraform has no way of finding custom third party providers. This document shows you how to locate providers so that `terraform init` can find them and copy them to your working terraform directory when creating a new project.

## Locate the provider binaries

Assuming you have followed other directions the deafult locations for these binaries (on a mac at least) are as follows:

```bash
which terraform-provider-proxmox
~/go-workspace/bin/terraform-provider-proxmox
which terraform-provisioner-proxmox
~/go-workspace/bin/terraform-provisioner-proxmox 
```

## Copy provider binaries

You need to copy these binaries to the ~/.terraform.d directory which will also need to have a plugins directory created:

```shell
cd ~/.terraform.d
mkdir plugins
cd plugins
cp ~/go-workspace/bin/terraform-provider-proxmox .
cp ~/go-workspace/bin/terraform-provisioner-proxmox .
```
Once this is done, simply create a new terraform directory and do usual terraforming (terraform init etc)
