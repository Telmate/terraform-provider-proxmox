[![Build Status](https://travis-ci.com/Telmate/terraform-provider-proxmox.svg?branch=master)](https://travis-ci.com/Telmate/terraform-provider-proxmox)

# Terraform provider plugin for Proxmox

This repository provides both a Terraform provider for the [Proxmox virtualization platform](https://pve.proxmox.com/pve-docs/).

## Getting started

In order to get started, use [the documentation included in this repository](docs/index.md). The documentation contains
a list of the options for the provider. Moreover, there are some guides available how to combine
options and start specific VMs.

## Known Limitations

This section is here to both serve as a reminder to contributers of areas for improvement, but also as a "head's up" to users so you don't have to run into it and then find it buried in some logged issue.

* `proxmox_vm_qemu`.`disk`.`size` attribute does not match what is displayed in the Proxmox UI.
* Updates to `proxmox_vm_qemu` resources almost always result as a failed task within the Proxmox UI. This appears to be harmless and the desired configuration changes do get applied.
* `proxmox_vm_qemu` does not (yet) validate vm names, be sure to only use alphanumeric and dashes otherwise you may get an opaque 400 Parameter Verification failed (indicating a bad value was sent to proxmox).

## Contributing

When contributing, please also add documentation to help other users.

## Useful links

* [Proxmox](https://www.proxmox.com/en/)
* [Proxmox documentation](https://pve.proxmox.com/pve-docs/)
* [Terraform](https://www.terraform.io/)
* [Terraform documentation](https://www.terraform.io/docs/index.html)
* [Recommended ISO builder](https://github.com/Telmate/terraform-ubuntu-proxmox-iso)
