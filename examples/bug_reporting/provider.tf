# Define the usual provider things
terraform {
  required_providers {
    proxmox = {
      # Use these two lines to run a locally compiled version of the provider
      #source = "registry.example.com/telmate/proxmox"
      #version = ">=1.0.0"

      # Normally we want the provider from the registry
      source = "Telmate/proxmox"
      # We can specificy a specific version. If we do not, terraform will use
      # the latest official release.
      #version = "=2.9.11"
      #version = "=3.0.1-rc1"
    }
  }
  required_version = ">= 0.14"
}

# To enable debugging, uncomment the following block. This will create a log
# file that is helpful in understanding what happened and getting help from
# others.
#provider "proxmox" {
#  pm_log_enable = true
#  pm_log_file   = "terraform-plugin-proxmox.log"
#  pm_debug      = true
#  pm_log_levels = {
#    _default    = "debug"
#    _capturelog = ""
#  }
#}

