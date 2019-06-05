provider "proxmox" {
    pm_tls_insecure = true
    pm_api_url = "https://proxmox.org/api2/json"
    pm_password = "supersecret"
    pm_user = "terraform-user@pve"
}

resource "proxmox_lxc" "lxc-test" {
    hostname = "terraform-new-container"
    ostemplate = "shared:vztmpl/centos-7-default_20171212_amd64.tar.xz"
    target_node = "node-01"
    network = {
        id = 0
        name = "eth0"
        bridge = "vmbr0"
        ip = "dhcp"
        ip6 = "dhcp"
    }
    storage = "local-lvm"
    pool = "terraform"
    password = "rootroot"
    force = true
}
