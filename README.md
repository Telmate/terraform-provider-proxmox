[![Build Status](https://travis-ci.com/Telmate/terraform-provider-proxmox.svg?branch=master)](https://travis-ci.com/Telmate/terraform-provider-proxmox)

# Terraform provider plugin for Proxmox

This repository provides a Terraform provider for
the [Proxmox virtualization platform](https://pve.proxmox.com/pve-docs/) and exposes Terraform resources to provision
QEMU VMs and LXC Containers.

## Getting Started

In order to get started, use [the documentation included in this repository](docs/index.md). The documentation contains
a list of the options for the provider. Moreover, there are some guides available how to combine options and start
specific VMs.

## Quick Start

Follow this [install guide](docs/guides/installation.md) to install the plugin.

## Known Limitations

* `proxmox_vm_qemu`.`disk`.`size` attribute does not match what is displayed in the Proxmox UI.
* Updates to `proxmox_vm_qemu` resources almost always result as a failed task within the Proxmox UI. This appears to be
  harmless and the desired configuration changes do get applied.
* When using the `proxmox_lxc` resource, the provider will crash unless `rootfs` is defined.
* When using the Network Boot mode (PXE), a valid NIC must be defined for the VM, and the boot order must specify network first.

## Contributing

When contributing, please also add documentation to help other users.

### Debugging the provider

Debugging is available for this provider through the Terraform Plugin SDK versions 2.0.0. Therefore, the plugin can be
started with the debugging flag `--debug`.

For example (using [delve](https://github.com/go-delve/delve) as Debugger):

```bash
dlv exec --headless ./terraform-provider-my-provider -- --debug
```

For more information about debugging a provider please
see: [Debugger-Based Debugging](https://www.terraform.io/docs/extend/debugging.html#debugger-based-debugging)

## Useful links

* [Proxmox](https://www.proxmox.com/en/)
* [Proxmox documentation](https://pve.proxmox.com/pve-docs/)
* [Terraform](https://www.terraform.io/)
* [Terraform documentation](https://www.terraform.io/docs/index.html)
* [Recommended ISO builder](https://github.com/Telmate/terraform-ubuntu-proxmox-iso)
