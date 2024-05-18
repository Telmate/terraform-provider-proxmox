# Cloud Init Guide

Proxmox has support for Cloud-Init, which allows changing settings in the guest when deploying. This is important
because you'll want to make sure the settings in your VM do not match the base image, or you'll have IP conflicts,
duplicate SSH host keys, SSH keys in authorized_keys files that you probably don't want in there and so forth.

Cloud-Init has many ways to get the configuration information to the guest, and two of them are supported by
Proxmox: [NoCloud (v1), and ConfigDrive
(v2)](https://pve.proxmox.com/wiki/Cloud-Init_FAQ#Step_3:_Install_and_configure_cloud-init). According to
[the documentation](https://pve.proxmox.com/wiki/Cloud-Init_Support#_cloud_init_specific_options), NoCloud is used for
Linux and ConfigDrive is used for Windows. However, in the
[FAQ](https://pve.proxmox.com/wiki/Cloud-Init_FAQ#Usage_in_Proxmox_VE), it mentions that ConfigDrive is not officially
supported for Windows.

Both of these use a special CD-ROM to make the Cloud-Init configuration values available to the guest. It is important
to note that this is not the same as the normal CD-ROM that is added to VMs, but rather one that will automatically
generate an ISO9660 image with the Cloud-Init settings when necessary. It can
be [added manually](https://pve.proxmox.com/wiki/Cloud-Init_FAQ#Usage_in_Proxmox_VE)
or via packer with `"cloud_init": true`.

It's also important that your base image has support for Cloud-Init as well, and be configured to use with either
NoCloud or ConfigDrive. Proxmox has documentation in the FAQ about
[creating a custom cloud image](https://pve.proxmox.com/wiki/Cloud-Init_FAQ#Creating_a_custom_cloud_image)
which should suffice.

It's important to note that you must not reboot the base image after installing and setting up the Cloud-Init. The first
boot will trigger a bunch of "first boot" configuration actions to take place. One of the standard ones is regenerating
SSH host keys. This means if you install the package with a pre-seed file, and then run a provisioner, you'll have
triggered these steps, which means they will not be triggered after the VM is cloned.

After the base image is set up, the deployment process goes like this:

1. Clone the base image
2. Start the cloned image
3. When the clone boots, the Cloud-Init service will search for any sources that should be used to configure the machine

When using the Terraform provider for Proxmox, you do not need to create a configuration file. All you need to do is
specify the settings that you want to pass into the guest. The most common one will be ipconfig0 to configure the first
network interface, but there are
[more listed in the Proxmox docs](https://pve.proxmox.com/wiki/Cloud-Init_Support#_cloud_init_specific_options).
There is an example of this in [examples/cloudinit_example.tf](../../examples/cloudinit_example.tf).

Now, there is one other way to get the configuration data into the guest without using this magical CloudInit CD-ROM and
that's by using cicustom. This allows you to create a NoCloud (v1) or ConfigDrive (v2) configuration file instead of
using the one that will be automatically generated for you. The example below shows how to use this. For help writing
the config file, see the
[NoCloud](https://cloudinit.readthedocs.io/en/latest/topics/datasources/nocloud.html)
or
[ConfigDrive](https://cloudinit.readthedocs.io/en/latest/topics/datasources/configdrive.html)
docs.

## Sample file

main.tf:

```hcl
/* Uses Cloud-Init options from Proxmox 5.2 */
resource "proxmox_vm_qemu" "cloudinit-test" {
  name        = "tftest1.xyz.com"
  desc        = "tf description"
  target_node = "proxmox1-xx"

  clone = "ci-ubuntu-template"

  # The destination resource pool for the new VM
  pool = "pool0"

  storage = "local"
  cores   = 3
  sockets = 1
  memory  = 2560
  disk_gb = 4
  nic     = "virtio"
  bridge  = "vmbr0"

  ssh_user        = "root"
  ssh_private_key = <<EOF
-----BEGIN RSA PRIVATE KEY-----
private ssh key root
-----END RSA PRIVATE KEY-----
EOF

  os_type   = "cloud-init"
  ipconfig0 = "ip=10.0.2.99/16,gw=10.0.2.2"

  sshkeys = <<EOF
ssh-rsa AABB3NzaC1kj...key1
ssh-rsa AABB3NzaC1kj...key2
EOF

  provisioner "remote-exec" {
    inline = [
      "ip a"
    ]
  }
}

# Modify path for templatefile and use the recommended extension of .tftpl for syntax hylighting in code editors.
resource "local_file" "cloud_init_user_data_file" {
  count    = var.vm_count
  content  = templatefile("${var.working_directory}/cloud-inits/cloud-init.cloud_config.tftpl", { ssh_key = var.ssh_public_key, hostname = var.name })
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

/* Configure Cloud-Init User-Data with custom config file */
resource "proxmox_vm_qemu" "cloudinit-test" {
  depends_on = [
    null_resource.cloud_init_config_files,
  ]

  name        = "tftest1.xyz.com"
  desc        = "tf description"
  target_node = "proxmox1-xx"

  clone = "ci-ubuntu-template"

  # The destination resource pool for the new VM
  pool = "pool0"

  storage = "local"
  cores   = 3
  sockets = 1
  memory  = 2560
  disk_gb = 4
  nic     = "virtio"
  bridge  = "vmbr0"

  ssh_user        = "root"
  ssh_private_key = <<EOF
-----BEGIN RSA PRIVATE KEY-----
private ssh key root
-----END RSA PRIVATE KEY-----
EOF

  os_type   = "cloud-init"
  ipconfig0 = "ip=10.0.2.99/16,gw=10.0.2.2"

  /*
    sshkeys and other User-Data parameters are specified with a custom config file.
    In this example each VM has its own config file, previously generated and uploaded to
    the snippets folder in the local storage in the Proxmox VE server.
  */
  cicustom                = "user=local:snippets/user_data_vm-${count.index}.yml"
  /* Create the Cloud-Init drive on the "local-lvm" storage */
  disks {
    ide {
      ide3 {
        cloudinit {
          storage = "local-lvm"
        }
      }
    }
  }

  provisioner "remote-exec" {
    inline = [
      "ip a"
    ]
  }
}

/* Uses custom eth1 user-net SSH portforward */
resource "proxmox_vm_qemu" "preprovision-test" {
  name        = "tftest1.xyz.com"
  desc        = "tf description"
  target_node = "proxmox1-xx"

  clone = "terraform-ubuntu1404-template"

  # The destination resource pool for the new VM
  pool = "pool0"

  cores    = 3
  sockets  = 1
  # Same CPU as the Physical host, possible to add cpu flags
  # Ex: "host,flags=+md-clear;+pcid;+spec-ctrl;+ssbd;+pdpe1gb"
  cpu      = "host"
  numa     = false
  memory   = 2560
  scsihw   = "lsi"
  # Boot from hard disk (c), CD-ROM (d), network (n)
  boot     = "cdn"
  # It's possible to add this type of material and use it directly
  # Possible values are: network,disk,cpu,memory,usb
  hotplug  = "network,disk,usb"
  # Default boot disk
  bootdisk = "virtio0"
  # HA, you need to use a shared disk for this feature (ex: rbd)
  hastate  = ""

  #Display
  vga {
    type   = "std"
    #Between 4 and 512, ignored if type is defined to serial
    memory = 4
  }

  network {
    id    = 0
    model = "virtio"
  }
  network {
    id     = 1
    model  = "virtio"
    bridge = "vmbr1"
  }
  disk {
    id           = 0
    type         = "virtio"
    storage      = "local-lvm"
    storage_type = "lvm"
    size         = "4G"
    backup       = true
  }
  # Serial interface of type socket is used by xterm.js
  # You will need to configure your guest system before being able to use it
  serial {
    id   = 0
    type = "socket"
  }
  preprovision    = true
  ssh_forward_ip  = "10.0.0.1"
  ssh_user        = "terraform"
  ssh_private_key = <<EOF
-----BEGIN RSA PRIVATE KEY-----
private ssh key terraform
-----END RSA PRIVATE KEY-----
EOF

  os_type           = "ubuntu"
  os_network_config = <<EOF
auto eth0
iface eth0 inet dhcp
EOF

  connection {
    type        = "ssh"
    user        = self.ssh_user
    private_key = self.ssh_private_key
    host        = self.ssh_host
    port        = self.ssh_port
  }

  provisioner "remote-exec" {
    inline = [
      "ip a"
    ]
  }
}

```
