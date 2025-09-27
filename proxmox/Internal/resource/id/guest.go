package id

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
)

const (
	GuestLxc  = "lxc"
	GuestQemu = "qemu"
)

type Guest struct {
	ID   pveSDK.GuestID
	Node pveSDK.NodeName
	Type string
}

func (g *Guest) Parse(resourceID string) error {
	idParts := strings.Split(resourceID, "/")
	if len(idParts) != 3 {
		return errors.New("failed to get resource format: '" + resourceID + "'. Must be <node>/<type>/<vmid>")
	}
	if idParts[0] == "" {
		return errors.New("failed to get node name: '" + idParts[0] + "'")
	}
	g.Node = pveSDK.NodeName(idParts[0])
	if idParts[1] != GuestLxc && idParts[1] != GuestQemu {
		return errors.New("failed to get guest type: '" + idParts[1] + "'. Must be 'lxc' or 'qemu'")
	}
	g.Type = idParts[1]
	tmpID, err := strconv.Atoi(idParts[2])
	if err != nil {
		return fmt.Errorf("failed to get vmid: '%s'. Must be an integer", idParts[2])
	}
	g.ID = pveSDK.GuestID(tmpID)
	return nil
}

func (g Guest) String() string {
	return g.Node.String() + "/" + g.Type + "/" + g.ID.String()
}
