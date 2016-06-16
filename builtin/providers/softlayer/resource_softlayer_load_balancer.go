package softlayer

import (
	"fmt"
	"log"
	"strconv"

	softlayer "github.com/TheWeatherCompany/softlayer-go/softlayer"
	"github.com/hashicorp/terraform/helper/schema"
)

const (
	LB_LARGE_150000_CONNECTIONS = 150000
	LB_SMALL_15000_CONNECTIONS  = 15000
)

func resourceSoftLayerLoadBalancer() *schema.Resource {
	return &schema.Resource{
		Create: resourceSoftLayerLoadBalancerCreate,
		Read:   resourceSoftLayerLoadBalancerRead,
		Delete: resourceSoftLayerLoadBalancerDelete,
		Exists: resourceSoftLayerLoadBalancerExists,

		Schema: map[string]*schema.Schema{
			"connections": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"location": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"ha_enabled": &schema.Schema{
				Type:     schema.TypeBool,
				Required: true,
				ForceNew: true,
			},
			"ip_address": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"subnet_id": &schema.Schema{
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func resourceSoftLayerLoadBalancerCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client).loadBalancerService
	if client == nil {
		return fmt.Errorf("The client is nil.")
	}

	opts := softlayer.SoftLayer_Load_Balancer_CreateOptions{
		Connections: d.Get("connections").(int),
		Location:    d.Get("location").(string),
		HaEnabled:   d.Get("ha_enabled").(bool),
	}

	log.Printf("[INFO] Creating load balancer")

	loadBalancer, err := client.CreateLoadBalancer(&opts)

	if err != nil {
		return fmt.Errorf("Error creating load balancer: %s", err)
	}

	d.SetId(fmt.Sprintf("%d", loadBalancer.Id))
	d.Set("connections", getConnectionLimit(loadBalancer.ConnectionLimit))
	d.Set("location", loadBalancer.SoftlayerHardware[0].Datacenter.Name)
	d.Set("ip_address", loadBalancer.IpAddress.IpAddress)
	d.Set("subnet_id", loadBalancer.IpAddress.SubnetId)
	d.Set("ha_enabled", loadBalancer.HaEnabled)

	log.Printf("[INFO] Load Balancer ID: %s", d.Id())

	return resourceSoftLayerLoadBalancerRead(d, meta)
}

func resourceSoftLayerLoadBalancerRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client).loadBalancerService
	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return fmt.Errorf("Not a valid ID, must be an integer: %s", err)
	}
	getObjectResult, err := client.GetObject(id)
	if err != nil {
		return fmt.Errorf("Error retrieving load balancer: %s", err)
	}

	d.SetId(strconv.Itoa(getObjectResult.Id))
	d.Set("connections", getConnectionLimit(getObjectResult.ConnectionLimit))
	d.Set("location", getObjectResult.SoftlayerHardware[0].Datacenter.Name)
	d.Set("ip_address", getObjectResult.IpAddress.IpAddress)
	d.Set("subnet_id", getObjectResult.IpAddress.SubnetId)
	d.Set("ha_enabled", getObjectResult.HaEnabled)

	return nil
}

func resourceSoftLayerLoadBalancerDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client).loadBalancerService
	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return fmt.Errorf("Not a valid ID, must be an integer: %s", err)
	}

	_, err = client.DeleteObject(id)

	if err != nil {
		return fmt.Errorf("Error deleting network application delivery controller load balancer: %s", err)
	}

	return nil
}

func resourceSoftLayerLoadBalancerExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	return true, nil
}

/* When requesting 15000 SL creates between 15000 and 150000. When requesting 150000 SL creates >= 150000 */
func getConnectionLimit(connectionLimit int) int {
	if connectionLimit >= LB_LARGE_150000_CONNECTIONS {
		return LB_LARGE_150000_CONNECTIONS
	} else if connectionLimit >= LB_SMALL_15000_CONNECTIONS &&
		connectionLimit < LB_LARGE_150000_CONNECTIONS {
		return LB_SMALL_15000_CONNECTIONS
	} else {
		return 0
	}
}
