package spoc

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	"github.com/umich-vci/gospoc"
)

func resourceClient() *schema.Resource {
	return &schema.Resource{
		Create: resourceClientCreate,
		Read:   resourceClientRead,
		Update: resourceClientUpdate,
		Delete: resourceClientDelete,
		Schema: map[string]*schema.Schema{
			"server_name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"authentication": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "Local",
				ValidateFunc: validation.StringInSlice([]string{"Local", "LDAP"}, false),
				ForceNew:     true,
			},
			"password": &schema.Schema{
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
			},
			"domain": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"contact": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"email": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"schedule": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"option_set": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				//Default:  "",
			},
			"deduplication": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  "ClientOrServer",
			},
			"ssl_required": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      "Default",
				ValidateFunc: validation.StringInSlice([]string{"Default", "YES", "NO"}, false),
			},
			"session_initiation": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      "ClientOrServer",
				ValidateFunc: validation.StringInSlice([]string{"ClientOrServer", "Serveronly"}, false),
			},
			"locked": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"link": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"decommissioned": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"split_large_objects": &schema.Schema{
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

func resourceClientRead(d *schema.ResourceData, meta interface{}) error {
	client, err := meta.(*Config).Client()
	if err != nil {
		return err
	}

	name := d.Get("name").(string)
	serverName := d.Get("server_name").(string)

	backupClientDetails, _, err := client.Clients.Details(context.Background(), serverName, name)
	if err != nil {
		return err
	}

	splitLargeObjects, err := parseYesNoBool(backupClientDetails.SplitLargeObjects)
	if err != nil {
		return err
	}

	locked, err := parseYesNoBool(backupClientDetails.Locked)
	if err != nil {
		return err
	}

	d.SetId(backupClientDetails.Name)
	d.Set("contact", backupClientDetails.Contact)
	d.Set("deduplication", backupClientDetails.Deduplication)
	d.Set("email", backupClientDetails.Email)
	d.Set("authentication", backupClientDetails.Authentication)
	d.Set("session_initiation", backupClientDetails.SessionInitiation)
	d.Set("locked", locked)
	d.Set("decommissioned", backupClientDetails.Decommissioned)
	d.Set("ssl_required", backupClientDetails.SSLRequired)
	d.Set("option_set", backupClientDetails.OptionSet)
	d.Set("split_large_objects", splitLargeObjects)

	return nil
}

func resourceClientCreate(d *schema.ResourceData, meta interface{}) error {
	client, err := meta.(*Config).Client()
	if err != nil {
		return err
	}

	serverName := d.Get("server_name").(string)
	name := d.Get("name").(string)

	_, _, err = client.Clients.Details(context.Background(), serverName, name)
	if err == nil {
		return fmt.Errorf("A node with name %s already exists on server %s", name, serverName)
	}

	register := new(gospoc.RegisterClientRequest)
	register.Name = name
	register.Authentication = d.Get("authentication").(string)
	register.Password = d.Get("password").(string)
	register.Domain = d.Get("domain").(string)
	register.Contact = d.Get("contact").(string)
	register.Email = d.Get("email").(string)

	if schedule, ok := d.GetOk("schedule"); ok {
		register.Schedule = schedule.(string)
	}

	if optionSet, ok := d.GetOk("option_set"); ok {
		register.OptionSet = optionSet.(string)
	}

	if deduplication, ok := d.GetOk("deduplication"); ok {
		register.Deduplication = deduplication.(string)
	}

	if sslRequired, ok := d.GetOk("ssl_required"); ok {
		register.SSLRequired = sslRequired.(string)
	}

	if sessionInitiation, ok := d.GetOk("session_initiation"); ok {
		register.SSLRequired = sessionInitiation.(string)
	}

	_, err = client.Clients.RegisterNode(context.Background(), serverName, register)
	if err != nil {
		return err
	}

	return resourceClientRead(d, meta)
}

func resourceClientUpdate(d *schema.ResourceData, meta interface{}) error {
	client, err := meta.(*Config).Client()
	if err != nil {
		return err
	}

	serverName := d.Get("server_name").(string)
	name := d.Get("name").(string)

	if d.HasChange("locked") {
		locked := d.Get("locked").(bool)
		if locked {
			_, err := client.Clients.Lock(context.Background(), serverName, name)
			if err != nil {
				return err
			}
		} else {
			_, err := client.Clients.Unlock(context.Background(), serverName, name)
			if err != nil {
				return err
			}
		}
	}

	if d.HasChange("password") {
		password := d.Get("password").(string)
		_, err := client.Clients.UpdatePassword(context.Background(), serverName, name, password)
		if err != nil {
			return err
		}
	}

	if d.HasChange("schedule") || d.HasChange("domain") {
		schedule := d.Get("schedule").(string)
		domain := d.Get("domain").(string)
		_, err := client.Clients.AssignSchedule(context.Background(), serverName, name, domain, schedule)
		if err != nil {
			return err
		}
	}

	return resourceClientRead(d, meta)
}

func resourceClientDelete(d *schema.ResourceData, meta interface{}) error {
	client, err := meta.(*Config).Client()
	if err != nil {
		return err
	}

	serverName := d.Get("server_name").(string)
	name := d.Get("name").(string)

	_, err = client.Clients.Decommission(context.Background(), serverName, name)
	if err != nil {
		return err
	}

	return nil
}
