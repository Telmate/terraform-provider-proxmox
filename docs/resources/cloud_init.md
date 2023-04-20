# Cloud Init Resource

This resource creates and manages a Proxmox Cloud Init disk.

## Example Usage

### Basic example

```hcl
resource "proxmox_cloud_init" "ci" {
  name      = var.name
  pve_node  = var.pve_node
  storage   = var.storage

  meta_data = <<EOF
instance-id: ${var.name}
local-hostname: ${var.name}
EOF

  user_data = <<EOF
#cloud-config
users:
  - name: foobar
    sudo: ALL=(ALL) NOPASSWD:ALL
EOF

  network_config = <<EOF
version: 2
ethernets:
  nic0:
    match:
      name: "en*"
    dhcp4: false
    dhcp6: false
    addresses: ["10.1.100.200/24"]
    gateway4: "10.1.100.1"
    nameservers:
      addresses: ["8.8.8.8", "1.1.1.1"]
EOF
}

resource "proxmox_vm_qemu" "my-vm" {
....
  // Define a disk block which will mount the generated cloud init ISO
  disk {
    type    = "ide"
    media   = "cdrom"
    storage = var.storage
    volume  = proxmox_cloud_init.ci.id
    size    = proxmox_cloud_init.ci.size
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
| `network_config` | `string` | `""`          | Content of the network-config file                                      |

## Attribute reference

In addition to the arguments listed above, the following computed attributes are exported:

- `id` - The volume identification of the ISO.
- `size` - The volume size
- `sha256` - The computed sha256 checksum
