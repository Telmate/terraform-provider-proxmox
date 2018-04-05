package proxmox

import (
	"fmt"
	"github.com/hashicorp/terraform/communicator"
	"github.com/hashicorp/terraform/communicator/remote"
	"github.com/hashicorp/terraform/helper/schema"
	// "github.com/hashicorp/terraform/terraform"
	// "github.com/mitchellh/go-linereader"
	"io"
	"log"
	"strconv"
	"strings"
	"time"
)

const eth0Payload = "echo $'%s' > /tmp/tf_eth0_payload"
const provisionPayload = "echo $'%s' > /tmp/tf_preprovision.sh"

// preprovision VM (setup eth0 and hostname)
const ubuntuPreprovisionScript = `
BOX_HOSTNAME=%s
BOX_SHORT_HOSTNAME=%s
SSH_CLIENT=$1
MY_IP=$(echo $SSH_CLIENT | awk "{ print \$1 }")
echo Using my ip $MY_IP to provision at $(date)
if [ -z "$(grep $BOX_SHORT_HOSTNAME /etc/hosts)" ]; then
	echo 127.0.1.1 $BOX_HOSTNAME $BOX_SHORT_HOSTNAME >> /etc/hosts
else
	echo Hosts file already set includes $BOX_SHORT_HOSTNAME
fi
echo $BOX_SHORT_HOSTNAME > /etc/hostname
hostname $BOX_SHORT_HOSTNAME
echo Hostname set $BOX_SHORT_HOSTNAME
if [ -z "$(grep eth0 /etc/network/interfaces)" ]; then
	echo Setting up eth0 for $BOX_HOSTNAME
	cat /tmp/tf_eth0_payload >> /etc/network/interfaces
else
	echo eth0 already setup for $BOX_HOSTNAME
fi

echo Attempting to bring up eth0
ip route add $MY_IP via 10.0.2.2
ip route del default via 10.0.2.2
ifup eth0
if [ -e /etc/auto_resize_vda.sh ]; then
	echo Auto-resizing file-system
	/etc/auto_resize_vda.sh
fi
echo Preprovision done at $(date)
`

// preprovision VM (setup eth0 and hostname)
const centosPreprovisionScript = `
BOX_HOSTNAME=%s
BOX_SHORT_HOSTNAME=%s
SSH_CLIENT=$1
MY_IP=$(echo $SSH_CLIENT | awk "{ print \$1 }")
echo Using my ip $MY_IP to provision at $(date)
if [ -z "$(grep $BOX_SHORT_HOSTNAME /etc/hosts)" ]; then
	echo 127.0.1.1 $BOX_HOSTNAME $BOX_SHORT_HOSTNAME >> /etc/hosts
else
	echo Hosts file already set includes $BOX_SHORT_HOSTNAME
fi
echo $BOX_SHORT_HOSTNAME > /etc/hostname
hostname $BOX_SHORT_HOSTNAME
echo Hostname set $BOX_SHORT_HOSTNAME
if [ -z "$(grep -i yes /etc/sysconfig/network-scripts/ifcfg-eth0)" ]; then
	echo Setting up eth0 for $BOX_HOSTNAME
	cat /tmp/tf_eth0_payload > /etc/sysconfig/network-scripts/ifcfg-eth0
else
	echo eth0 already setup for $BOX_HOSTNAME
fi

echo Attempting to bring up eth0
ip route add $MY_IP via 10.0.2.2
ip route del default via 10.0.2.2
ifup eth0
if [ -e /etc/auto_resize_vda.sh ]; then
	echo Auto-resizing file-system
	/etc/auto_resize_vda.sh
fi
echo Preprovision done at $(date)
`

func preProvisionCentos(d *schema.ResourceData) error {
	return preProvisionLinux(d, centosPreprovisionScript)
}

func preProvisionUbuntu(d *schema.ResourceData) error {
	return preProvisionLinux(d, ubuntuPreprovisionScript)
}

func preProvisionLinux(d *schema.ResourceData, provisionBash string) error {

	// Get a new communicator
	log.Print("[DEBUG] connecting to SSH on linux")
	comm, err := communicator.New(d.State())
	if err != nil {
		return err
	}

	log.Print("[DEBUG] sending os_network_config")

	// on this first one allow some retries to connect
	rr := 0
	for {
		rr++
		err = runCommand(comm, fmt.Sprintf(eth0Payload, strings.Trim(strconv.Quote(d.Get("os_network_config").(string)), "\"")))
		if err != nil {
			if rr > 3 {
				return err
			}
			time.Sleep(2 * time.Second)
			continue
		}
		break
	}

	hostname := d.Get("name").(string)
	pScript := fmt.Sprintf(provisionBash, hostname, strings.Split(hostname, ".")[0])

	log.Print("[DEBUG] sending provisionPayload")
	err = runCommand(comm, fmt.Sprintf(provisionPayload, strings.Trim(strconv.Quote(pScript), "\"")))
	if err != nil {
		return err
	}

	log.Print("[DEBUG] running provisionPayload")
	err = runCommand(comm, "sudo bash /tmp/tf_preprovision.sh \"$SSH_CLIENT\" >> /tmp/tf_preprovision.log 2>&1")
	if err != nil {
		return err
	}

	log.Print("[DEBUG] disconnecting SSH")
	comm.Disconnect()

	return err
}

// runCommand is used to run already prepared commands
func runCommand(
	comm communicator.Communicator,
	command string) error {

	_, outW := io.Pipe()
	_, errW := io.Pipe()
	//outDoneCh := make(chan struct{})
	//errDoneCh := make(chan struct{})
	// go copyOutput(o, outR, outDoneCh)
	// go copyOutput(o, errR, errDoneCh)

	cmd := &remote.Cmd{
		Command: command,
		Stdout:  outW,
		Stderr:  errW,
	}

	err := comm.Start(cmd)
	if err != nil {
		return fmt.Errorf("Error executing command %q: %v", cmd.Command, err)
	}

	if err := cmd.Wait(); err != nil {
		return err
	}

	// Wait for output to clean up
	outW.Close()
	errW.Close()
	//<-outDoneCh
	//<-errDoneCh

	return err
}

//
// func copyOutput(o terraform.UIOutput, r io.Reader, doneCh chan<- struct{}) {
// 	defer close(doneCh)
// 	lr := linereader.New(r)
// 	for line := range lr.Ch {
// 		o.Output(line)
// 	}
// }
