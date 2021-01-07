# VM Qemu Resource

Resources are the most important element in the Terraform language. Each resource block describes one or more 
infrastructure objects, such as virtual networks, compute instances, or higher-level components such as DNS records.

This resource manages a Proxmox VM Qemu container.

## Create a Qemu VM resource

You can start from either an ISO or clone an existing VM. Optimally, you could create a VM resource you will use a clone 
base with an ISO, and make the rest of the VM resources depend on that base "template" and clone it.

When creating a VM Qemu resource, you create a `proxmox_vm_qemu` resource block. The name and target node of the VM are
the only required parameters.

```hcl
resource "proxmox_vm_qemu" "resource-name" {
    name = "VM name"
    target_node = "Node to create the VM on"
}
```

## Preprovision

With preprovision you can provision a VM directly from the resource block. This provisioning method is therefore ran
**before** provision blocks. When using preprovision, there are three `os_type` options: `ubuntu`, `centos` or `cloud-init`.

```hcl
resource "proxmox_vm_qemu" "prepprovision-test" {
    ...
    preprovision = true
    os_type = "ubuntu"
}
```

### Preprovision for Linux (Ubuntu / CentOS)

There is a pre-provision phase which is used to set a hostname, intialize eth0, and resize the VM disk to available 
space. This is done over SSH with the `ssh_forward_ip`, `ssh_user` and `ssh_private_key`. Disk resize is done if the file 
[/etc/auto_resize_vda.sh](https://github.com/Telmate/terraform-ubuntu-proxmox-iso/blob/master/auto_resize_vda.sh) exists.

```hcl
resource "proxmox_vm_qemu" "prepprovision-test" {
    ...
    preprovision = true
    os_type = "ubuntu"
    ssh_forward_ip = "10.0.0.1"
    ssh_user = "terraform"
    ssh_private_key = <<EOF
-----BEGIN RSA PRIVATE KEY-----
private ssh key terraform
-----END RSA PRIVATE KEY-----
EOF
    os_network_config =  <<EOF
auto eth0
iface eth0 inet dhcp
EOF
    
    connection {
        type = "ssh"
        user = "${self.ssh_user}"
        private_key = "${self.ssh_private_key}"
        host = "${self.ssh_host}"
        port = "${self.ssh_port}"
    }
}
```


## Preprovision for Cloud-Init

Cloud-init VMs must be cloned from a [cloud-init ready template](https://pve.proxmox.com/wiki/Cloud-Init_Support). When
creating a resource that is using Cloud-Init, there are multi configurations possible. You can use either the `ciconfig`
parameter to create based on [a Cloud-init configuration file](https://cloudinit.readthedocs.io/en/latest/topics/examples.html)
or use the Proxmox variable `ciuser`, `cipassword`, `ipconfig0`, `ipconfig1`, `searchdomain`, `nameserver` and `sshkeys`.

For more information, see the [Cloud-init guide](docs/guides/cloud_init.md).

## Argument reference

The following arguments are supported in the resource block.

|Argument|Type|Required?|Default Value|Description|
|--------|----|---------|-------------|-----------|
|`name`|`string`|**Yes**||The name of the VM.|
|`target_node`|`string`|**Yes**||The name of the Proxmox Node on which to place the VM.|
|`desc`|`string`|**Yes**|``||

|`vmid`|`integer`|No|`0`|The ID of the VM in Proxmox. The default value of `0` indicates it should use the next available ID in the sequence.|
|`desc`|`string`|No|`""`|The description of the VM. Shows as the 'Notes' field in the Proxmox GUI.|
|`define_connection_info`|`bool`|No|`true`|Define the (SSH) connection parameters for preprovisioners, see config block below.|
|`bios`|`string`|No|`"seabios"`|The BIOS to use, options are `seabios` or `ovmf` for UEFI.|
|`onboot`|`bool`|No|`true`|Whether to have the VM startup after the PVE node starts.|
|`boot`|`string`|No|`"cdn"`|The boot order for the VM. Ordered string of characters representing: floppy (a), hard disk (c), CD-ROM (d), or network (n).|
|`bootdisk`|`string`|No|*Computed*|Enable booting from specified disk. This value is computed by terraform, so you shouldn't need to change it under most circumstances.|
|`agent`|`integer`|No|`0`|Whether to enable the QEMU Guest Agent. Note, you must still install the [`qemu-guest-agent`](https://pve.proxmox.com/wiki/Qemu-guest-agent) daemon in the quest for this to have any effect.|
|`iso`|`string`|No|`""`|The name of the ISO image to mount to the VM. Only applies when `clone` is not set.|
|`clone`|`string`|No|`""`|The base VM from which to clone to create the new VM.|
|`full_clone`|`bool`|No|`true`|Set to `true` to create a full clone, or `false` to create a linked clone. See the [docs about cloning](https://pve.proxmox.com/pve-docs/chapter-qm.html#qm_copy_and_clone) for more info. Only applies when `clone` is set.|
|`hastate`|`string`|No|`""`|Requested HA state for the resource. One of "started", "stopped", "enabled", "disabled", or "ignored". See the [docs about HA](https://pve.proxmox.com/pve-docs/chapter-ha-manager.html#ha_manager_resource_config) for more info.|
|`qemu_os`|`string`|No|`"l26"`|The type of OS in the guest. Set properly to allow Proxmox to enable optimizations for the appropriate guest OS.|
|`memory`|``|No|``||

* `` - (Optional; defaults to 512)
* `balloon` - (Optional; defaults to 0)
* `cores` - (Optional; defaults to 1)
* `sockets` - (Optional; defaults to 1)
* `vcpus` - (Optional; defaults to 0)
* `vcpus` - (Optional; defaults to 0)
* `cpu` - (Optional; defaults to host)
* `numa` - (Optional; defaults to false)
* `hotplug` - (Optional; defaults to network,disk,usb)
* `scsihw` - (Optional; defaults to the empty string)
* `vga` - (Optional)
    * `type` (Optional; defauls to std)
    * `memory` (Optional)
* `network` - (Optional)
    * `id` (Required)
    * `model` (Required)
    * `macaddr` (Optional)
    * `bridge` (Optional; defaults to nat)
    * `tag` (Optional; defaults to -1)
    * `firewall` (Optional; defaults to false)
    * `rate` (Optional; defaults to -1)
    * `queues` (Optional; defaults to -1)
    * `link_down` (Optional; defaults to false)
* `disk` - (Optional)
    * `id` (Required)
    * `type` (Required)
    * `storage` (Required)
    * `size` (Required)
    * `format` (Optional; defaults to raw)
    * `cache` (Optional; defaults to none)
    * `backup` (Optional; defaults to false)
    * `iothread` (Optional; defaults to false)
    * `replicate` (Optional; defaults to false)
    * `ssd` (Optional; defaults to false) //Whether to expose this drive as an SSD, rather than a rotational hard disk.
    * `file` (Optional)
    * `media` (Optional)
    * `discard` (Optional; defaults to ignore) //Controls whether to pass discard/trim requests to the underlying storage. discard=<ignore | on>
    * `mbps` (Optional; defaults to unlimited being 0) Maximum r/w speed in megabytes per second
    * `mbps_rd` (Optional; defaults to unlimited being 0) Maximum read speed in megabytes per second
    * `mbps_rd_max` (Optional; defaults to unlimited being 0) Maximum unthrottled read pool in megabytes per second
    * `mbps_wr` (Optional; defaults to unlimited being 0) //Maximum write speed in megabytes per second
    * `mbps_wr_max` (Optional; defaults to unlimited being 0) //Maximum unthrottled write pool in megabytes per second
* `serial` - (Optional)
    * `id` (Required)
    * `type` (Required)
* `pool` - (Optional)
* `force_create` - (Optional; defaults to true)
* `clone_wait` - (Optional; defaults to 15 seconds) Amount of time to wait after a clone operation and after an UpdateConfig operation.
* `additional_wait` - (Optional; defaults to 15 seconds) Provider will wait n/2 seconds after a clone operation and n seconds after an UpdateConfig operation.
* `preprovision` - (Optional; defaults to true)
* `os_type` - (Optional) Which provisioning method to use, based on the OS type. Possible values: ubuntu, centos, cloud-init.
* `force_recreate_on_change_of` (Optional) // Allows this to depend on another resource, that when changed, needs to re-create this vm. An example where this is useful is a cloudinit configuration (as the `cicustom` attribute points to a file not the content).

The following arguments are specifically for Linux for preprovisioning (requires `define_connection_info` to be true).

* `os_network_config` - (Optional) Linux provisioning specific, `/etc/network/interfaces` for Ubuntu and `/etc/sysconfig/network-scripts/ifcfg-eth0` for CentOS.
* `ssh_forward_ip` - (Optional) Address used to connect to the VM
* `ssh_host` - (Optional)
* `ssh_port` - (Optional)
* `ssh_user` - (Optional) Username to login in the VM when preprovisioning.
* `ssh_private_key` - (Optional; sensitive) Private key to login in the VM when preprovisioning.

The following arguments are specifically for Cloud-init for preprovisioning.

* `ci_wait` - (Optional) Cloud-init specific, how to long to wait for preprovisioning.
* `ciuser` - (Optional) Cloud-init specific, overwrite image default user.
* `cipassword` - (Optional) Cloud-init specific, password to assign to the user.
* `cicustom` - (Optional) Cloud-init specific, location of the custom cloud-config files.
* `searchdomain` - (Optional) Cloud-init specific, sets DNS search domains for a container.
* `nameserver` - (Optional) Cloud-init specific, sets DNS server IP address for a container.
* `sshkeys` - (Optional) Cloud-init specific, public ssh keys, one per line
* `ipconfig0` - (Optional) Cloud-init specific, [gw=<GatewayIPv4>] [,gw6=<GatewayIPv6>] [,ip=<IPv4Format/CIDR>] [,ip6=<IPv6Format/CIDR>]
* `ipconfig1` - (Optional) Cloud-init specific, see ipconfig0
* `ipconfig2` - (Optional) Cloud-init specific, see ipconfig0

Deprecated arguments.

* `disk_gb` - (Optional; use disk.size instead)
* `storage` - (Optional; use disk.storage instead)
* `storage_type` - (Optional; use disk.type instead)
* `nic` - (Optional; use network instead)
* `bridge` - (Optional; use network.bridge instead)
* `vlan` - (Optional; use network.tag instead)
* `mac` - (Optional; use network.macaddr instead)
