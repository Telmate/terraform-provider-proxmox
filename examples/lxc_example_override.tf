provider "proxmox" {
    pm_api_url = "https://proxmox.wolke4.org/api2/json"
    pm_password = "^U)kV+^Yv}9_KiXN6QY,3;NZO"
    pm_user = "terraform@pve"
}

resource "proxmox_lxc" "lxc-test" {
    target_node = "wolke4"
}
