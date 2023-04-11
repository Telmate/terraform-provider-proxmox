package proxmox

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// TODO is there a better place for this config?
var testAccProxmoxTargetNode string = "testproxmox"

// testAccPreCheck validates the necessary test API keys exist
// in the testing environment
func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("PM_API_URL"); v == "" {
		t.Fatal("PM_API_URL must be set for acceptance tests")
	}
	if v := os.Getenv("PM_USER"); v == "" {
		t.Fatal("PM_USER must be set for acceptance tests")
	}
	if v := os.Getenv("PM_PASS"); v == "" {
		t.Fatal("PM_PASS must be set for acceptance tests")
	}
}

func testAccProxmoxProviderFactory() map[string]*schema.Provider {
	providers := map[string]*schema.Provider{
		"proxmox": Provider(),
	}
	// TODO move this log configuration elsewhere, it doesn't make sense here but
	// it's a short term solution to test the testing out
	ConfigureLogger(true, "../../acctest.log", map[string]string{"_default": "debug", "_capturelog": ""})
	return providers
}

// testAccCheckVmCreate tests a simple creation/destruction cycle
//func testAccCheckExampleResourceDestroy(s *terraform.State) error {
//	// retrieve the connection established in Provider configuration
//	conn := testAccProvider.Meta().(*ExampleClient)
//
//	// loop through the resources in state, verifying each widget
//	// is destroyed
//	for _, rs := range s.RootModule().Resources {
//		if rs.Type != "example_widget" {
//			continue
//		}
//
//		// Retrieve our widget by referencing it's state ID for API lookup
//		request := &example.DescribeWidgets{
//			IDs: []string{rs.Primary.ID},
//		}
//
//		response, err := conn.DescribeWidgets(request)
//		if err == nil {
//			if len(response.Widgets) > 0 && *response.Widgets[0].ID == rs.Primary.ID {
//				return fmt.Errorf("Widget (%s) still exists.", rs.Primary.ID)
//			}
//
//			return nil
//		}
//
//		// If the error is equivalent to 404 not found, the widget is destroyed.
//		// Otherwise return the error
//		if !strings.Contains(err.Error(), "Widget not found") {
//			return err
//		}
//	}
//
//	return nil
//}

// testAccExampleQemuBasic generates the most simplistic VM we're able to make
// this confirms we can spin up a vm using just default values
func testAccExampleQemuBasic(name string, targetNode string) string {
	return fmt.Sprintf(`
resource "proxmox_vm_qemu" "%s" {
  name = "%s"
  target_node = "%s"
  iso = "local:iso/SpinRite.iso"
}
`, name, name, targetNode)
}

// testAccExampleQemuPxe generates the most simplistic PXE boot VM
// we're able to make this confirms we can spin up a PXE boot VM
// using just default values, a valid Network must be specified
// for the VM to be able to Network boot
func testAccExampleQemuPxe(name string, targetNode string) string {
	return fmt.Sprintf(`
resource "proxmox_vm_qemu" "%s" {
  name = "%s"
  target_node = "%s"
	pxe = true
	boot = "order=scsi0;net0"
	network {
    bridge    = "vmbr0"
    firewall  = false
    link_down = false
    model     = "e1000"
  }
}
`, name, name, targetNode)
}

// testAccExampleResource generates a virtual machine and uses the disk
// slot setting to assign a non-standard disk position (scsi5 vs scsi0)
// func testAccExampleQemuWithDiskSlot(name string, diskSlot int, targetNode string) string {
// 	return fmt.Sprintf(`
// resource "proxmox_vm_qemu" "%s" {
//   name = "%s"
//   target_node = "%s"
//   iso = "local:iso/SpinRite.iso"
//   disk {
//     size = "1G"
//     type = "scsi"
//     storage = "local"
//     slot = %v
//   }
// }
// `, name, name, targetNode, diskSlot)
// }

// testAccExampleResource generates a configured VM with a 1G disk
// the goal with this resource is to make a "basic" but "standard" virtual machine
// using a configuration that would apply to a usable vm (but NOT a cloud config'd one)
func testAccExampleQemuStandard(name string, targetNode string) string {
	return fmt.Sprintf(`
resource "proxmox_vm_qemu" "%s" {
  name = "%s"
  target_node = "%s"
  iso = "local:iso/SpinRite.iso"
  disk {
    size = "1G"
    type = "scsi"
    storage = "local"
  }
}
`, name, name, targetNode)
}

// testAccExampleQemuOvmf generates a simple VM which uses EFI bios
func testAccExampleQemuOvmf(name string, targetNode string) string {
	return fmt.Sprintf(`
resource "proxmox_vm_qemu" "%s" {
  name = "%s"
  target_node = "%s"
  iso = "local:iso/SpinRite.iso"
  bios = "ovmf"
}
`, name, name, targetNode)
}

// testAccExampleResourceClone generate two simply configured VMs where the second is a
// clone of the first.
func testAccExampleQemuClone(name string, name_clone string, targetNode string) string {
	source_resource := testAccExampleQemuStandard(name, targetNode)
	clone_resource := fmt.Sprintf(`
resource "proxmox_vm_qemu" "%s" {
  name = "%s"
  target_node = "%s"
  clone = "%s"
  disk {
    size = "1G"
    type = "scsi"
    storage = "local"
  }
  depends_on = [proxmox_vm_qemu.%s]
}
`, name_clone, name_clone, targetNode, name, name)
	return strings.Join([]string{source_resource, clone_resource}, "\n")
}

// testAccExampleResourceClone generate two simply configured VMs where the second is a
// clone of the first.
func testAccExampleQemuCloneWithTwoDisks(name string, name_clone string, targetNode string) string {
	source_resource := testAccExampleQemuStandard(name, targetNode)
	clone_resource := fmt.Sprintf(`
resource "proxmox_vm_qemu" "%s" {
  name = "%s"
  target_node = "%s"
  clone = "%s"
  disk {
    size = "2G"
    type = "scsi"
    storage = "local"
  }
  disk {
    size = "3G"
    type = "scsi"
    storage = "local"
  }
  depends_on = [proxmox_vm_qemu.%s]
}
`, name_clone, name_clone, targetNode, name, name)
	return strings.Join([]string{source_resource, clone_resource}, "\n")
}

// TestAccProxmoxVmQemu_BasicCreate tests a simple creation and destruction of the smallest, but
// but still viable, configuration for a VM we can create.
func TestAccProxmoxVmQemu_BasicCreate(t *testing.T) {
	resourceName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resourcePath := fmt.Sprintf("proxmox_vm_qemu.%s", resourceName)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProxmoxProviderFactory(),
		//CheckDestroy: testAccCheckExampleResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccExampleQemuBasic(resourceName, testAccProxmoxTargetNode),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourcePath, "name", resourceName),
				),
			},
		},
	})
}

// TestAccProxmoxVmQemu_BasicCreate tests a simple creation and destruction of the smallest, but
// but still viable, configuration for a VM we can create.
func TestAccProxmoxVmQemu_StandardCreate(t *testing.T) {
	resourceName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resourcePath := fmt.Sprintf("proxmox_vm_qemu.%s", resourceName)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProxmoxProviderFactory(),
		//CheckDestroy: testAccCheckExampleResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccExampleQemuStandard(resourceName, testAccProxmoxTargetNode),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourcePath, "name", resourceName),
				),
			},
		},
	})
}

// TODO:  this test FAILS - it looks like the api library isn't actually sending the slot request to proxmox? needs further investigation.
// TestAccProxmoxVmQemu_DiskSlot tests we can correctly create a vm disk assigned to a particular disk slot
//func TestAccProxmoxVmQemu_DiskSlot(t *testing.T) {
//	resourceName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
//	resourcePath := fmt.Sprintf("proxmox_vm_qemu.%s", resourceName)
//	diskSlot := 5
//
//	resource.Test(t, resource.TestCase{
//		PreCheck:  func() { testAccPreCheck(t) },
//		Providers: testAccProxmoxProviderFactory(),
//		//CheckDestroy: testAccCheckExampleResourceDestroy,
//		Steps: []resource.TestStep{
//			{
//				Config: testAccExampleQemuWithDiskSlot(resourceName, diskSlot, testAccProxmoxTargetNode),
//				Check: resource.ComposeTestCheckFunc(
//					resource.TestCheckResourceAttr(resourcePath, "name", resourceName),
//					resource.TestCheckResourceAttr(resourcePath, "disk.0.slot", fmt.Sprintf("%v", diskSlot)),
//				),
//			},
//		},
//	})
//}

// TestAccProxmoxVmQemu_BasicCreateClone create a minimally configured VM, then creates a similar
// minimally configured clone from the original VM.
func TestAccProxmoxVmQemu_BasicCreateClone(t *testing.T) {
	resourceName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resourcePath := fmt.Sprintf("proxmox_vm_qemu.%s", resourceName)
	cloneName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	clonePath := fmt.Sprintf("proxmox_vm_qemu.%s", cloneName)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProxmoxProviderFactory(),
		//CheckDestroy: testAccCheckExampleResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccExampleQemuClone(resourceName, cloneName, testAccProxmoxTargetNode),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourcePath, "name", resourceName),
					resource.TestCheckResourceAttr(clonePath, "name", cloneName),
					// check for unused_disk.0.file existance as that means an extra disk popped up
					// which would be a regression of https://github.com/Telmate/terraform-provider-proxmox/issues/239
					resource.TestCheckNoResourceAttr(clonePath, "unused_disk.0.file"),
				),
			},
		},
	})
}

// TestAccProxmoxVmQemu_CreateCloneWithTwoDisks create a minimally configured VM, then creates a cloned vm
// with two disks, each increased in size compared to the original vm
func TestAccProxmoxVmQemu_CreateCloneWithTwoDisks(t *testing.T) {
	resourceName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resourcePath := fmt.Sprintf("proxmox_vm_qemu.%s", resourceName)
	cloneName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	clonePath := fmt.Sprintf("proxmox_vm_qemu.%s", cloneName)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProxmoxProviderFactory(),
		//CheckDestroy: testAccCheckExampleResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccExampleQemuCloneWithTwoDisks(resourceName, cloneName, testAccProxmoxTargetNode),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourcePath, "name", resourceName),
					resource.TestCheckResourceAttr(clonePath, "name", cloneName),
					// check for unused_disk.0.file existance as that means an extra disk popped up
					// which would be a regression of https://github.com/Telmate/terraform-provider-proxmox/issues/239
					resource.TestCheckNoResourceAttr(clonePath, "unused_disk.0.file"),
					resource.TestCheckResourceAttr(clonePath, "disk.0.size", "2G"),
					resource.TestCheckResourceAttr(clonePath, "disk.1.size", "3G"),
				),
			},
		},
	})
}

// TestAccProxmoxVmQemu_PxeCreate tests a simple creation and destruction of the smallest, but
// but still viable, configuration for a PXE Network boot VM we can create.
func TestAccProxmoxVmQemu_PxeCreate(t *testing.T) {
	resourceName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resourcePath := fmt.Sprintf("proxmox_vm_qemu.%s", resourceName)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProxmoxProviderFactory(),
		//CheckDestroy: testAccCheckExampleResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccExampleQemuPxe(resourceName, testAccProxmoxTargetNode),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourcePath, "name", resourceName),
				),
			},
		},
	})
}

// TestAccProxmoxVmQemu_StandardUpdateNoReboot tests a simple update of a vm_qemu resource,
// and the modified parameters can be applied without reboot.
func TestAccProxmoxVmQemu_UpdateNoReboot(t *testing.T) {
	resourceName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resourcePath := fmt.Sprintf("proxmox_vm_qemu.%s", resourceName)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProxmoxProviderFactory(),

		Steps: []resource.TestStep{
			{
				Config: testAccExampleQemuBasic(resourceName, testAccProxmoxTargetNode),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourcePath, "name", resourceName),
				),
			},
			{
				// since we're just renaming there should be no reboot
				Config: strings.Replace(testAccExampleQemuBasic(resourceName, testAccProxmoxTargetNode),
					"name = \""+resourceName+"\"", "name = \""+resourceName+"-renamed\"", 1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourcePath, "name", resourceName+"-renamed"),
				),
			},
		},
	})
}

// TestAccProxmoxVmQemu_UpdateRebootRequired tests a simple update of a vm_qemu resource,
// and the modified parameters can be only applied with reboot.
func TestAccProxmoxVmQemu_UpdateRebootRequired(t *testing.T) {
	resourceName := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	resourcePath := fmt.Sprintf("proxmox_vm_qemu.%s", resourceName)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProxmoxProviderFactory(),

		Steps: []resource.TestStep{
			{
				Config: testAccExampleQemuBasic(resourceName, testAccProxmoxTargetNode),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourcePath, "bios", "seabios"),
				),
			},
			{
				// changing the BIOS platform always requires a reboot
				Config: testAccExampleQemuOvmf(resourceName, testAccProxmoxTargetNode),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourcePath, "bios", "ovmf"),
				),
			},
		},
	})
}
