# Proxmox  4 Terraform

Terraform provider plugin for proxmox


## Working prototype

## Build

Requires https://github.com/Telmate/proxmox-api-go

```
go build -o terraform-provider-proxmox
cp terraform-provider-proxmox $GOPATH/bin
cp terraform-provider-proxmox $GOPATH/bin/terraform-provisioner-proxmox
```

Note: this plugin is both a provider and provisioner in one, which is why it needs to be in the $GOPATH/bin/ twice.

Recommended ISO builder https://github.com/Telmate/terraform-ubuntu-proxmox-iso


## Run

```
terraform apply
```

### Sample file

main.tf:
```
provider "proxmox" {
}

resource "proxmox_vm_qemu" "test" {
	name = "tftest1.xyz.com"
	desc = "tf description"
	target_node = "proxmox1-xx"

	clone = "terraform-ubuntu1404-template"
	storage = "local"
	cores = 3
	sockets = 1
	memory = 2560
	disk_gb = 4
	nic = "virtio"
	bridge = "vmbr1"
	ssh_forward_ip = "10.0.0.1"
	ssh_user = "terraform"
	ssh_private_key = <<EOF
-----BEGIN RSA PRIVATE KEY-----
private ssh key terraform
-----END RSA PRIVATE KEY-----
EOF

	os_type = "ubuntu"
	os_network_config = <<EOF
auto eth0
iface eth0 inet dhcp
EOF

	provisioner "remote-exec" {
		inline = [
			"ip a"
		]
	}

	provisioner "proxmox" {
		action = "sshbackward"
	}
}

```
### Provider usage
You can start from either an ISO or clone an existing VM.

Optimally, you could create a VM resource you will use a clone base with an ISO, and make the rest of the VM resources depend on that base "template" and clone it.

Interesting parameters:

**ssh_forward_ip** - should be the IP or hostname of the target node or bridge IP. This is where proxmox will create a port forward to your VM with via a user_net.

**os_type** - ubuntu (https://github.com/Telmate/terraform-ubuntu-proxmox-iso) or centos (TODO: centos iso template)


### Preprovision (internal)

There is a pre-provision phase which is used to set a hostname, intialize eth0, and resize the VM disk to available space. This is done over SSH with the ssh_forward_ip, ssh_user and ssh_private_key.

Disk resize is done if the file /etc/auto_resize_vda.sh exists. Source: https://github.com/Telmate/terraform-ubuntu-proxmox-iso/blob/master/auto_resize_vda.sh

### Provisioner usage


Remove the temporary net1 adapter.
Inside the VM this usually triggers the routes back to the provisioning machine on net0.
```
	provisioner "proxmox" {
		action = "sshbackward"
	}

```

Replace the temporary net1 adapter with a new persistent net1.
```
	provisioner "proxmox" {
		action = "reconnect"
		net1 = "virtio,bridge=vmbr0,tag=99"
	}

```
If net1 needs a config other than DHCP you should prior to this use provisioner "remote-exec" to modify the network config.
