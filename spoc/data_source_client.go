package spoc

import (
	"context"

	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceClient() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceClientRead,
		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"server_name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"platform": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"domain": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"locked": &schema.Schema{
				Type:     schema.TypeBool,
				Computed: true,
			},
			"version": &schema.Schema{
				Type:     schema.TypeInt,
				Computed: true,
			},
			"vm_owner": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"guid": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"link": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"type": &schema.Schema{
				Type:     schema.TypeInt,
				Computed: true,
			},
			"vm_type": &schema.Schema{
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func dataSourceClientRead(d *schema.ResourceData, meta interface{}) error {
	client, err := meta.(*Config).Client()
	if err != nil {
		return err
	}

	name := d.Get("name").(string)
	serverName := d.Get("server_name").(string)

	backupClient, _, err := client.Clients.Details(context.Background(), serverName, name)
	if err != nil {
		return err
	}

	locked := backupClient.Locked != 0

	d.SetId(backupClient.Name)
	d.Set("platform", backupClient.Platform)
	d.Set("domain", backupClient.Domain)
	d.Set("locked", locked)
	d.Set("version", backupClient.Version)
	d.Set("vm_owner", backupClient.VMOwner)
	d.Set("guid", backupClient.GUID)
	d.Set("link", backupClient.Link)
	d.Set("type", backupClient.Type)
	d.Set("vm_type", backupClient.VMType)

	return nil
}
