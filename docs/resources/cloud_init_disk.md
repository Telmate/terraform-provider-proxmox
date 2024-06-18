This documentation is about how to manually create a cloud-init CD-ROM image.
This is only for unusual situations where the standard cloud-init support is
not enough. For a simple cloud-init example, see
[examples/cloudinit_example.tf](../../examples/cloudinit_example.tf).

# Cloud Init Disk Resource

This resource creates and manages a Proxmox Cloud Init disk.

## Example Usage

```hcl
locals {
  vm_name          = "awesome-vm"
  pve_node         = "pve-node-1"
  iso_storage_pool = "cephfs"
}

resource "proxmox_cloud_init_disk" "ci" {
  name      = local.vm_name
  pve_node  = local.pve_node
  storage   = local.iso_storage_pool

  meta_data = yamlencode({
    instance_id    = sha1(local.vm_name)
    local-hostname = local.vm_name
  })

  user_data = <<-EOT
  #cloud-config
  users:
    - default
  ssh_authorized_keys:
    - ssh-rsa AAAAB3N......
  EOT

  network_config = yamlencode({
    version = 1
    config = [{
      type = "physical"
      name = "eth0"
      subnets = [{
        type            = "static"
        address         = "192.168.1.100/24"
        gateway         = "192.168.1.1"
        dns_nameservers = [
          "1.1.1.1", 
          "8.8.8.8"
          ]
      }]
    }]
  })
}

resource "proxmox_vm_qemu" "vm" {
...
  // Define a disk block with media type cdrom which reference the generated cloud-init disk
  disks {
    scsi {
      scsi0 {
        cdrom {
          iso = "${proxmox_cloud_init_disk.ci.id}"
        }
      }
    }
  }
...
}
```

## Argument reference

### Top Level Block

The following arguments are supported in the top level resource block.

| Argument         | Type     | Default Value | Description                                                             |
| ---------------- | -------- | ------------- | ----------------------------------------------------------------------- |
| `name`           | `string` |               | **Required** The name of the Cloud Init disk.                           |
| `pve_node`       | `string` |               | **Required** The name of the Proxmox Node on which to place the ISO.    |
| `storage`        | `string` |               | **Required** The name of the Proxmox Storage on which to place the ISO. |
| `meta_data`      | `string` | `""`          | Content of the meta-data file                                           |
| `user_data`      | `string` | `""`          | Content of the user-data file                                           |
| `vendor_data`    | `string` | `""`          | Content of the vendor-data file                                         |
| `network_config` | `string` | `""`          | Content of the network-config file                                      |

## Attribute reference

In addition to the arguments listed above, the following computed attributes are exported:

- `id` - The volume identification of the ISO.
- `size` - The volume size
- `sha256` - The computed sha256 checksum
