provider "proxmox" {
    pm_tls_insecure = true
    pm_api_url = "https://proxmox-server01.example.com:8006/api2/json"
    pm_password = "secret"
    pm_user = "terraform-user@pve"
}

resource "proxmox_vm_qemu" "cloudinit-test" {
    name = "terraform-test-vm"
    desc = "A test for using terraform and cloudinit"

    # Node name has to be the same name as within the cluster
    # this might not include the FQDN
    target_node = "proxmox-server02"

    # The destination resource pool for the new VM
    pool = "pool0"

    # The template name to clone this vm from
    clone = "linux-cloudinit-template"

    # Activate QEMU agent for this VM
    agent = 1

    os_type = "cloud-init"
    cores = "2"
    sockets = "1"
    memory = "2048"

    # Setup the disk. The id has to be unique
    disk {
        id = 0
        size = 32
        type = "virtio"
        storage = "ceph-storage-pool"
        storage_type = "rbd"
        iothread = true
    }

    # Setup the network interface and assign a vlan tag: 256
    network {
        id = 0
        model = "virtio"
        bridge = "vmbr0"
        tag = 256
    }

    # Setup the ip address using cloud-init.
    # Keep in mind to use the CIDR notation for the ip.
    ipconfig0 = "ip=192.168.10.20/24,gw=192.168.10.1"

    sshkeys = <<EOF
    ssh-rsa 9182739187293817293817293871== user@pc
    EOF
}
