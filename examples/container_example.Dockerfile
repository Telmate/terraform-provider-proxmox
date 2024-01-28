## Create the Dockerfile (copy of this file) at the root of your terraform code.
## You can pass in the ARM_ACCESS_KEY as an environment variable and reference it like the variables below in a Makefile.

## Get the latest version of the terraform-provider-proxmox-azrm which has the Azure resource manager built in.
FROM docker.io/clincha/terraform-provider-proxmox-azrm:1.0.12

## If you have additional dependencies, you can add them to the Dockerfile.
# RUN apk add py3-pip

## Copy your terraform code into the container
COPY . .

### To build and run the container, you can use the following commands:

## Build this container image and tag it
# podman build . --file container_example.Dockerfile --tag docker.io/clincha/my-infrastructure:1.0.0

## terraform plan
# podman run --entrypoint sh --env="TF_VAR*" --env="ARM_ACCESS_KEY=${ARM_ACCESS_KEY}" docker.io/clincha/my-infrastructure:1.0.0 -c "terraform init && terraform plan"

## terraform apply
# podman run --entrypoint sh --env="TF_VAR*" --env="ARM_ACCESS_KEY=${ARM_ACCESS_KEY}" docker.io/clincha/my-infrastructure:1.0.0 -c "terraform init && terraform apply -auto-approve"
