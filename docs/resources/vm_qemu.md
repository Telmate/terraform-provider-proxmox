# VM Qemu Resource

Resources are the most important element in the Terraform language. Each resource block describes one or more infrastructure objects, such as virtual networks, compute instances, or higher-level components such as DNS records.

This resource manages a Proxmox VM Qemu container.

## Create a Qemu VM resource

You can start from either an ISO or clone an existing VM. Optimally, you could create a VM resource you will use a clone base with an ISO, and make the rest of the VM resources depend on that base "template" and clone it.

When creating a VM Qemu resource, you create a `proxmox_vm_qemu` resource block. The name and target node of the VM are the only required parameters.

```hcl
resource "proxmox_vm_qemu" "resource-name" {
    name = "VM name"
    target_node = "Node to create the VM on"
}
```

## Preprovision

With preprovision you can provision a VM directly from the resource block. This provisioning method is therefore ran **before** provision blocks. When using preprovision, there are three `os_type` options: `ubuntu`, `centos` or `cloud-init`.

```hcl
resource "proxmox_vm_qemu" "prepprovision-test" {
    ...
    preprovision = true
    os_type = "ubuntu"
}
```

### Preprovision for Linux (Ubuntu / CentOS)

There is a pre-provision phase which is used to set a hostname, intialize eth0, and resize the VM disk to available space. This is done over SSH with the `ssh_forward_ip`, `ssh_user` and `ssh_private_key`. Disk resize is done if the file [/etc/auto_resize_vda.sh](https://github.com/Telmate/terraform-ubuntu-proxmox-iso/blob/master/auto_resize_vda.sh) exists.

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

Cloud-init VMs must be cloned from a [cloud-init ready template](https://pve.proxmox.com/wiki/Cloud-Init_Support). When creating a resource that is using Cloud-Init, there are multi configurations possible. You can use either the `ciconfig` parameter to create based on [a Cloud-init configuration file](https://cloudinit.readthedocs.io/en/latest/topics/examples.html) or use the Proxmox variable `ciuser`, `cipassword`, `ipconfig0`, `ipconfig1`, `searchdomain`, `nameserver` and `sshkeys`.

For more information, see the [Cloud-init guide](docs/guides/cloud_init.md).

## Argument reference

The following arguments are supported in the resource block.

|Argument|Type|Required?|Default Value|Description|
|--------|----|---------|-------------|-----------|
|`name`|`string`|**Yes**||The name of the VM within Proxmox.|
|`target_node`|`string`|**Yes**||The name of the Proxmox Node on which to place the VM.|
|`vmid`|`integer`|No|`0`|The ID of the VM in Proxmox. The default value of `0` indicates it should use the next available ID in the sequence.|
|`desc`|`string`|No|`""`|The description of the VM. Shows as the 'Notes' field in the Proxmox GUI.|
|`define_connection_info`|`bool`|No|`true`|Whether to let terraform define the (SSH) connection parameters for preprovisioners, see config block below.|
|`bios`|`string`|No|`"seabios"`|The BIOS to use, options are `seabios` or `ovmf` for UEFI.|
|`onboot`|`bool`|No|`true`|Whether to have the VM startup after the PVE node starts.|
|`boot`|`string`|No|`"cdn"`|The boot order for the VM. Ordered string of characters denoting boot order. Options: floppy (`a`), hard disk (`c`), CD-ROM (`d`), or network (`n`).|
|`bootdisk`|`string`|No|*Computed*|Enable booting from specified disk. This value is computed by terraform, so you shouldn't need to change it under most circumstances.|
|`agent`|`integer`|No|`0`|Whether to enable the QEMU Guest Agent. Note, you must still install the [`qemu-guest-agent`](https://pve.proxmox.com/wiki/Qemu-guest-agent) daemon in the quest for this to have any effect.|
|`iso`|`string`|No|`""`|The name of the ISO image to mount to the VM. Only applies when `clone` is not set.|
|`clone`|`string`|No|`""`|The base VM from which to clone to create the new VM.|
|`full_clone`|`bool`|No|`true`|Set to `true` to create a full clone, or `false` to create a linked clone. See the [docs about cloning](https://pve.proxmox.com/pve-docs/chapter-qm.html#qm_copy_and_clone) for more info. Only applies when `clone` is set.|
|`hastate`|`string`|No|`""`|Requested HA state for the resource. One of "started", "stopped", "enabled", "disabled", or "ignored". See the [docs about HA](https://pve.proxmox.com/pve-docs/chapter-ha-manager.html#ha_manager_resource_config) for more info.|
|`qemu_os`|`string`|No|`"l26"`|The type of OS in the guest. Set properly to allow Proxmox to enable optimizations for the appropriate guest OS.|
|`memory`|`integer`|No|`512`|The amount of memory to allocate to the VM in bytes.|
|`balloon`|`integer`|No|`0`|Whether to add the ballooning device to the VM. Options are `1` and `0`. See the [docs about memory](https://pve.proxmox.com/pve-docs/chapter-qm.html#qm_memory) for more info.|
|`sockets`|`integer`|No|`1`|The number of CPU sockets to allocate to the VM.|
|`cores`|`integer`|No|`1`|The number of CPU cores per CPU socket to allocate to the VM.|
|`vcpus`|`integer`|No|`0`|The number of vCPUs plugged into the VM when it starts. If 0, this is set automatically by Proxmox to `sockets * cores`.|
|`cpu`|`string`|No|`"host"`|The type of CPU to emulate in the Guest. See the [docs about CPU Types](https://pve.proxmox.com/pve-docs/chapter-qm.html#qm_cpu) for more info.|
|`numa`|`bool`|No|`false`|Whether to enable [Non-Uniform Memory Access](https://pve.proxmox.com/pve-docs/chapter-qm.html#qm_cpu) in the guest.|
|`hotplug`|`string`|No|`"network,disk,usb"`|Comma delimited list of hotplug features to enable. Options: `network`, `disk`, `cpu`, `memory`, `usb`. Set to `0` to disable hotplug.|
|`scsihw`|`string`|No|*Computed*|The SCSI controller to emulate, if left empty, proxmox default to `lsi`. Options: `lsi`, `lsi53c810`, `megasas`, `pvscsi`, `virtio-scsi-pci`, `virtio-scsi-single`.|
|`pool`|`string`|No|`""`|The resource pool to which the VM will be added.|
|`force_create`|`bool`|No|`false`|If `false`, and a vm of the same name, on the same node exists, terraform will attempt to reconfigure that VM with these settings. Set to true to always create a new VM (note, the name of the VM must still be unique, otherwise an error will be produced.)|
|`clone_wait`|`integer`|No|`15`|The amount of time in seconds to wait between cloning a VM and performing post-clone actions such as updating the VM.|
|`additional_wait`|`integer`|No|`15`|Provider will wait `additional_wait`/2 seconds after a clone operation and `additional_wait` seconds after an UpdateConfig operation.|
|`preprovision`|`bool`|No|`true`|Whether to preprovision the VM. See [Preprovision](#Preprovision) above for more info.|
|`os_type`|`string`|No|`""`|Which provisioning method to use, based on the OS type. Options: `ubuntu`, `centos`, `cloud-init`.|
|`force_recreate_on_change_of`|`string`|No|`""`|If the value of this string changes, the VM will be recreated. Useful for allowing this resource to be recreated when arbitrary attributes change. An example where this is useful is a cloudinit configuration (as the `cicustom` attribute points to a file not the content).|
|`os_network_config`|`string`|No|`""`|Only applies when `define_connection_info` is true. Network configuration to be copied into the VM when preprovisioning `ubuntu` or `centos` guests. The specified configuration is added to `/etc/network/interfaces` for Ubuntu, or `/etc/sysconfig/network-scripts/ifcfg-eth0` for CentOS. Forces re-creation on change.|
|`ssh_forward_ip`|`string`|No|`""`|Only applies when `define_connection_info` is true. The IP (and optional colon separated port), to use to connect to the host for preprovisioning. If using cloud-init, this can be left blank.|
|`ssh_user`|`string`|No|`""`|Only applies when `define_connection_info` is true. The user with which to connect to the guest for preprovisioning. Forces re-creation on change.|
|`ssh_private_key`|`string`|No|`""`|Only applies when `define_connection_info` is true. The private key to use when connecting to the guest for preprovisioning. Sensitive.|
|`ci_wait`|`integer`|No|`30`|How to long in seconds to wait for before provisioning.|
|`ciuser`|`string`|No|`""`|Override the default cloud-init user for provisioning.|
|`cipassword`|`string`|No|`""`|Override the default cloud-init user's password. Sensitive.|
|`cicustom`|``|No|`""`|Instead specifying ciuser, cipasword, etc... you can specify the path to a custom cloud-init config file here. Grants more flexibility in configuring cloud-init.|
|`searchdomain`|`string`|No|`""`|Sets default DNS search domain suffix.|
|`nameserver`|`string`|No|`""`|Sets default DNS server for guest.|
|`sshkeys`|`string`|No|`""`|Newline delimited list of SSH public keys to add to authorized keys file for the cloud-init user.|
|`ipconfig0`|`string`|No|`""`|The first IP address to assign to the guest. Format: `[gw=<GatewayIPv4>] [,gw6=<GatewayIPv6>] [,ip=<IPv4Format/CIDR>] [,ip6=<IPv6Format/CIDR>]`|
|`ipconfig1`|`string`|No|`""`|The second IP address to assign to the guest. Same format as `ipconfig0`|
|`ipconfig2`|`string`|No|`""`|The third IP address to assign to the guest. Same format as `ipconfig0`|

### VGA Block

The `vga` block is used to configure the display device. It may be specified multiple times, however only the first instance of the block will be used.

|Argument|Type|Required?|Default Value|Description|
|--------|----|---------|-------------|-----------|
|`type`|`string`|No|`"std"`|The type of display to virtualize. See the [docs about display](https://pve.proxmox.com/pve-docs/chapter-qm.html#qm_display) for more details. Options: `cirrus`, `none`, `qxl`, `qxl2`, `qxl3`, `qxl4`, `serial0`, `serial1`, `serial2`, `serial3`, `std`, `virtio`, `vmware`|
|`type`|`integer`|No||Sets the VGA memory (in MiB). Has no effect with serial display type.|

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

## Attribute Reference

In addition to all the arguments above, the following attributes can be referenced from this resource.

|Attribute|Type|Description|
|---------|----|-----------|
|`ssh_host`|`string`|Read-only attribute. Only applies when `define_connection_info` is true. The hostname or IP to use to connect to the VM for preprovisioning. This can be overridden by defining `ssh_forward_ip`, but if you're using cloud-init and `ipconfig0=dhcp`, the IP reported by qemu-guest-agent is used, otherwise the IP defined in `ipconfig0` is used.|
|`ssh_port`|`string`|Read-only attribute. Only applies when `define_connection_info` is true. The port to connect to the VM over SSH for preprovisioning. If using cloud-init and a port is not specified in `ssh_forward_ip`, then 22 is used. If not using cloud-init, a port on the `target_node` will be forwarded to port 22 in the guest, and this attribute will be set to the forwarded port.|

Deprecated arguments.

* `disk_gb` - (Optional; use disk.size instead)
* `storage` - (Optional; use disk.storage instead)
* `storage_type` - (Optional; use disk.type instead)
* `nic` - (Optional; use network instead)
* `bridge` - (Optional; use network.bridge instead)
* `vlan` - (Optional; use network.tag instead)
* `mac` - (Optional; use network.macaddr instead)
