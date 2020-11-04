package proxmox

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	//"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"os"
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
	providers := map[string]*schema.Provider {
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

func testAccExampleResource(name string, target_node string) string {
	return fmt.Sprintf(`
resource "proxmox_vm_qemu" "%s" {
  name = "%s"
  target_node = "%s"
  iso = "local:iso/SpinRite.iso"
}
`, name, name, target_node)
}

func TestAccProxmoxVmQemu_BasicCreate(t *testing.T) {
	resourceName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resourcePath := fmt.Sprintf("proxmox_vm_qemu.%s", resourceName)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProxmoxProviderFactory(),
		//CheckDestroy: testAccCheckExampleResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccExampleResource(resourceName, testAccProxmoxTargetNode),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourcePath, "name", resourceName),
				),
			},
		},
	})
}
