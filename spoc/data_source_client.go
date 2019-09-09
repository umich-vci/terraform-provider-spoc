package spoc

import (
	"context"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/umich-vci/gospoc"
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
			"contact": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"deduplication": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"email": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"authentication": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"session_initiation": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"decommissioned": &schema.Schema{
				Type:     schema.TypeBool,
				Computed: true,
			},
			"ssl_required": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"option_set": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"split_large_objects": &schema.Schema{
				Type:     schema.TypeBool,
				Computed: true,
			},
			"at_risk": &schema.Schema{
				Type:     schema.TypeBool,
				Computed: true,
			},
			"schedules": &schema.Schema{
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"start_time": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"server_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"schedule_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"run_time": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"domain_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
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

	backupClients, _, err := client.Clients.List(context.Background())
	if err != nil {
		return err
	}
	var backupClient gospoc.BackupClient
	for i := range backupClients {
		if backupClients[i].Name == name && backupClients[i].Server == serverName {
			backupClient = backupClients[i]
		}
	}

	d.SetId(backupClient.Name)
	d.Set("platform", backupClient.Platform)
	d.Set("domain", backupClient.Domain)
	d.Set("locked", backupClient.Locked != 0)
	d.Set("version", backupClient.Version)
	d.Set("vm_owner", backupClient.VMOwner)
	d.Set("guid", backupClient.GUID)
	d.Set("link", backupClient.Link)
	d.Set("type", backupClient.Type)
	d.Set("vm_type", backupClient.VMType)

	backupClientDetails, _, err := client.Clients.Details(context.Background(), serverName, name)
	if err != nil {
		return err
	}

	splitLargeObjects, err := parseYesNoBool(backupClientDetails.SplitLargeObjects)
	if err != nil {
		return err
	}

	d.Set("contact", backupClientDetails.Contact)
	d.Set("deduplication", backupClientDetails.Deduplication)
	d.Set("email", backupClientDetails.Email)
	d.Set("authentication", backupClientDetails.Authentication)
	d.Set("session_initiation", backupClientDetails.SessionInitiation)
	d.Set("decommissioned", backupClientDetails.Decommissioned)
	d.Set("ssl_required", backupClientDetails.SSLRequired)
	d.Set("option_set", backupClientDetails.OptionSet)
	d.Set("split_large_objects", splitLargeObjects)

	backupClientAtRisk, _, err := client.Clients.AtRisk(context.Background(), serverName, name)
	if err != nil {
		return err
	}

	atRisk, err := parseYesNoBool(backupClientAtRisk.AtRisk)
	if err != nil {
		return err
	}
	d.Set("at_risk", atRisk)

	rawBackupClientSchedules, _, err := client.Clients.Schedules(context.Background(), serverName, backupClient.Domain, name)
	if err != nil {
		return err
	}

	backupClientSchedules := make([]map[string]interface{}, 0, len(rawBackupClientSchedules))
	for i := range rawBackupClientSchedules {
		backupClientSchedules = append(backupClientSchedules, map[string]interface{}{
			"start_time":    rawBackupClientSchedules[i].StartTime,
			"server_name":   rawBackupClientSchedules[i].ServerName,
			"schedule_name": rawBackupClientSchedules[i].ScheduleName,
			"run_time":      rawBackupClientSchedules[i].RunTime,
			"domain_name":   rawBackupClientSchedules[i].DomainName,
		})
	}

	d.Set("schedules", backupClientSchedules)

	return nil
}
