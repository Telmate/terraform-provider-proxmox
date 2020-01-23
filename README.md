[![Build Status](https://travis-ci.com/Telmate/terraform-provider-proxmox.svg?branch=master)](https://travis-ci.com/Telmate/terraform-provider-proxmox)

# Proxmox 4 Terraform

Terraform provider plugin for proxmox


## Working prototype


## Go Install

```
go install github.com/Telmate/terraform-provider-proxmox/cmd/terraform-provider-proxmox
go install github.com/Telmate/terraform-provider-proxmox/cmd/terraform-provisioner-proxmox
```
Note: this plugin is both a provider and provisioner in one, which is why it needs two install commands.

## Build local source

Requires https://github.com/Telmate/proxmox-api-go

```
make
make install
```

Recommended ISO builder https://github.com/Telmate/terraform-ubuntu-proxmox-iso

## Credentials

```bash
# Credentials and URL optionally defined in the environment
export PM_API_URL="https://xxxx.com:8006/api2/json"
export PM_USER=user@pam
export PM_PASS=password
```
If a 2FA OTP code is required
```bash
# Optional 2FA OTP code
export PM_OTP=otpcode
```

## Run

```
terraform init
terraform plan
terraform apply
```

### Sample file

main.tf:
```
provider "proxmox" {
  pm_tls_insecure = true
  /*
    // Credentials here or environment
    pm_api_url = "https://proxmox-server01.example.com:8006/api2/json"
    pm_password = "secret"
    pm_user = "terraform-user@pve"
    //Optional
    pm_otp = "otpcode"
  */
}

/* Uses cloud-init options from Proxmox 5.2 */
resource "proxmox_vm_qemu" "cloudinit-test" {
  name = "tftest1.xyz.com"
  desc = "tf description"
  target_node = "proxmox1-xx"

  clone = "ci-ubuntu-template"

  # The destination resource pool for the new VM
  pool = "pool0"

  storage = "local"
  cores = 3
  sockets = 1
  memory = 2560
  disk_gb = 4
  nic = "virtio"
  bridge = "vmbr0"

  ssh_user = "root"
  ssh_private_key = <<EOF
-----BEGIN RSA PRIVATE KEY-----
private ssh key root
-----END RSA PRIVATE KEY-----
EOF

  os_type = "cloud-init"
  ipconfig0 = "ip=10.0.2.99/16,gw=10.0.2.2"

  sshkeys = <<EOF
ssh-rsa AAAAB3NzaC1kj...key1
ssh-rsa AAAAB3NzaC1kj...key2
EOF

  provisioner "remote-exec" {
    inline = [
      "ip a"
    ]
  }
}

/* Null resource that generates a cloud-config file per vm */
data "template_file" "user_data" {
  count    = var.vm_count
  template = file("${path.module}/files/user_data.cfg")
  vars = {
    pubkey   = file("~/.ssh/id_rsa.pub")
    hostname = "vm-${count.index}"
    fqdn     = "vm-${count.index}.${var.domain_name}"
  }
}
resource "local_file" "cloud_init_user_data_file" {
  count    = var.vm_count
  content  = data.template_file.user_data[count.index].rendered
  filename = "${path.module}/files/user_data_${count.index}.cfg"
}

resource "null_resource" "cloud_init_config_files" {
  count = var.vm_count
  connection {
    type     = "ssh"
    user     = "${var.pve_user}"
    password = "${var.pve_password}"
    host     = "${var.pve_host}"
  }

  provisioner "file" {
    source      = local_file.cloud_init_user_data_file[count.index].filename
    destination = "/var/lib/vz/snippets/user_data_vm-${count.index}.yml"
  }
}

/* Configure cloud-init User-Data with custom config file */
resource "proxmox_vm_qemu" "cloudinit-test" {
  depends_on = [
    null_resource.cloud_init_config_files,
  ]

  name = "tftest1.xyz.com"
  desc = "tf description"
  target_node = "proxmox1-xx"

  clone = "ci-ubuntu-template"

  # The destination resource pool for the new VM
  pool = "pool0"

  storage = "local"
  cores = 3
  sockets = 1
  memory = 2560
  disk_gb = 4
  nic = "virtio"
  bridge = "vmbr0"

  ssh_user = "root"
  ssh_private_key = <<EOF
-----BEGIN RSA PRIVATE KEY-----
private ssh key root
-----END RSA PRIVATE KEY-----
EOF

  os_type = "cloud-init"
  ipconfig0 = "ip=10.0.2.99/16,gw=10.0.2.2"

  /*
    sshkeys and other User-Data parameters are specified with a custom config file.
    In this example each VM has its own config file, previously generated and uploaded to
    the snippets folder in the local storage in the Proxmox VE server.
  */
  cicustom = "user=local:snippets/user_data_vm-${count.index}.yml"

  provisioner "remote-exec" {
    inline = [
      "ip a"
    ]
  }
}

/* Uses custom eth1 user-net SSH portforward */
resource "proxmox_vm_qemu" "prepprovision-test" {
  name = "tftest1.xyz.com"
  desc = "tf description"
  target_node = "proxmox1-xx"

  clone = "terraform-ubuntu1404-template"

  # The destination resource pool for the new VM
  pool = "pool0"

  cores = 3
  sockets = 1
  # Same CPU as the Physical host, possible to add cpu flags
  # Ex: "host,flags=+md-clear;+pcid;+spec-ctrl;+ssbd;+pdpe1gb"
  cpu = "host"
  numa = false
  memory = 2560
  scsihw = "lsi"
  # Boot from hard disk (c), CD-ROM (d), network (n)
  boot = "cdn"
  # It's possible to add this type of material and use it directly
  # Possible values are: network,disk,cpu,memory,usb
  hotplug = "network,disk,usb"
  # Default boot disk
  bootdisk = "virtio0"
  # HA, you need to use a shared disk for this feature (ex: rbd)
  hastate = ""
  
  #Display
  vga {
    type = "std"
    #Between 4 and 512, ignored if type is defined to serial
    memory = 4
  }
  
  network {
    id = 0
    model = "virtio"
  }
  network {
    id = 1
    model = "virtio"
    bridge = "vmbr1"
  }
  disk {
    id = 0
    type = virtio
    storage = local-lvm
    storage_type = lvm
    size = 4G
    backup = true
  }
  # Serial interface of type socket is used by xterm.js
  # You will need to configure your guest system before being able to use it
  serial {
    id = 0
    type = "socket"
  }
  preprovision = true
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

  connection {
    type = "ssh"
    user = "${self.ssh_user}"
    private_key = "${self.ssh_private_key}"
    host = "${self.ssh_host}"
    port = "${self.ssh_port}"
  }

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
**preprovision** - to enable or disable internal pre-provisioning (e.g. if you already have another way to provision VMs). Conflicts with: `ssh_forward_ip`, `ssh_user`, `ssh_private_key`, `os_type`, `os_network_config`.
**os_type** -
* cloud-init  - from Proxmox 5.2
* ubuntu -(https://github.com/Telmate/terraform-ubuntu-proxmox-iso)
* centos - (TODO: centos iso template)

**ssh_forward_ip** - should be the IP or hostname of the target node or bridge IP. This is where proxmox will create a port forward to your VM with via a user_net. (for pre-cloud-init provisioning)

### Cloud-Init

Cloud-init VMs must be cloned from a cloud-init ready template.
See: https://pve.proxmox.com/wiki/Cloud-Init_Support

* ciuser - User name to change ssh keys and password for instead of the image’s configured default user.
* cipassword - Password to assign the user.
* searchdomain - Sets DNS search domains for a container.
* nameserver - Sets DNS server IP address for a container.
* sshkeys - public ssh keys, one per line
* ipconfig0 - [gw=<GatewayIPv4>] [,gw6=<GatewayIPv6>] [,ip=<IPv4Format/CIDR>] [,ip6=<IPv6Format/CIDR>]
* ipconfig1 - optional, same as ipconfig0 format

Alternatively, cloud-init configuration can be customized with config files that reside in a volume in the Proxmox VE server. Use the attribute `cicustom` to indicate the location of these files. 

* cicustom - location of cloud-config files that Proxmox will put in the generated cloud-init config iso image, e.g. `cicustom = "user=local:snippets/user_data.yaml"`. For more info about this attribute, see the details of the parameter `--cicustom` in the section "Custom Cloud-Init Configuration" from the [Proxmox Cloud-Init support](https://pve.proxmox.com/wiki/Cloud-Init_Support) page.

### Preprovision (internal alternative to Cloud-Init)

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
