package proxmox

import (
	"context"
	"errors"
	"path"
	"strconv"

	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func guestDelete(ctx context.Context, d *schema.ResourceData, meta any, kind string) diag.Diagnostics {
	pconf := meta.(*providerConfiguration)
	lock := pmParallelBegin(pconf)
	defer lock.unlock()

	client := pconf.Client
	rawID, _ := strconv.Atoi(path.Base(d.Id()))
	guestID := pveSDK.GuestID(rawID)

	if err := pveSDK.NewVmRef(guestID).Delete(ctx, client); err != nil {
		if errors.Is(err, pveSDK.Error.GuestDoesNotExist()) {
			return diag.Diagnostics{{
				Summary:  "guest of type " + kind + " with ID " + guestID.String() + " already removed",
				Severity: diag.Warning}}
		}
		return diag.FromErr(err)
	}
	return nil
}

func guestGetSourceVmr(
	ctx context.Context,
	client *pveSDK.Client,
	name pveSDK.GuestName,
	id pveSDK.GuestID,
	preferredNode pveSDK.NodeName,
	guest pveSDK.GuestType,
	fieldName, fieldID string) (*pveSDK.VmRef, error) {
	if name != "" {
		rawGuests, err := pveSDK.ListGuests(ctx, client)
		if err != nil {
			return nil, err
		}
		return guestGetSourceVmrByNode(rawGuests, name, preferredNode, guest)
	} else if id != 0 {
		rawGuests, err := pveSDK.ListGuests(ctx, client)
		if err != nil {
			return nil, err
		}
		rawGuest, ok := rawGuests.SelectID(id)
		if !ok {
			return nil, errors.New("guest with ID '" + id.String() + "' does not exist")
		}
		if rawGuest.GetType() != guest {
			return nil, errors.New("guest with ID '" + id.String() + "' is not of type '" + string(guest) + "'")
		}
		guestRef := pveSDK.NewVmRef(rawGuest.GetID())
		guestRef.SetNode(string(rawGuest.GetNode()))
		guestRef.SetVmType(guest)
		return guestRef, nil
	}
	return nil, errors.New("either '" + fieldName + "' or '" + fieldID + "' must be specified")
}

func guestGetSourceVmrByNode(raw pveSDK.RawGuestResources, name pveSDK.GuestName, preferredNode pveSDK.NodeName, guest pveSDK.GuestType) (*pveSDK.VmRef, error) {
	var guestRef *pveSDK.VmRef
	for i := range raw {
		if raw[i].GetName() == name {
			if raw[i].GetType() == guest {
				if node := raw[i].GetNode(); node == preferredNode { // Prefer source VM on the same node
					guestRef = pveSDK.NewVmRef(raw[i].GetID())
					guestRef.SetNode(string(node))
					guestRef.SetVmType(guest)
					return guestRef, nil
				}
				if guestRef == nil { // remember the first we find
					guestRef = pveSDK.NewVmRef(raw[i].GetID())
					guestRef.SetNode(string(raw[i].GetNode()))
					guestRef.SetVmType(guest)
				}
			}
		}
	}
	if guestRef == nil {
		return nil, errors.New("no guest with name '" + name.String() + "' found")
	}
	return guestRef, nil
}
