# VM Qemu Resource

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

The following arguments are supported in the resource block:

* `name` - (Required) Name of the VM
* `target_node` - (Required) Node to place the VM on
* `vmid` - (Optional; integer) ID of the VM in Proxmox, defaults to 0 which indicates it should use the next number in the sequence.
* `desc` - (Optional) Description of the VM
* `define_connection_info` - (Optional; defaults to true) define the (SSH) connection parameters for preprovisioners, see config block below.
* `bios` - (Optional; defaults to seabios)
* `onboot` - (Optional)
* `boot` - (Optional; defaults to cdn)
* `bootdisk` - (Optional; defaults to true)
* `agent` - (Optional; defaults to 0)
* `iso` - (Optional)
* `clone` - (Optional) - The name of the VM to clone into a new VM
* `full_clone` - (Optional)
* `hastate` - (Optional) 
* `qemu_os` - (Optional; defaults to l26)
* `memory` - (Optional; defaults to 512)
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
* `clone_wait` - (Optional)
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
