terraform {
  required_version = ">= 1.1.0"
  required_providers {
    proxmox = {
      source  = "telmate/proxmox"
      version = ">= 2.9.5"
    }
  }
}

provider "proxmox" {
    pm_tls_insecure = true
    pm_api_url = "https://proxmox01.example.com:8006/api2/json"
    pm_password = "password"
    pm_user = "root@pam"
    pm_otp = ""
}

resource "proxmox_vm_qemu" "pxe-example" {
    name                      = "pxe-example"
    desc                      = "A test VM for PXE boot mode."
# PXE option enables the network boot feature
    pxe                       = true
# unless your PXE installed system includes the Agent in the installed
# OS, do not use this, especially for PXE boot VMs
    agent                     = 0
    automatic_reboot          = true
    balloon                   = 0
    bios                      = "seabios"
# boot order MUST include network, this is enforced in the Provider
# Optinally, setting a disk first means that PXE will be used first boot
# and future boots will run off the disk
    boot                      = "order=scsi0;net0"
    cores                     = 2
    cpu                       = "host"
    define_connection_info    = true
    force_create              = false
    hotplug                   = "network,disk,usb"
    kvm                       = true
    memory                    = 2048
    numa                      = false
    onboot                    = false
    vm_state                  = "running"
    os_type                   = "Linux 5.x - 2.6 Kernel"
    qemu_os                   = "l26"
    scsihw                    = "virtio-scsi-pci"
    sockets                   = 1
    protection                = false
    tablet                    = true
    target_node               = "test"
    vcpus                     = 0

    disks {
        scsi {
            scsi0 {
                disk {
                    backup             = true
                    cache              = "none"
                    discard            = true
                    emulatessd         = true
                    iothread           = true
                    mbps_r_burst       = 0.0
                    mbps_r_concurrent  = 0.0
                    mbps_wr_burst      = 0.0
                    mbps_wr_concurrent = 0.0
                    replicate          = true
                    size               = 32
                    storage            = "local-lvm"
                }
            }
        }
    }

    network {
        bridge    = "vmbr0"
        firewall  = false
        link_down = false
        model     = "e1000"
    }

    smbios {
        family       = "VM"
        manufacturer = "Hashibrown"
        product      = "Terraform"
        sku          = "dQw4w9WgXcQ"
        uuid         = "5b710d2f-4ea2-4d49-9eaa-c18392fd734d"
        version      = "v1.0"
        serial       = "ABC123"
    }
}
