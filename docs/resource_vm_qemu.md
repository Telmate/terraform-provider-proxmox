# Terraform VM Qemu Resource

Resources are the most important element in the Terraform language. Each resource block describes one or more 
infrastructure objects, such as virtual networks, compute instances, or higher-level components such as DNS records.

## Create a Qemu VM resource

You can start from either an ISO or clone an existing VM. Optimally, you could create a VM resource you will use a clone 
base with an ISO, and make the rest of the VM resources depend on that base "template" and clone it.

## Preprovision

With preprovision you can provision a VM directly from the resource block.

```tf
resource "proxmox_vm_qemu" "prepprovision-test" {
    ...
    preprovision = true
    os_type = "ubuntu"  // ubuntu, centos or cloud-init
}
```

### Preprovision for Linux (Ubuntu / CentOS)

There is a pre-provision phase which is used to set a hostname, intialize eth0, and resize the VM disk to available 
space. This is done over SSH with the ssh_forward_ip, ssh_user and ssh_private_key. Disk resize is done if the file 
[/etc/auto_resize_vda.sh](https://github.com/Telmate/terraform-ubuntu-proxmox-iso/blob/master/auto_resize_vda.sh) exists.


## Preprovision for Cloud-Init

Cloud-init VMs must be cloned from a [cloud-init ready template](https://pve.proxmox.com/wiki/Cloud-Init_Support).

* ciuser - User name to change ssh keys and password for instead of the image’s configured default user.
* cipassword - Password to assign the user.
* cicustom - location of cloud-config files that Proxmox will put in the generated cloud-init config iso image.
* searchdomain - Sets DNS search domains for a container.
* nameserver - Sets DNS server IP address for a container.
* sshkeys - public ssh keys, one per line
* ipconfig0 - [gw=<GatewayIPv4>] [,gw6=<GatewayIPv6>] [,ip=<IPv4Format/CIDR>] [,ip6=<IPv6Format/CIDR>]
* ipconfig1 - optional, same as ipconfig0 format


## Argument reference

* `cicustom` - (Optional) Location of cloud-config files that Proxmox will put in the generated cloud-init config iso 
  image, e.g. `cicustom = "user=local:snippets/user_data.yaml"`. For more info about this attribute, see the details of 
  the parameter `--cicustom` in the section "Custom Cloud-Init Configuration" from the [Proxmox Cloud-Init support](https://pve.proxmox.com/wiki/Cloud-Init_Support) page.
* `ssh_forward_ip` - (Optional) IP or hostname of the target node or bridge IP. This is where proxmox will create a port
  forward to your VM with via a user_net. (for pre-cloud-init provisioning).
