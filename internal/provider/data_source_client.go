package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/umich-vci/gospoc"
)

func dataSourceClient() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceClientRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"server_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"platform": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"domain": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"locked": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"version": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"vm_owner": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"guid": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"link": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"type": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"vm_type": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"contact": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"deduplication": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"email": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"authentication": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"session_initiation": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"decommissioned": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"ssl_required": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"option_set": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"split_large_objects": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"at_risk": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"schedules": {
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
			"filespaces": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"fs_logical_mb": {
							Type:     schema.TypeFloat,
							Computed: true,
						},
						"fs_copy_pool": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"link": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"fs_num_files": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"fs_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
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
	client := meta.(*apiClient).Client

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

	if backupClient.Name == "" {
		return fmt.Errorf("no backup client named %s found on server %s", name, serverName)
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

	decommissioned, err := parseYesNoBool(backupClientDetails.Decommissioned)
	if err != nil {
		return err
	}

	d.Set("contact", backupClientDetails.Contact)
	d.Set("deduplication", backupClientDetails.Deduplication)
	d.Set("email", backupClientDetails.Email)
	d.Set("authentication", backupClientDetails.Authentication)
	d.Set("session_initiation", backupClientDetails.SessionInitiation)
	d.Set("decommissioned", decommissioned)
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

	rawFilespaces, _, err := client.Clients.FileSpaces(context.Background(), serverName, name)
	if err != nil {
		return err
	}
	backupFilespaces := make([]map[string]interface{}, 0, len(rawFilespaces))
	for i := range rawFilespaces {
		backupFilespaces = append(backupFilespaces, map[string]interface{}{
			"fs_logical_mb": rawFilespaces[i].FSLogicalMB,
			"fs_copy_pool":  rawFilespaces[i].FSCopyPool,
			"link":          rawFilespaces[i].Link,
			"fs_num_files":  rawFilespaces[i].FSNumFiles,
			"id":            rawFilespaces[i].ID,
			"fs_type":       rawFilespaces[i].FSType,
			"name":          rawFilespaces[i].Name,
		})
	}
	d.Set("filespaces", backupFilespaces)

	return nil
}
