package proxmox

import (
	"fmt"
	"github.com/hashicorp/terraform/communicator"
	"github.com/hashicorp/terraform/communicator/remote"
	"github.com/hashicorp/terraform/helper/schema"
	// "github.com/hashicorp/terraform/terraform"
	// "github.com/mitchellh/go-linereader"
	"io"
)

// preprovision VM (setup eth0 and hostname)

func preProvisionUbuntu(d *schema.ResourceData) error {

	// Get a new communicator
	comm, err := communicator.New(d.State())
	if err != nil {
		return err
	}

	err = runCommand(comm, "echo cool > /tmp/test")

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

	cmd.Wait()
	if cmd.ExitStatus != 0 {
		err = fmt.Errorf(
			"Command %q exited with non-zero exit status: %d", cmd.Command, cmd.ExitStatus)
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
