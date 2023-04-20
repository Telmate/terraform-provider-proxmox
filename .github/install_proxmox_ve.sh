#!/bin/bash
packages_installation () {
    apt install wget sudo git lsb-core curl software-properties-common -y
}
vagrant_installation () {
    wget -O- https://apt.releases.hashicorp.com/gpg | sudo gpg --dearmor -o /usr/share/keyrings/hashicorp-archive-keyring.gpg
    echo "deb [signed-by=/usr/share/keyrings/hashicorp-archive-keyring.gpg] https://apt.releases.hashicorp.com $(lsb_release -cs) main" | sudo tee /etc/apt/sources.list.d/hashicorp.list
    sudo apt update && sudo apt install vagrant
}

virtualbox_installation () {
    sudo apt-get install virtualbox -y
    sudo apt-get install virtualbox—ext–pack -y
}

packer_installation () {
    curl -fsSL https://apt.releases.hashicorp.com/gpg | sudo apt-key add -
    sudo apt-add-repository "deb [arch=amd64] https://apt.releases.hashicorp.com $(lsb_release -cs) main" -y
    sudo apt-get install packer -y
}

proxmox_installation () {
    git clone https://github.com/rgl/proxmox-ve
    cd proxmox-ve
    make build-virtualbox
    vagrant box add -f proxmox-ve-amd64 proxmox-ve-amd64-virtualbox.box
    cd example
    sed -i 's/10.10.10.2/<NEW_ADRESS>/g' Vagrantfile
    vagrant up --no-destroy-on-error --provider=virtualbox
}

packages_installation
vagrant_installation
virtualbox_installation
packer_installation
proxmox_installation