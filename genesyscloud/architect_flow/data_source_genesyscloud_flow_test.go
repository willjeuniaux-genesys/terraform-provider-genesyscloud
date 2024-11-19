package architect_flow

import (
	"fmt"
	"path/filepath"
	"strconv"
	"terraform-provider-genesyscloud/genesyscloud/provider"
	"terraform-provider-genesyscloud/genesyscloud/util"
	"testing"

	"github.com/google/uuid"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceFlow(t *testing.T) {
	var (
		flowDataSource    = "flow-data"
		flowName          = "test_data_flow" + uuid.NewString()
		inboundcallConfig = fmt.Sprintf("inboundCall:\n  name: %s\n  defaultLanguage: en-us\n  startUpRef: ./menus/menu[mainMenu]\n  initialGreeting:\n    tts: Archy says hi!!!\n  menus:\n    - menu:\n        name: Main Menu\n        audio:\n          tts: You are at the Main Menu, press 9 to disconnect.\n        refId: mainMenu\n        choices:\n          - menuDisconnect:\n              name: Disconnect\n              dtmf: digit_9", flowName)

		flowResource = "test_flow"
		filePath     = filepath.Join("..", "..", "examples", "resources", "genesyscloud_flow", "inboundcall_flow_example.yaml")
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { util.TestAccPreCheck(t) },
		ProviderFactories: provider.GetProviderFactories(providerResources, providerDataSources),
		Steps: []resource.TestStep{
			{
				Config: GenerateFlowResource(
					flowResource,
					filePath,
					inboundcallConfig,
					false,
				) + generateFlowDataSource(
					flowDataSource,
					resourceName+"."+flowResource,
					flowName,
					strconv.Quote("inboundcall"),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName+"."+flowResource, "id",
						fmt.Sprintf("data.%s.%s", resourceName, flowDataSource), "id"),
				),
			},
			{
				// Import/Read
				ResourceName:            resourceName + "." + flowResource,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"filepath", "force_unlock", "file_content_hash"},
			},
		},
		CheckDestroy: testVerifyFlowDestroyed,
	})
}

func generateFlowDataSource(
	resourceID,
	dependsOn,
	name,
	varType string) string {
	return fmt.Sprintf(`data "%s" "%s" {
		name = "%s"
		type = %s
		depends_on = [%s]
	}
	`, resourceName, resourceID, name, varType, dependsOn)
}