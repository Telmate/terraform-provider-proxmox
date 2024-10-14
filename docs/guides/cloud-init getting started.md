# Cloud-Init Getting Started

This guide will help you get started with Cloud-Init on Proxmox Virtual Environment `PVE`. Cloud Init is a multi-distribution package that handles early initialization of a virtual machine. It is used for configuring the hostname, setting up SSH keys, and other tasks that need to be done before the virtual machine is ready for use.

Note: **all command are performed from the PVE shell**.

## Creating a Cloud Init Template

Before you can use Cloud-Init, you need to create a template that will be used to clone new virtual machines. This template will have the Cloud-Init package installed and configured. The following steps will guide you through creating a Cloud Init template:

### Downloading a Cloud-Init Image

For this guide, we will use the Debian 12 Cloud-Init image. You can download the image from the following link:

```bash
wget https://cloud.debian.org/images/cloud/bookworm/latest/debian-12-genericcloud-amd64.qcow2
```

### Importing the Cloud-Init Image

Before we can import the Cloud-Init image, we need to create a VM to give the image to. The following command will create a new VM with the ID `9000`:

```bash
qm create 9000 --name debian12-cloudinit
```

Note: **Terraform is meant to manage the full life cycle of the VM, therefore we won't make any further changes to the VM**.

Once the VM is created, we can import the Cloud-Init image using the following command:

```bash
qm set 9000 --scsi0 local-lvm:0,import-from=/root/debian-12-genericcloud-amd64.qcow2
```

### Creating a Template from the VM

Now that we have the Cloud-Init image imported, we can create a template from the VM. The following command will convert the VM with ID `9000` to a template:

```bash
qm template 9000
```

## Creating a Snippet

Snippets are used to pass additional configuration to the Cloud-Init package. For this guide we will create a snippet that ensures the `qemu-guest-agent` package is installed on the virtual machine. Before we can create a snippet, we need to create a place to store it. Preferably in the same storage as the template. Do keep in mind that the cloned VMs can't start if the snippet is not accessible. Throughout this guide we will use the `local` storage.

```bash
mkdir /var/lib/vz/snippets
```

Now that we have a place to store the snippet, we can create the snippet itself. The following command will create a snippet that installs the `qemu-guest-agent.yml` package:

```bash
tee /var/lib/vz/snippets/qemu-guest-agent.yml <<EOF
#cloud-config
runcmd:
  - apt update
  - apt install -y qemu-guest-agent
  - systemctl start qemu-guest-agent
EOF
```

## Terraform Configuration

Now that we have a Cloud-Init template and a snippet, we can use Terraform to create a new VM from the template. The following Terraform configuration will create a new VM with the ID `100`:

```hcl
resource "proxmox_vm_qemu" "cloudinit-example" {
  vmid        = 100
  name        = "test-terraform0"
  target_node = "pve"
  agent       = 1
  cores       = 2
  memory      = 1024
  boot        = "order=scsi0" # has to be the same as the OS disk of the template
  clone       = "debian12-cloudinit" # The name of the template
  scsihw      = "virtio-scsi-single"
  vm_state    = "running"
  automatic_reboot = true

  # Cloud-Init configuration
  cicustom   = "vendor=local:snippets/qemu-guest-agent.yml" # /var/lib/vz/snippets/qemu-guest-agent.yml
  ciupgrade  = true
  nameserver = "1.1.1.1 8.8.8.8"
  ipconfig0  = "ip=192.168.1.10/24,gw=192.168.1.1,ip6=dhcp"
  skip_ipv6  = true
  ciuser     = "root"
  cipassword = "Enter123!"
  sshkeys    = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIE/Pjg7YXZ8Yau9heCc4YWxFlzhThnI+IhUx2hLJRxYE Cloud-Init@Terraform"

  # Most cloud-init images require a serial device for their display
  serial {
    id = 0
  }

  disks {
    scsi {
      scsi0 {
        # We have to specify the disk from our template, else Terraform will think it's not supposed to be there
        disk {
          storage = "local-lvm"
          # The size of the disk should be at least as big as the disk in the template. If it's smaller, the disk will be recreated
          size    = "2G" 
        }
      }
    }
    ide {
      # Some images require a cloud-init disk on the IDE controller, others on the SCSI or SATA controller
      ide1 {
        cloudinit {
          storage = "local-lvm"
        }
      }
    }
  }

  network {
    bridge = "vmbr0"
    model  = "virtio"
  }
}

terraform {
  required_providers {
    proxmox = {
      source = "Telmate/proxmox"
      version = ">=3.0.1rc4"
    }
  }
}
```
