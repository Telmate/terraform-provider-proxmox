# VM Qemu Resource

This resource manages a Proxmox VM Qemu container.

## Create a Qemu VM resource

You can start from either an ISO, PXE boot the VM, or clone an existing VM. Optimally, you could
create a VM resource you will use a clone base with an ISO, and make the rest of the VM resources
depend on that base "template" and clone it.

When creating a VM Qemu resource, you create a `proxmox_vm_qemu` resource block. For ISO and clone
modes, the name and target node of the VM are the only required parameters.

For the PXE mode, the `boot` directive must contain a *Network* in its boot order.  Generally, PXE
boot VMs should NOT contain the Agent config (`agent = 1`).  PXE boot mode also requires external
infrastructure to support the Network PXE boot request by the VM.

```hcl
resource "proxmox_vm_qemu" "resource-name" {
  name        = "VM-name"
  target_node = "Node to create the VM on"

  disks {
    ide {
      ide2 {
        cdrom {
          iso = "ISO file"
        }
      }
    }
  }

  ### or for a Clone VM operation
  # clone = "template to clone"

  ### or for a PXE boot VM operation
  # pxe = true
  # boot = "scsi0;net0"
  # agent = 0
}
```

## Provision through Cloud-Init

Cloud-init VMs must be cloned from a [cloud-init ready template](https://pve.proxmox.com/wiki/Cloud-Init_Support). When
creating a resource that is using Cloud-Init, there are multi configurations possible. You can use either the `cicustom`
parameter to create based
on [a Cloud-init configuration file](https://cloudinit.readthedocs.io/en/latest/topics/examples.html) or use the Proxmox
variable `ciuser`, `cipassword`, `ipconfig0`, `ipconfig1`, `ipconfig2`, `ipconfig3`, `ipconfig4`, `ipconfig5`,
`ipconfig6`, `ipconfig7`, `ipconfig8`, `ipconfig9`, `ipconfig10`, `ipconfig11`, `ipconfig12`, `ipconfig13`,
`ipconfig14`,`ipconfig15`, `searchdomain`, `nameserver` and `sshkeys`.

For more information, see the [Cloud-init guide](../guides/cloud_init.md).

## Provision through PXE Network Boot

Specifying the `pxe = true` option will enable the Virtual Machine to perform a Network Boot (PXE).
In addition to enabling the PXE mode, a few other options should be specified to ensure successful
boot of the VM.  A minimal Resource stanza for a PXE boot VM might look like this:

```hcl
resource "proxmox_vm_qemu" "pxe-minimal-example" {
    name                      = "pxe-minimal-example"
    agent                     = 0
    boot                      = "order=scsi0;net0"
    pxe                       = true
    target_node               = "test"
    network {
        id = 0
        bridge    = "vmbr0"
        firewall  = false
        link_down = false
        model     = "e1000"
    }
}
```

The primary options that effect the correct operation of Network PXE boot mode are:

* `boot`: a valid boot order must be specified with Network type included (eg `order=scsi0;net0`)
* a valid NIC attached to a network with a PXE boot server must be added to the VM
* generally speaking, disable the Agent (`agent = 0`) unless the installed OS contains the Agent in OS install configurations

## Argument reference

**Note: Except where explicitly stated in the description, all arguments are assumed to be optional.**

### Top Level Block

The following arguments are supported in the top level resource block.

| Argument                      | Type     | Default Value        | Description |
| ----------------------------- | -------- | -------------------- | ----------- |
| `name`                        | `str`    |                      | **Required** The name of the VM within Proxmox. |
| `target_node`                 | `str`    |                      | The name of the PVE Node on which to place the VM.|
| `target_nodes`                | `str`    |                      | A list of PVE node names on which to place the VM.|
| `vmid`                        | `int`    |                      | The ID of the VM in Proxmox. When unset it should use the next available ID in the sequence. |
| `description`                 | `str`    |                      | The description of the VM. Shows as the 'Notes' field in the Proxmox GUI. |
| `define_connection_info`      | `bool`   | `true`               | Whether to let terraform define the (SSH) connection parameters for preprovisioners, see config block below. |
| `bios`                        | `str`    | `"seabios"`          | The BIOS to use, options are `seabios` or `ovmf` for UEFI. |
| `start_at_node_boot`          | `bool`   | `false`              | Whether the guest should start automatically when the Proxmox node boots.|
| `startup_shutdown`            | `nested` |                      | Startup and shutdown configuration of the guest, see [Startup and Shutdown Reference](#startup-and-shutdown-reference).|
| `vm_state`                    | `string` | `"running"`          | The desired state of the VM, options are `running`, `stopped` and `started`. Do note that `started` will only start the vm on creation and won't fully manage the power state unlike `running` and `stopped` do. |
| `oncreate`                    | `bool`   | `true`               | Whether to have the VM startup after the VM is created (deprecated, use `vm_state` instead) |
| `protection`                  | `bool`   | `false`              | Enable/disable the VM protection from being removed. The default value of `false` indicates the VM is removable. |
| `tablet`                      | `bool`   | `true`               | Enable/disable the USB tablet device. This device is usually needed to allow absolute mouse positioning with VNC. |
| `boot`                        | `str`    |                      | The boot order for the VM. For example: `order=scsi0;ide2;net0`. The deprecated `legacy=` syntax is no longer supported. See the `boot` option in the [Proxmox manual](https://pve.proxmox.com/wiki/Manual:_qm.conf#_options) for more information. |
| `bootdisk`                    | `str`    |                      | Enable booting from specified disk. You shouldn't need to change it under most circumstances. |
| `agent`                       | `int`    | `0`                  | Set to `1` to enable the QEMU Guest Agent. Note, you must run the [`qemu-guest-agent`](https://pve.proxmox.com/wiki/Qemu-guest-agent) daemon in the guest for this to have any effect. |
| `pxe`                         | `bool`   | `false`              | If set to `true`, enable PXE boot of the VM.  Also requires a `boot` order be set with Network included (eg `boot = "order=scsi0;net0"`).  Note that `pxe` is mutually exclusive with `clone` modes. |
| `clone`                       | `str`    |                      | The base VM name from which to clone to create the new VM.  Note that `clone` is mutually exclusive with `clone_id` and `pxe` modes. |
| `clone_id`                    | `int`    |                      | The base VM id from which to clone to create the new VM.  Note that `clone_id` is mutually exclusive with `clone` and `pxe` modes. |
| `full_clone`                  | `bool`   | `true`               | Set to `true` to create a full clone, or `false` to create a linked clone. See the [docs about cloning](https://pve.proxmox.com/pve-docs/chapter-qm.html#qm_copy_and_clone) for more info. Only applies when `clone` is set. |
| `hastate`                     | `str`    |                      | Requested HA state for the resource. One of "started", "stopped", "enabled", "disabled", or "ignored". See the [docs about HA](https://pve.proxmox.com/pve-docs/chapter-ha-manager.html#ha_manager_resource_config) for more info. |
| `hagroup`                     | `str`    |                      | The HA group identifier the resource belongs to (requires `hastate` to be set!). See the [docs about HA](https://pve.proxmox.com/pve-docs/chapter-ha-manager.html#ha_manager_resource_config) for more info. |
| `qemu_os`                     | `str`    | `"l26"`              | The type of OS in the guest. Set properly to allow Proxmox to enable optimizations for the appropriate guest OS. It takes the value from the source template and ignore any changes to resource configuration parameter. |
| `memory`                      | `int`    | `512`                | The amount of memory to allocate to the VM in Megabytes. |
| `balloon`                     | `int`    | `0`                  | The minimum amount of memory to allocate to the VM in Megabytes, when Automatic Memory Allocation is desired. Proxmox will enable a balloon device on the guest to manage dynamic allocation. See the [docs about memory](https://pve.proxmox.com/pve-docs/chapter-qm.html#qm_memory) for more info. |
| `hotplug`                     | `str`    | `"network,disk,usb"` | Comma delimited list of hotplug features to enable. Options: `network`, `disk`, `cpu`, `memory`, `usb`. Set to `0` to disable hotplug. |
| `scsihw`                      | `str`    | `"lsi"`              | The SCSI controller to emulate. Options: `lsi`, `lsi53c810`, `megasas`, `pvscsi`, `virtio-scsi-pci`, `virtio-scsi-single`. |
| `pool`                        | `str`    |                      | The resource pool to which the VM will be added. |
| `tags`                        | `str`    |                      | Tags of the VM. Comma-separated values (e.g. `tag1,tag2,tag3`). Tag may not start with `-` and may only include the following characters: `[a-z]`, `[0-9]`, `_` and `-`. This is only meta information. |
| `rng`                         | `struct` |                      | The RNG device to add to the VM, more info in [RNG Block](#rng-block) section. |
| `tpm_state`                   | `struct` |                      | The TPM device to add to the VM, more info in [TPM Block](#tpm-block) section. |
| `force_create`                | `bool`   | `false`              | If `false`, and a vm of the same name, on the same node exists, terraform will attempt to reconfigure that VM with these settings. Set to true to always create a new VM (note, the name of the VM must still be unique, otherwise an error will be produced.) |
| `os_type`                     | `str`    |                      | Which provisioning method to use, based on the OS type. Options: `ubuntu`, `centos`, `cloud-init`. |
| `force_recreate_on_change_of` | `str`    |                      | If the value of this string changes, the VM will be recreated. Useful for allowing this resource to be recreated when arbitrary attributes change. An example where this is useful is a cloudinit configuration (as the `cicustom` attribute points to a file not the content). |
| `os_network_config`           | `str`    |                      | Only applies when `define_connection_info` is true. Network configuration to be copied into the VM when preprovisioning `ubuntu` or `centos` guests. The specified configuration is added to `/etc/network/interfaces` for Ubuntu, or `/etc/sysconfig/network-scripts/ifcfg-eth0` for CentOS. Forces re-creation on change. |
| `ssh_forward_ip`              | `str`    |                      | Only applies when `define_connection_info` is true. The IP (and optional colon separated port), to use to connect to the host for preprovisioning. If using cloud-init, this can be left blank. |
| `ssh_user`                    | `str`    |                      | Only applies when `define_connection_info` is true. The user with which to connect to the guest for preprovisioning. Forces re-creation on change. |
| `ssh_private_key`             | `str`    |                      | Only applies when `define_connection_info` is true. The private key to use when connecting to the guest for preprovisioning. Sensitive. |
| `ci_wait`                     | `int`    | `30`                 | How to long in seconds to wait for before provisioning. |
| `ciuser`                      | `str`    |                      | Override the default cloud-init user for provisioning. |
| `cipassword`                  | `str`    |                      | Override the default cloud-init user's password. Sensitive. |
| `cicustom`                    | `str`    |                      | Instead specifying ciuser, cipasword, etc... you can specify the path to a custom cloud-init config file here. Grants more flexibility in configuring cloud-init. |
| `ciupgrade`                   | `bool`   | `false`              | Whether to upgrade the packages on the guest during provisioning. Restarts the VM when set to `true`. |
| `searchdomain`                | `str`    |                      | Sets default DNS search domain suffix. |
| `nameserver`                  | `str`    |                      | Sets default DNS server for guest. |
| `sshkeys`                     | `str`    |                      | Newline delimited list of SSH public keys to add to authorized keys file for the cloud-init user. |
| `ipconfig0`                   | `str`    | `''`                 | The first IP address to assign to the guest. Format: `[gw=<GatewayIPv4>] [,gw6=<GatewayIPv6>] [,ip=<IPv4Format/CIDR>] [,ip6=<IPv6Format/CIDR>]`. When `os_type` is `cloud-init` not setting `ip=` is equivalent to `skip_ipv4` == `true` and `ip6=` to `skip_ipv6` == `true` .|
| `ipconfig1` to `ipconfig15`   | `str`    |                      | The second IP address to assign to the guest. Same format as `ipconfig0`. |
| `automatic_reboot`            | `bool`   | `true`               | Automatically reboot the VM when parameter changes require this. If disabled the provider will emit a warning or error when the VM needs to be rebooted, this can be configured with `automatic_reboot_severity`.|
| `automatic_reboot_severity`   | `string`  | `error`              | Sets the severity of the error/warning when `automatic_reboot` is `false`. Values can be `error` or `warning`.|
| `skip_ipv4`                   | `bool`   | `false`              | Tells proxmox that acquiring an IPv4 address from the qemu guest agent isn't required, it will still return an ipv4 address if it could obtain one. Useful for reducing retries in environments without ipv4.|
| `skip_ipv6`                   | `bool`   | `false`              | Tells proxmox that acquiring an IPv6 address from the qemu guest agent isn't required, it will still return an ipv6 address if it could obtain one. Useful for reducing retries in environments without ipv6.|
| `agent_timeout`               | `int`    | `90`                 | Timeout in seconds to keep trying to obtain an IP address from the guest agent one we have a connection. |
| `current_node`                | `string` |                      | **Computed** The current node of the Qemu guest is on.|

### CPU Block

The `cpu` block is used to configure the CPU settings. It may be specified once.
See the [docs about CPU](https://pve.proxmox.com/pve-docs/chapter-qm.html#qm_cpu) for more details.

| Argument  | Type  | Default Value | Description |
| --------- | ----- | ------------- |:----------- |
| `affinity`| `str` | `""`          | The CPU affinity for the Qemu guest. This is a comma separated list of values and ranges which define to which CPU cores the Qemu guest is bound. Example: `1,3-5`.|
| `cores`   | `int` | `1`           | The number of CPU cores to allocate to the Qemu guest.|
| `limit`   | `int` | `0`           | The CPU limit for the Qemu guest. `0` means unlimited.|
| `numa`    | `bool`| `false`       | Whether to enable [Non-Uniform Memory Access](https://pve.proxmox.com/pve-docs/chapter-qm.html#qm_cpu) in the Qemu guest.|
| `sockets` | `int` | `1`           | The number of CPU sockets to allocate to the Qemu guest.|
| `type`    | `str` | `"host"`      | The CPU type to emulate. See the [docs about CPU Types](https://pve.proxmox.com/pve-docs/chapter-qm.html#_cpu_type) for more info.|
| `units`   | `int` | `0`        | The CPU units for the Qemu guest. This is a relative value which defines the CPU weight of the Qemu guest. The default value of `0` indicates the PVE default is used.|
| `vcores`  | `int` | `0`        | The number of virtual cores exposed to the Qemu guest. If `0`, this is set automatically by Proxmox to `sockets * cores`.|
| `flags`   | `list`|         | The CPU flags to enable for the Qemu guest. |

#### CPU Flags Block

The CPU flags to enable for the Qemu guest. A flag can be set with `on` ot `off`, when a flag isn't specified, it will be set to the default value in PVE and will be inherited from the `cpu.type`.

| Argument     | Type | Description |
| ------------ | ---- | ----------- |
| `md_clear`   | `str`| Required to let the guest OS know if MDS is mitigated correctly.|
| `pcid`       | `str`| Meltdown fix cost reduction on Westmere, Sandy-, and IvyBridge Intel CPUs.|
| `spec_ctrl`  | `str`| Allows improved Spectre mitigation with Intel CPUs.|
| `ssbd`       | `str`| Protection for "Speculative Store Bypass" for Intel models.|
| `ibpb`       | `str`| Allows improved Spectre mitigation with AMD CPUs.|
| `virt_ssbd`  | `str`| Basis for "Speculative Store Bypass" protection for AMD models.|
| `amd_ssbd`   | `str`| Improves Spectre mitigation performance with AMD CPUs, best used with "virt-ssbd".|
| `amd_no_ssb` | `str`| Notifies guest OS that host is not vulnerable for Spectre on AMD CPUs.|
| `pbpe1gb`    | `str`| Allow guest OS to use 1GB size pages, if host HW supports it.|
| `hv_tlbflush`| `str`| Improve performance in overcommitted Windows guests. May lead to guest bluescreens on old CPUs.|
| `hv_evmcs`   | `str`| Improve performance for nested virtualization. Only supported on Intel CPUs.|
| `aes`        | `str`| Activate AES instruction set for HW acceleration.|

### VGA Block

The `vga` block is used to configure the display device. It may be specified multiple times, however only the first
instance of the block will be used.

See the [docs about display](https://pve.proxmox.com/pve-docs/chapter-qm.html#qm_display) for more details.

| Argument | Type  | Default Value | Description                                                                                                                                                         |
| -------- | ----- | ------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `type`   | `str` | `"std"`       | The type of display to virtualize. Options: `cirrus`, `none`, `qxl`, `qxl2`, `qxl3`, `qxl4`, `serial0`, `serial1`, `serial2`, `serial3`, `std`, `virtio`, `vmware`. |
| `memory` | `int` |               | Sets the VGA memory (in MiB). Has no effect with serial display type.                                                                                               |

### Network Block

The `network` block is used to configure the network devices. It may be specified multiple times. The order in which the
blocks are specified determines the ID for each net device. i.e. The first `network` block will become `net0`, the
second will be `net1` etc...

See the [docs about network devices](https://pve.proxmox.com/pve-docs/chapter-qm.html#qm_network_device) for more
details.

| Argument    | Type   | Default Value | Description |
| ----------- | ------ | ------------- | ----------- |
| `id`        | `int`  |               | **Required** The ID of the network device `0`-`31`. |
| `model`     | `str`  |               | **Required** Network Card Model. The virtio model provides the best performance with very low CPU overhead. If your guest does not support this driver, it is usually best to use e1000. Options: `e1000`, `e1000-82540em`, `e1000-82544gc`, `e1000-82545em`, `i82551`, `i82557b`, `i82559er`, `ne2k_isa`, `ne2k_pci`, `pcnet`, `rtl8139`, `virtio`, `vmxnet3`. |
| `macaddr`   | `str`  |               | Override the randomly generated MAC Address for the VM. Requires the MAC Address be Unicast.  |
| `bridge`    | `str`  | `"nat"`       | Bridge to which the network device should be attached. The Proxmox VE standard bridge is called `vmbr0`. |
| `tag`       | `int`  | `0`           | The VLAN tag to apply to packets on this device. `0` disables VLAN tagging. |
| `firewall`  | `bool` | `false`       | Whether to enable the Proxmox firewall on this network device. |
| `mtu`       | `int`  |               | The MTU value for the network device. On ``virtio`` models, set to ``1`` to inherit the MTU value from the underlying bridge. |
| `rate`      | `int`  | `0`           | Network device rate limit in mbps (megabytes per second) as floating point number. Set to `0` to disable rate limiting. |
| `queues`    | `int`  | `1`           | Number of packet queues to be used on the device. Requires `virtio` model to have an effect. |
| `link_down` | `bool` | `false`       | Whether this interface should be disconnected (like pulling the plug). |

### Disk Block

The `disk` block is used to configure the disk devices. It may be specified multiple times. This block does not diff as pretty as the `disks` block, but it is more flexible for modules. Putting the disks in alphanumeric order based on the value of `slot` is recommended for readability.

For `type` there is a special `ignore` value. This will tell Terraform to not manage the disk in that slot, useful when another tool manages the disks.

Due to the complexity of the `disk` block, there is a settings matrix that can be found in the [Disk compatibility matrix](#disk-compatibility-matrix).

| Argument             | Type   |Default| Description|
|:---------------------|:------:|:-----:|:-----------|
|`asyncio`             |`string`|       |The drive's asyncio setting. Options: `io_uring`, `native`, `threads`|
|`backup`              |`bool`  |`true` |Whether the drive should be included when making backups.|
|`cache`               |`string`|       |The drive’s cache mode. Options: `directsync`, `none`, `unsafe`, `writeback`, `writethrough`.|
|`discard`             |`bool`  |`false`|Controls whether to pass discard/trim requests to the underlying storage. Only effective when the underlying storage supports thin provisioning. There are other caveats too, see the [docs about disks](https://pve.proxmox.com/pve-docs/chapter-qm.html#qm_hard_disk) for more info.|
|`disk_file`           |`string`|       |The path to the disk file. **Required** when `type`=`disk` and `passthrough`=`true`|
|`emulatessd`          |`bool`  |`false`|Whether to expose this drive as an SSD, rather than a rotational hard disk.|
|`format`              |`string`|`raw`  |The drive’s backing file’s data format. Default only applies when `type`=`disk` and `passthrough`=`false`|
|`id`                  |`int`   |       |**Computed** Unique id of the disk.|
|`iops_r_burst`        |`int`   |`0`    |Maximum number of iops while reading in short bursts. `0` means unlimited.|
|`iops_r_burst_length` |`int`   |`0`    |Length of the read burst duration in seconds. `0` means the default duration dictated by proxmox.|
|`iops_r_concurrent`   |`int`   |`0`    |Maximum number of iops while reading concurrently. `0` means unlimited.|
|`iops_wr_burst`       |`int`   |`0`    |Maximum number of iops while writing in short bursts. `0` means unlimited.|
|`iops_wr_burst_length`|`int`   |`0`    |Length of the write burst duration in seconds. `0` means the default duration dictated by proxmox.|
|`iops_wr_concurrent`  |`int`   |`0`    |Maximum number of iops while writing concurrently. `0` means unlimited.|
|`iothread`            |`bool`  |`false`|Whether to use iothreads for this drive. Only effective when the the emulated controller type (`scsihw` top level block argument) is `virtio-scsi-single`.|
|`iso`                 |`string`|       |The name of the ISO image to mount to the VM in the format: [storage pool]:iso/[name of iso file]. Note that `iso` is mutually exclusive with `passthrough`.|
|`linked_disk_id`      |`int`   |       |**Computed** The `vmid` of the linked vm this disk was cloned from.|
|`mbps_r_burst`        |`float` |`0.0`  |Maximum read speed in megabytes per second. `0` means unlimited.|
|`mbps_r_concurrent`   |`float` |`0.0`  |Maximum read speed in megabytes per second. `0` means unlimited.|
|`mbps_wr_burst`       |`float` |`0.0`  |Maximum write speed in megabytes per second. `0` means unlimited.|
|`mbps_wr_concurrent`  |`float` |`0.0`  |Maximum throttled write pool in megabytes per second. `0` means unlimited.|
|`passthrough`         |`bool`  |`false`|Wether the physical cdrom drive should be passed through.|
|`readonly`            |`bool`  |`false`|Whether the drive should be readonly.|
|`replicate`           |`bool`  |`false`|Whether the drive should considered for replication jobs.|
|`serial`              |`string`|       |The serial number of the disk.|
|`size`                |`string`|       |The size of the created disk. Accepts `K` for kibibytes, `M` for mebibytes, `G` for gibibytes, `T` for tibibytes. When only a number is provided gibibytes is assumed. **Required** when `type`=`disk` and `passthrough`=`false`, **Computed** when `type`=`disk` and `passthrough`=`true`. |
|`slot`                |`string`|       |**Required** The slot id of the disk - must be one of 'ide0', 'ide1', 'ide2', 'sata0', 'sata1', 'sata2', 'sata3', 'sata4', 'sata5', 'scsi0', 'scsi1', 'scsi2', 'scsi3', 'scsi4', 'scsi5', 'scsi6', 'scsi7', 'scsi8', 'scsi9', 'scsi10', 'scsi11', 'scsi12', 'scsi13', 'scsi14', 'scsi15', 'scsi16', 'scsi17', 'scsi18', 'scsi19', 'scsi20', 'scsi21', 'scsi22', 'scsi23', 'scsi24', 'scsi25', 'scsi26', 'scsi27', 'scsi28', 'scsi29', 'scsi30', 'virtio0', 'virtio1', 'virtio2', 'virtio3', 'virtio4', 'virtio5', 'virtio6', 'virtio7', 'virtio8', 'virtio9', 'virtio10', 'virtio11', 'virtio12', 'virtio13', 'virtio14', 'virtio15'|
|`storage`             |`string`|       |Required when `type`=`disk` and `passthrough`=`false`. The name of the storage pool on which to store the disk.|
|`type`                |`string`|`disk` |The type of disk to create. Options: `cdrom`, `cloudinit` ,`disk` ,`ignore`.|
|`wwn`                 |`string`|       |The WWN of the disk.|

Example `Disk block` using an existing vm as a template.

```hcl
disk {
  type        = "disk"
  disk_file   = "local-lvm:vm-<<<vmid>>>-disk-<<<disk number>>>"
  passthrough = true
  slot        = "scsi0"
}
```

#### Disk compatibility matrix

**Note** `cloudinit` can only be used with `ide`, `sata` and `scsi` disk types.

| Argument             | Disk Type         | Disk Slot           |Passthrough|
|:---------------------|:-----------------:|:-------------------:|:---------:|
|`asyncio`             |`disk`             |`all`                |`both`     |
|`backup`              |`disk`             |`all`                |`both`     |
|`cache`               |`disk`             |`all`                |`both`     |
|`discard`             |`disk`             |`all`                |`both`     |
|`disk_file`           |`disk`             |`all`                |`true`     |
|`emulatessd`          |`disk`             |`ide`, `sata`, `scsi`|`both`     |
|`format`              |`disk`             |`all`                |`both`     |
|`id`                  |`disk`             |`all`                |`false`    |
|`iops_r_burst`        |`disk`             |`all`                |`both`     |
|`iops_r_burst_length` |`disk`             |`all`                |`both`     |
|`iops_r_concurrent`   |`disk`             |`all`                |`both`     |
|`iops_wr_burst`       |`disk`             |`all`                |`both`     |
|`iops_wr_burst_length`|`disk`             |`all`                |`both`     |
|`iops_wr_concurrent`  |`disk`             |`all`                |`false`    |
|`iothread`            |`disk`             |`scsi`, `virtio`     |`false`    |
|`iso`                 |`iso`              |`all`                |`false`    |
|`linked_disk_id`      |`disk`             |`all`                |`false`    |
|`mbps_r_burst`        |`disk`             |`all`                |`false`    |
|`mbps_r_concurrent`   |`disk`             |`all`                |`false`    |
|`mbps_wr_burst`       |`disk`             |`all`                |`false`    |
|`mbps_wr_concurrent`  |`disk`             |`all`                |`false`    |
|`passthrough`         |`disk`, `iso`      |`all`                |`false`    |
|`readonly`            |`disk`             |`scsi`, `virtio`     |`false`    |
|`replicate`           |`disk`             |`all`                |`false`    |
|`serial`              |`disk`             |`all`                |`false`    |
|`size`                |`disk`             |`all`                |`false`    |
|`slot`                |`disk`, `iso`      |`all`                |`false`    |
|`storage`             |`disk`, `cloudinit`|`all`                |`false`    |
|`type`                |`disk`             |`all`                |`false`    |
|`wwn`                 |`disk`             |`all`                |`false`    |

### Disks Block

The `disks` block is used to configure the disk devices. It may be specified once. There are four types of disk `ide`,`sata`,`scsi` and `virtio`. Configuration for these sub types can be found in their respective chapters:

* `ide`: [Disks.Ide Block](#diskside-block).
* `sata`: [Disks.Sata Block](#diskssata-block).
* `scsi`: [Disks.Scsi Block](#disksscsi-block).
* `virtio`: [Disks.Virtio Block](#disksvirtio-block).

For each disk slot there is a special `ignore` setting that can be set to `true`. This will tell Terraform to not manage the disk in that slot, useful when another tool manages the disks.

```hcl
resource "proxmox_vm_qemu" "resource-name" {
  //<arguments omitted for brevity...>

  disks {
    ide {
      //<arguments omitted for brevity...>
    }
    sata {
      //<arguments omitted for brevity...>
    }
    scsi {
      //<arguments omitted for brevity...>
    }
    virtio {
      //<arguments omitted for brevity...>
    }
  }
}
```

### Disks.Ide Block

The `disks.ide` block is used to configure disks of type ide. It may only be specified once. It has the options `ide0` through `ide3`. Each disk can have only one of the following mutually exclusive sub types `cdrom`, `cloudinit`, `disk`, `passthrough`, `ignore`. Configuration for these sub types can be found in their respective chapters:

* `cdrom`: [Disks.x.Cdrom Block](#disksxcdrom-block).
* `cloudinit`: [Disks.x.Cloudinit Block](#disksxcloudinit-block).
* `disk`: [Disks.x.Disk Block](#disksxdisk-block).
* `passthrough`: [Disks.x.Passthrough Block](#disksxpassthrough-block).

```hcl
resource "proxmox_vm_qemu" "resource-name" {
  //<arguments omitted for brevity...>

  disks {
    ide {
      ide0 {
        cdrom {
          //<arguments omitted for brevity...>
        }
      }
      ide1 {
        cloudinit {
          //<arguments omitted for brevity...>
        }
      }
      ide2 {
        disk {
          //<arguments omitted for brevity...>
        }
      }
      ide3 {
        passthrough {
          //<arguments omitted for brevity...>
        }
      }
    }
    //<arguments omitted for brevity...>
  }
}
```

### Disks.Sata Block

The `disks.sata` block is used to configure disks of type sata. It may only be specified once. It has the options `sata0` through `sata5`. Each disk can have only one of the following mutually exclusive sub types `cdrom`, `cloudinit`, `disk`, `passthrough`, `ignore`. Configuration for these sub types can be found in their respective chapters:

* `cdrom`: [Disks.x.Cdrom Block](#disksxcdrom-block).
* `cloudinit`: [Disks.x.Cloudinit Block](#disksxcloudinit-block).
* `disk`: [Disks.x.Disk Block](#disksxdisk-block).
* `passthrough`: [Disks.x.Passthrough Block](#disksxpassthrough-block).

```hcl
resource "proxmox_vm_qemu" "resource-name" {
  //<arguments omitted for brevity...>

  disks {
    sata {
      sata0 {
        cdrom {
          //<arguments omitted for brevity...>
        }
      }
      sata1 {
        cloudinit {
          //<arguments omitted for brevity...>
        }
      }
      sata2 {
        disk {
          //<arguments omitted for brevity...>
        }
      }
      sata3 {
        passthrough {
          //<arguments omitted for brevity...>
        }
      }
      sata4 {
        ignore = true
      }
      //<arguments omitted for brevity...>
    }
    //<arguments omitted for brevity...>
  }
}
```

### Disks.Scsi Block

The `disks.scsi` block is used to configure disks of type scsi. It may only be specified once. It has the options `scsi0` through `scsi30`. Each disk can have only one of the following mutually exclusive sub types `cdrom`, `cloudinit`, `disk`, `passthrough` `ignore`. Configuration for these sub types can be found in their respective chapters:

* `cdrom`: [Disks.x.Cdrom Block](#disksxcdrom-block).
* `cloudinit`: [Disks.x.Cloudinit Block](#disksxcloudinit-block).
* `disk`: [Disks.x.Disk Block](#disksxdisk-block).
* `passthrough`: [Disks.x.Passthrough Block](#disksxpassthrough-block).

```hcl
resource "proxmox_vm_qemu" "resource-name" {
  //<arguments omitted for brevity...>

  disks {
    scsi {
      scsi0 {
        cdrom {
          //<arguments omitted for brevity...>
        }
      }
      scsi1 {
        cloudinit {
          //<arguments omitted for brevity...>
        }
      }
      scsi2 {
        disk {
          //<arguments omitted for brevity...>
        }
      }
      scsi3 {
        passthrough {
          //<arguments omitted for brevity...>
        }
      }
      scsi4 {
        ignore = true
      }
      //<arguments omitted for brevity...>
    }
    //<arguments omitted for brevity...>
  }
}
```

### Disks.Virtio Block

The `disks.virtio` block is used to configure disks of type virtio. It may only be specified once. It has the options `virtio0` through `virtio15`. Each disk can have only one of the following mutually exclusive sub types `cdrom`, `disk`, `passthrough`, `ignore`. Configuration for these sub types can be found in their respective chapters:

* `cdrom`: [Disks.x.Cdrom Block](#disksxcdrom-block).
* `disk`: [Disks.x.Disk Block](#disksxdisk-block).
* `passthrough`: [Disks.x.Passthrough Block](#disksxpassthrough-block).

```hcl
resource "proxmox_vm_qemu" "resource-name" {
  //<arguments omitted for brevity...>

  disks {
    virtio {
      virtio0 {
        cdrom {
          //<arguments omitted for brevity...>
        }
      }
      virtio1 {
        disk {
          //<arguments omitted for brevity...>
        }
      }
      virtio2 {
        passthrough {
          //<arguments omitted for brevity...>
        }
      }
      virtio3 {
        ignore = true
      }
      //<arguments omitted for brevity...>
    }
    //<arguments omitted for brevity...>
  }
}
```

### Disks.x.Cdrom Block

| Argument      | Type   | Default Value | Description                                                                                                                                                  |
| :------------ | :----- | :-----------: | :----------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `iso`         | `str`  |               | The name of the ISO image to mount to the VM in the format: [storage pool]:iso/[name of iso file]. Note that `iso` is mutually exclusive with `passthrough`. |
| `passthrough` | `bool` |    `false`    | Wether the physical cdrom drive should be passed through.                                                                                                    |

When `iso` and `passthrough` are omitted an empty cdrom drive will be created.

### Disks.x.Cloudinit Block

Only **one** `cloudinit` block can be specified globally. This block is used to configure the cloud-init drive.

| Argument  | Type  | Default Value | Description                                                                                              |
| :-------- | :---- | :------------ | :------------------------------------------------------------------------------------------------------- |
| `storage` | `str` |               | The name of the storage pool on which to store the cloud-init drive. **Required** when using cloud-init. |

### Disks.x.Disk Block

See the [docs about disks](https://pve.proxmox.com/pve-docs/chapter-qm.html#qm_hard_disk) for more details.

| Argument               |  Type   | Default Value |      Disk Types       | Description                                                                                                                                                                                                                                                                            |
| :--------------------- | :-----: | :-----------: | :-------------------: | :------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `asyncio`              |  `str`  |               |         `all`         | The drive's asyncio setting. Options: `io_uring`, `native`, `threads`                                                                                                                                                                                                                  |
| `backup`               | `bool`  |    `true`     |         `all`         | Whether the drive should be included when making backups.                                                                                                                                                                                                                              |
| `cache`                |  `str`  |               |         `all`         | The drive’s cache mode. Options: `directsync`, `none`, `unsafe`, `writeback`, `writethrough`.                                                                                                                                                                                          |
| `discard`              | `bool`  |    `false`    |         `all`         | Controls whether to pass discard/trim requests to the underlying storage. Only effective when the underlying storage supports thin provisioning. There are other caveats too, see the [docs about disks](https://pve.proxmox.com/pve-docs/chapter-qm.html#qm_hard_disk) for more info. |
| `emulatessd`           | `bool`  |    `false`    | `ide`, `sata`, `scsi` | Whether to expose this drive as an SSD, rather than a rotational hard disk.                                                                                                                                                                                                            |
| `format`               |  `str`  |     `raw`     |         `all`         | The drive’s backing file’s data format.                                                                                                                                                                                                                                                |
| `id`                   |  `int`  |               |         `all`         | **Computed** Unique id of the disk.                                                                                                                                                                                                                                                    |
| `iops_r_burst`         |  `int`  |      `0`      |         `all`         | Maximum number of iops while reading in short bursts. `0` means unlimited.                                                                                                                                                                                                             |
| `iops_r_burst_length`  |  `int`  |      `0`      |         `all`         | Length of the read burst duration in seconds. `0` means the default duration dictated by proxmox.                                                                                                                                                                                      |
| `iops_r_concurrent`    |  `int`  |      `0`      |         `all`         | Maximum number of iops while reading concurrently. `0` means unlimited.                                                                                                                                                                                                                |
| `iops_wr_burst`        |  `int`  |      `0`      |         `all`         | Maximum number of iops while writing in short bursts. `0` means unlimited.                                                                                                                                                                                                             |
| `iops_wr_burst_length` |  `int`  |      `0`      |         `all`         | Length of the write burst duration in seconds. `0` means the default duration dictated by proxmox.                                                                                                                                                                                     |
| `iops_wr_concurrent`   |  `int`  |      `0`      |         `all`         | Maximum number of iops while writing concurrently. `0` means unlimited.                                                                                                                                                                                                                |
| `iothread`             | `bool`  |    `false`    |   `scsi`, `virtio`    | Whether to use iothreads for this drive. Only effective when the the emulated controller type (`scsihw` top level block argument) is `virtio-scsi-single`.                                                                                                                             |
| `linked_disk_id`       |  `int`  |               |         `all`         | **Computed** The `vmid` of the linked vm this disk was cloned from.                                                                                                                                                                                                                    |
| `mbps_r_burst`         | `float` |     `0.0`     |         `all`         | Maximum read speed in megabytes per second. `0` means unlimited.                                                                                                                                                                                                                       |
| `mbps_r_concurrent`    | `float` |     `0.0`     |         `all`         | Maximum read speed in megabytes per second. `0` means unlimited.                                                                                                                                                                                                                       |
| `mbps_wr_burst`        | `float` |     `0.0`     |         `all`         | Maximum write speed in megabytes per second. `0` means unlimited.                                                                                                                                                                                                                      |
| `mbps_wr_concurrent`   | `float` |     `0.0`     |         `all`         | Maximum throttled write pool in megabytes per second. `0` means unlimited.                                                                                                                                                                                                             |
| `readonly`             | `bool`  |    `false`    |   `scsi`, `virtio`    | Whether the drive should be readonly.                                                                                                                                                                                                                                                  |
| `replicate`            | `bool`  |    `false`    |         `all`         | Whether the drive should considered for replication jobs.                                                                                                                                                                                                                              |
| `serial`               |  `str`  |               |         `all`         | The serial number of the disk.                                                                                                                                                                                                                                                         |
| `size`                 | `string`|               |         `all`         | **Required** The size of the created disk. Accepts `K` for kibibytes, `M` for mebibytes, `G` for gibibytes, `T` for tibibytes. When only a number is provided gibibytes is assumed.|
| `storage`              |  `str`  |               |         `all`         | **Required** The name of the storage pool on which to store the disk.                                                                                                                                                                                                                  |

### Disks.x.Passthrough Block

See the [docs about disks](https://pve.proxmox.com/pve-docs/chapter-qm.html#qm_hard_disk) for more details.

| Argument               |  Type   | Default Value |      Disk Types       | Description                                                                                                                                                                                                                                                                            |
| :--------------------- | :-----: | :-----------: | :-------------------: | :------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `asyncio`              |  `str`  |               |         `all`         | The drive's asyncio setting. Options: `io_uring`, `native`, `threads`                                                                                                                                                                                                                  |
| `backup`               | `bool`  |    `true`     |         `all`         | Whether the drive should be included when making backups.                                                                                                                                                                                                                              |
| `cache`                |  `str`  |               |         `all`         | The drive’s cache mode. Options: `directsync`, `none`, `unsafe`, `writeback`, `writethrough`.                                                                                                                                                                                          |
| `discard`              | `bool`  |    `false`    |         `all`         | Controls whether to pass discard/trim requests to the underlying storage. Only effective when the underlying storage supports thin provisioning. There are other caveats too, see the [docs about disks](https://pve.proxmox.com/pve-docs/chapter-qm.html#qm_hard_disk) for more info. |
| `emulatessd`           | `bool`  |    `false`    | `ide`, `sata`, `scsi` | Whether to expose this drive as an SSD, rather than a rotational hard disk.                                                                                                                                                                                                            |
| `file`                 |  `str`  |               |         `all`         | **Required** The full unix file path to the disk.                                                                                                                                                                                                                                      |
| `iops_r_burst`         |  `int`  |      `0`      |         `all`         | Maximum number of iops while reading in short bursts. `0` means unlimited.                                                                                                                                                                                                             |
| `iops_r_burst_length`  |  `int`  |      `0`      |         `all`         | Length of the read burst duration in seconds. `0` means the default duration dictated by proxmox.                                                                                                                                                                                      |
| `iops_r_concurrent`    |  `int`  |      `0`      |         `all`         | Maximum number of iops while reading concurrently. `0` means unlimited.                                                                                                                                                                                                                |
| `iops_wr_burst`        |  `int`  |      `0`      |         `all`         | Maximum number of iops while writing in short bursts. `0` means unlimited.                                                                                                                                                                                                             |
| `iops_wr_burst_length` |  `int`  |      `0`      |         `all`         | Length of the write burst duration in seconds. `0` means the default duration dictated by proxmox.                                                                                                                                                                                     |
| `iops_wr_concurrent`   |  `int`  |      `0`      |         `all`         | Maximum number of iops while writing concurrently. `0` means unlimited.                                                                                                                                                                                                                |
| `iothread`             | `bool`  |    `false`    |   `scsi`, `virtio`    | Whether to use iothreads for this drive. Only effective when the the emulated controller type (`scsihw` top level block argument) is `virtio-scsi-single`.                                                                                                                             |
| `mbps_r_burst`         | `float` |     `0.0`     |         `all`         | Maximum read speed in megabytes per second. `0` means unlimited.                                                                                                                                                                                                                       |
| `mbps_r_concurrent`    | `float` |     `0.0`     |         `all`         | Maximum read speed in megabytes per second. `0` means unlimited.                                                                                                                                                                                                                       |
| `mbps_wr_burst`        | `float` |     `0.0`     |         `all`         | Maximum write speed in megabytes per second. `0` means unlimited.                                                                                                                                                                                                                      |
| `mbps_wr_concurrent`   | `float` |     `0.0`     |         `all`         | Maximum throttled write pool in megabytes per second. `0` means unlimited.                                                                                                                                                                                                             |
| `readonly`             | `bool`  |    `false`    |   `scsi`, `virtio`    | Whether the drive should be readonly.                                                                                                                                                                                                                                                  |
| `replicate`            | `bool`  |    `false`    |         `all`         | Whether the drive should considered for replication jobs.                                                                                                                                                                                                                              |
| `serial`               |  `str`  |               |         `all`         | The serial number of the disk.                                                                                                                                                                                                                                                         |
| `size`                 | `string`|               |         `all`         | **Computed** Size of the disk, `K` for kibibytes, `M` for mebibytes, `G` for gibibytes, `T` for tibibytes.|

### EFI Disk Block

The `efidisk` block is used to configure the disk used for EFI data storage. There may only be one EFI disk block.
The EFI disk will be automatically pre-loaded with distribution-specific and Microsoft Standard Secure Boot keys.

```hcl
resource "proxmox_vm_qemu" "resource-name" {
  // ...

  efidisk {
    efitype = "4m"
    storage = "local-lvm"
  }
}
```

See the [docs about EFI disks](https://pve.proxmox.com/pve-docs/chapter-qm.html#qm_bios_and_uefi) for more details.

| Argument            | Type   | Default Value | Description                                                           |
| ------------------- | ------ | ------------- | --------------------------------------------------------------------- |
| `pre_enrolled_keys` | `bool` | `false`       | Whether or not to pre-enroll secure boot keys and thus enable secure boot |
| `efitype`           | `str`  | `"4m"`        | The type of efi disk device to add. Options: `2m`, `4m`               |
| `storage`           | `str`  |               | **Required** The name of the storage pool on which to store the disk. |

### PCI Block

The `pci` block is used to configure PCI devices. It may be specified multiple times.
Don't need it in a module? Use the [PCIs Block](#pcis-block) instead.

| Argument        | Type   | Default Value | Description |
| :-------------- | :----: | :-----------: | :---------- |
| `id`            | `str`  |               | **Required** The id of the PCI device. Range `0` - `15`. |
| `mapping_id`    | `str`  |               | **Required\*** The id of the mapping. Conflicts with `raw_id`.|
| `raw_id`        | `str`  |               | **Required\*** The id of the raw device. Conflicts with `mapping_id`.|
| `pcie`          | `bool` | `false`       | Whether this device is a `PCIe` device. |
| `primary_gpu`   | `bool` | `false`       | Whether this PCI device is the primary GPU. |
| `rombar`        | `bool` | `true`        | Whether to enable the ROM-BAR. |
| `device_id`     | `str`  |               | The device id of the PCI device. |
| `vendor_id`     | `str`  |               | The vendor id of the PCI device. |
| `sub_device_id` | `str`  |               | The sub device id of the PCI device. |
| `sub_vendor_id` | `str`  |               | The sub vendor id of the PCI device. |
| `mdev`          | `str`  |               | The mediated device. |

\* Either `mapping_id` or `raw_id` is required.

### PCIs Block

The `pcis` block is used to configure PCI devices.
There are two types of PCI devices `mapping`, and `raw`. Each of these types has their own block with their own arguments.

These types share the following arguments, with minor differences:

| Argument        | Type   | Default Value | PCI types        |Description |
| :-------------- | :----: | :-----------: | :--------------: | :--------- |
| `mapping_id`    | `str`  |               | `mapping`        | **Required** The id of the mapping. |
| `raw_id`        | `str`  |               | `raw`            | **Required** The id of the raw device. |
| `pcie`          | `bool` | `false`       | `mapping`, `raw` | Whether this device is a `PCIe` device. |
| `primary_gpu`   | `bool` | `false`       | `mapping`, `raw` | Whether this PCI device is the primary GPU. |
| `rombar`        | `bool` | `true`        | `mapping`, `raw` | Whether to enable the ROM-BAR. |
| `device_id`     | `str`  |               | `mapping`, `raw` | The device id of the PCI device. |
| `vendor_id`     | `str`  |               | `mapping`, `raw` | The vendor id of the PCI device. |
| `sub_device_id` | `str`  |               | `mapping`, `raw` | The sub device id of the PCI device. |
| `sub_vendor_id` | `str`  |               | `mapping`, `raw` | The sub vendor id of the PCI device. |
| `mdev`          | `str`  |               | `mapping`, `raw` | The mediated device. |

The range of pci devices is from `pci0` to `pci15`.

Example:

```hcl
resource "proxmox_vm_qemu" "resource-name" {
  // ...
  pcis {
    pci0 {
      mapping {
        mapping_id = "mapping-id"
        pcie = true
        primary_gpu = true
        rombar = true
        device_id = "device-id"
        vendor_id = "vendor-id"
        sub_device_id = "sub-device-id"
        sub_vendor_id = "sub-vendor-id"
      }
    }
    pci15 {
      raw {
        raw_id = "raw-id"
        pcie = true
        primary_gpu = true
        rombar = true
        device_id = "device-id"
        vendor_id = "vendor-id"
        sub_device_id = "sub-device-id"
        sub_vendor_id = "sub-vendor-id"
      }
    }
  }
}
```

### Serial Block

Create a serial device inside the VM (up to a maximum of 4 can be specified), and either pass through a host serial device (i.e. /dev/ttyS0), or create a unix socket on the host side. The order in which `serial` blocks are declared does not matter.

**WARNING**: Use with caution, as the docs indicate this device is experimental and users have reported issues with it.

See the [options for serial](https://pve.proxmox.com/pve-docs/chapter-qm.html#qm_options) in the PVE docs for more
details.

| Argument | Type  | Default Value | Description                                                                                                            |
| -------- | ----- | ------------- | ---------------------------------------------------------------------------------------------------------------------- |
| `id`     | `int` |               | **Required** The ID of the serial device. Must be unique, and between `0-3`.                                           |
| `type`   | `str` | `socket`      | The type of serial device to create. Options: `socket`, or the path to a serial device like `/dev/ttyS0`. |

### TPM Block

The `tpm_state` block is used to configure a TPM disk. It may only be specified once.

| Argument | Type | Default Value | Description |
| -------- | ---- | ------------- | ----------- |
| `storage`| `str`|               | **Required** The name of the storage backend on which to store the TPM disk.|
| `version`| `str`| `v2.0`        | The version of the TPM to use. Options: `v1.2`, `v2.0`.|

### USB Block

The `usb` block is used to configure USB devices. It may be specified multiple times. When no `device_id`, `mapping_id`, or `port_id` is specified, it will be a `spice` device.
In order to have a normal diff put the `usb` blocks in alphanumeric order based on the value of `id`.
Don't need it in a module? Use the [USBs](#usbs-block) instead.

See the [docs about USB passthrough](https://pve.proxmox.com/pve-docs/chapter-qm.html#qm_usb_passthrough) for more
details.

| Argument     | Type     | Default Value | Description |
| ------------ | -------- | ------------- | ----------- |
| `id`         | `int`    |               | **Required** The ID of the USB device. Must be unique, and between `0-4`. |
| `device_id`  | `string` |               | The USB device ID, mutually exclusive with `mapping_id` and `port_id`. |
| `mapping_id` | `string` |               | The USB mapping ID, mutually exclusive with `device_id` and `port_id`. |
| `port_id`    | `string` |               | The USB port ID, mutually exclusive with `device_id` and `mapping_id`. |
| `usb3`       | `bool`   | `false`       | Specifies whether the USB device or port is USB3. |

### USBs Block

The `usbs` block is used to configure USB devices.
There are four types of USB devices `device`, `mapping`, `port`, and `spice`. Configuration for these sub types can be found in their respective chapters:

* `device`: [USBs.x.Device Block](#usbsxdevice-block).
* `mapping`: [USBs.x.Mapping Block](#usbsxmapping-block).
* `port`: [USBs.x.Port Block](#usbsxport-block).
* `spice`: [USBs.x.Spice Block](#usbsxspice-block).

```hcl
resource "proxmox_vm_qemu" "resource-name" {
  //<arguments omitted for brevity...>

  usbs {
    usb0 {
      device {
        device_id = "e0bc:40a9"
        usb3 = true
      }
    }
    usb1 {
      mapping {
        mapping_id = "mapped-device"
        usb3 = true
      }
    }
    usb2 {
      port {
        port_id = "1-1"
        usb3 = true
      }
    }
    usb4 {
      spice {
        usb3 = true
      }
    }
  }
}
```

### USBs.x.Device Block

| Argument    | Type     | Default Value | Description |
| ----------- | -------- | ------------- | ----------- |
| `device_id` | `string` |               | **Required** The USB device ID, mutually exclusive with `mapping_id` and `port_id`. |
| `usb3`      | `bool`   | `false`       | Specifies whether the USB device or port is USB3. |

### USBs.x.Mapping Block

| Argument     | Type     | Default Value | Description |
| ------------ | -------- | ------------- | ----------- |
| `mapping_id` | `string` |               | **Required** The USB mapping ID, mutually exclusive with `device_id` and `port_id`. |
| `usb3`       | `bool`   | `false`       | Specifies whether the USB device or port is USB3. |

### USBs.x.Port Block

| Argument  | Type     | Default Value | Description |
| --------- | -------- | ------------- | ----------- |
| `port_id` | `string` |               | **Required** The USB port ID, mutually exclusive with `device_id` and `mapping_id`. |
| `usb3`    | `bool`   | `false`       | Specifies whether the USB device or port is USB3. |

### USBs.x.Spice Block

| Argument | Type   | Default Value | Description |
| -------- | ------ | ------------- | ----------- |
| `usb3`   | `bool` | `false`       | Specifies whether the USB device or port is USB3. |

### RNG Block

The `rng` block is used to configure a random number generator device. It can only be specified once.

| Argument | Type     | Default Value | Description |
| -------- | -------- | ------------- | ----------- |
| `limit`  | `int`    | `1024`        | The maximum number of bytes per `period` to read from the RNG device.|
| `period` | `int`    |               | The period in milliseconds to read from the RNG device. `0` for unlimited.|
| `source` | `string` | `/dev/urandom`| The source of the random number generator. Options: `/dev/random`, `/dev/urandom`, `/dev/hwrng`. |

### Startup and Shutdown Reference

The `startup_shutdown` field is used to configure the startup and shutdown settings. It may ony be specified once.

| Argument            | Type | Default Value | Description |
|:--------------------|------|---------------|:------------|
| `order`             | `int`| `-1`          | Startup order `-1` means any.|
| `shutdown_timeout`  | `int`| `-1`          | Shutdown timeout in seconds, `-1` means default.|
| `startup_delay`     | `int`| `-1`          | Startup delay in seconds, `-1` means default.|


## SMBIOS Block

The `smbios` block sets SMBIOS type 1 settings for the VM.

| Argument       | Type     | Description               |
| -------------- | -------- | ------------------------- |
| `family`       | `string` | The SMBIOS family.        |
| `manufacturer` | `string` | The SMBIOS manufacturer.  |
| `serial`       | `string` | The SMBIOS serial number. |
| `product`      | `string` | The SMBIOS product.       |
| `sku`          | `string` | The SMBIOS SKU.           |
| `uuid`         | `string` | The SMBIOS UUID.          |
| `version`      | `string` | The SMBIOS version.       |

## Attribute Reference

In addition to the arguments above, the following attributes can be referenced from this resource.

| Attribute              | Type  | Description                                                                                                                                                                                                                                                                                                                                                                      |
| ---------------------- | ----- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `ssh_host`             | `str` | Read-only attribute. Only applies when `define_connection_info` is true. The hostname or IP to use to connect to the VM for preprovisioning. This can be overridden by defining `ssh_forward_ip`, but if you're using cloud-init and `ipconfig0=dhcp`, the IP reported by qemu-guest-agent is used, otherwise the IP defined in `ipconfig0` is used.                             |
| `ssh_port`             | `str` | Read-only attribute. Only applies when `define_connection_info` is true. The port to connect to the VM over SSH for preprovisioning. If using cloud-init and a port is not specified in `ssh_forward_ip`, then 22 is used. If not using cloud-init, a port on the `target_node` will be forwarded to port 22 in the guest, and this attribute will be set to the forwarded port. |
| `default_ipv4_address` | `str` | Read-only attribute. Only applies when `agent` is `1` and Proxmox can actually read the ip the vm has. The settings `ipconfig0` and `skip_ipv4` have influence on this.|
| `default_ipv6_address` | `str` | Read-only attribute. Only applies when `agent` is `1` and Proxmox can actually read the ip the vm has. The settings `ipconfig0` and `skip_ipv6` have influence on this.|

## Import

A VM Qemu Resource can be imported using its node, type and VM ID i.e.:

```bash
 terraform [global options] import [options] ADDRESS <node>/<type>/<vmId>
```

`ADDRESS` must correspond to a resource block. `<type>` will always be `qemu` for VMs.

#### Example

> Creating a VM via the Proxmox GUI Wizard and importing it to Terraform to understand how different options maps to Terraform config code

1. Create a dummy file, e.g. `test.tf`, containing a dummy resource block `resource "proxmox_vm_qemu" "import_test" { }`
2. Run the Terraform import `terraform import proxmox_vm_qemu.import_test mynode/qemu/106`
3. The state gets imported to your `terraform.tfstate`, you can open that file and explore the imported state as Terraform config code
