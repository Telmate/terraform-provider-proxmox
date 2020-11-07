package proxmox

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"os"
	"strings"
	"testing"
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
	ConfigureLogger(true, "acctest.log", map[string]string{"_default": "debug", "_capturelog": ""})
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

// testAccExampleResource generates a virtual machine and uses the disk
// slot setting to assign a non-standard disk position (scsi5 vs scsi0)
func testAccExampleQemuWithDiskSlot(name string, diskSlot int, targetNode string) string {
	return fmt.Sprintf(`
resource "proxmox_vm_qemu" "%s" {
  name = "%s"
  target_node = "%s"
  iso = "local:iso/SpinRite.iso"
  disk {
    size = "1G"
    type = "scsi"
    storage = "local"
    slot = %v
  }
}
`, name, name, targetNode, diskSlot)
}

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
				),
			},
		},
	})
}
