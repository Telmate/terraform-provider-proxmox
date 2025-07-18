# You will need to customize these variables to suit your environment
variable "target_node" { default = "larry" }
variable "ip_address" { default = "192.168.73.57" }
variable "cidr" { default = "24" }
variable "gateway" { default = "192.168.73.1" }
variable "nameservers" { default = "192.168.73.100 192.168.73.200" }
variable "name" { default = "test.example.com" }
variable "storage_backend" { default = "local-zfs" }
