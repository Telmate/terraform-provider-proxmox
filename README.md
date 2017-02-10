# Proxmox  4 Terraform

Terraform provider plugin for proxmox


## Work in progress

### TODO

* document terraform-ubuntu1404-template creation process
* implement pre-provision phase

## Build

Requires https://github.com/Telmate/proxmox-api-go

```
go build -o terraform-provider-proxmox
cp terraform-provider-proxmox $GOPATH/bin
```

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
	ssh_forward_ip = "10.0.0.1"
	clone = "terraform-ubuntu1404-template"
	storage = "local"
	cores = 3
	sockets = 1
	memory = 2560
	disk_gb = 4
	nic = "virtio"
	bridge = "vmbr1"
	os_type = "ubnutu"
	os_network_config = <<EOF
auto eth0
iface eth0 inet dhcp
EOF
}

```


