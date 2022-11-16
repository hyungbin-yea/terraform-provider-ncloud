package ncloud

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"os"
	"testing"
)

func TestAccDataSourceNcloudSESCluster(t *testing.T) {
	dataName := "data.ncloud_ses_cluster.cluster"
	resourceName := "ncloud_ses_cluster.cluster"
	testClusterName := getTestClusterName()
	searchEngineVersionCode := "133"
	region := os.Getenv("NCLOUD_REGION")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceSESClusterConfig(testClusterName, TF_TEST_SES_LOGIN_KEY, searchEngineVersionCode, region),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceID(dataName),
					resource.TestCheckResourceAttrPair(dataName, "service_group_instance_no", resourceName, "service_group_instance_no"),
					resource.TestCheckResourceAttrPair(dataName, "cluster_name", resourceName, "cluster_name"),
					resource.TestCheckResourceAttrPair(dataName, "vpc_no", resourceName, "vpc_no"),
					resource.TestCheckResourceAttrPair(dataName, "software_product_code", resourceName, "software_product_code"),
				),
			},
		},
	})
}

func testAccDataSourceSESClusterConfig(testClusterName string, loginKey string, version string, region string) string {
	return fmt.Sprintf(`
resource "ncloud_vpc" "vpc" {
	name               = "%[1]s"
	ipv4_cidr_block    = "192.168.0.0/16"
}

resource "ncloud_subnet" "node_subnet" {
	vpc_no             = ncloud_vpc.vpc.vpc_no
	name               = "%[1]s"
	subnet             = "192.168.1.0/24"
	zone               = "%[4]s-1"
	network_acl_no     = ncloud_vpc.vpc.default_network_acl_no
	subnet_type        = "PRIVATE"
	usage_type         = "GEN"
}
data "ncloud_ses_versions" "version" {
}

data "ncloud_ses_software_product" "os_version" {
}

data "ncloud_ses_node_product" "product_codes" {
  software_product_code = data.ncloud_ses_software_product.os_version.codes.0.value
  subnet_no = ncloud_subnet.node_subnet.id
}

resource "ncloud_login_key" "loginkey" {
  key_name = "%[2]s"
}

resource "ncloud_ses_cluster" "cluster" {
  cluster_name                  = "%[1]s"
  software_product_code         = data.ncloud_ses_software_product.os_version.codes.0.value
  vpc_no                        = ncloud_vpc.vpc.id
  search_engine {
	  version_code    			= "%[3]s"
	  user_name       			= "admin"
	  user_password   			= "qwe123!@#"
  }
  manager_node {  
	  is_dual_manager           = false
	  product_code     			= data.ncloud_ses_node_product.product_codes.codes.0.value
	  subnet_no        			= ncloud_subnet.node_subnet.id
  }
  data_node {
	  product_code       		= data.ncloud_ses_node_product.product_codes.codes.0.value
	  subnet_no           		= ncloud_subnet.node_subnet.id
	  count            		    = 3
	  storage_size        		= 100
  }
  master_node {
	  is_master_only_node_activated = false
  }
  login_key_name                = ncloud_login_key.loginkey.key_name
}

data "ncloud_ses_cluster" "cluster" {
	service_group_instance_no = ncloud_ses_cluster.cluster.uuid
}


`, testClusterName, loginKey, version, region)
}
