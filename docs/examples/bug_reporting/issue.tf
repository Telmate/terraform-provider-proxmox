resource "proxmox_vm_qemu" "server" {
  name              = var.name
  target_node       = var.target_node
  clone             = "debian-12"
  os_type           = "cloud-init"
  scsihw            = "virtio-scsi-pci"
# Pre-3.0 format for disks
#  disk {
#    size            = "20G"
#    type            = "virtio"
#    cache           = "writeback"
#    storage         = var.storage_backend
#  }
# New & improved disk format (3.0 and later)
  disks {
    virtio {
      virtio0 {
        disk {
          size            = 20
          cache           = "writeback"
          storage         = var.storage_backend
        }
      }
    }
  }
  network {
    model           = "virtio"
    bridge          = "vmbr0"
  }

  # Cloud Init Settings
  # Reference: https://pve.proxmox.com/wiki/Cloud-Init_Support
  cloudinit_cdrom_storage = var.storage_backend
  boot = "order=virtio0;ide3"
  ipconfig0 = "ip=${var.ip_address}/${var.cidr},gw=${var.gateway}"
  nameserver = var.nameservers
  # If your SSH public key is named differently, change the path below
  sshkeys = file("${path.root}/test.pub")
}
