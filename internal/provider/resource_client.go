package provider

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/umich-vci/gospoc"
)

const (
	reservedCharactersWarning = "May not contain ' ', '*', or '?'."
	nameRegexWarning          = "May only contain alphanumeric characters, '.', or '-'."
)

var nameRegex, _ = regexp.Compile(`^[a-zA-Z0-9\.\-]*$`)

var reservedCharacters, _ = regexp.Compile("^[^ *?]*$")

func resourceClient() *schema.Resource {
	return &schema.Resource{
		Create: resourceClientCreate,
		Read:   resourceClientRead,
		Update: resourceClientUpdate,
		Delete: resourceClientDelete,
		Schema: map[string]*schema.Schema{
			"server_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(reservedCharacters, reservedCharactersWarning),
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(nameRegex, nameRegexWarning),
			},
			"authentication": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "Local",
				ValidateFunc: validation.StringInSlice([]string{"Local", "LDAP"}, false),
				ForceNew:     true,
			},
			"password": {
				Type:         schema.TypeString,
				Required:     true,
				Sensitive:    true,
				ValidateFunc: validation.StringMatch(reservedCharacters, reservedCharactersWarning),
			},
			"domain": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringMatch(reservedCharacters, reservedCharactersWarning),
			},
			"contact": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(reservedCharacters, reservedCharactersWarning),
			},
			"email": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(reservedCharacters, reservedCharactersWarning),
			},
			"schedule": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringMatch(reservedCharacters, reservedCharactersWarning),
			},
			"option_set": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(reservedCharacters, reservedCharactersWarning),
			},
			"deduplication": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      "ClientOrServer",
				ValidateFunc: validation.StringInSlice([]string{"ClientOrServer", "ServerOnly"}, false),
			},
			"ssl_required": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      "Default",
				ValidateFunc: validation.StringInSlice([]string{"Default", "YES", "NO"}, false),
			},
			"session_initiation": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      "ClientOrServer",
				ValidateFunc: validation.StringInSlice([]string{"ClientOrServer", "ServerOnly"}, false),
			},
			"locked": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"link": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"decommissioned": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"split_large_objects": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

func resourceClientRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*apiClient).Client

	id := d.Id()
	name := strings.Split(id, "/")[1]
	serverName := strings.Split(id, "/")[0]

	backupClientDetails, resp, err := client.Clients.Details(context.Background(), serverName, name)
	if err != nil {
		if backupClientDetails == nil && resp.StatusCode == 200 {
			d.SetId("")
			return nil
		}
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
	client := meta.(*apiClient).Client

	serverName := d.Get("server_name").(string)
	name := d.Get("name").(string)

	_, _, err := client.Clients.Details(context.Background(), serverName, name)
	if err == nil {
		return fmt.Errorf("a node with name %s already exists on server %s", name, serverName)
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

	id := serverName + "/" + name
	d.SetId(id)

	return resourceClientRead(d, meta)
}

func resourceClientUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*apiClient).Client

	id := d.Id()
	name := strings.Split(id, "/")[1]
	serverName := strings.Split(id, "/")[0]

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

	if d.HasChanges("schedule", "domain") {
		schedule := d.Get("schedule").(string)
		domain := d.Get("domain").(string)
		_, err := client.Clients.AssignSchedule(context.Background(), serverName, name, domain, schedule)
		if err != nil {
			return err
		}
	}

	// if d.HasChange("name") {
	// 	newName := d.Get("name").(string)
	// 	command := "RENAME NODE " + name + " " + newName
	// 	client.CLI.IssueCommand(context.Background(), serverName, command)
	// }

	return resourceClientRead(d, meta)
}

func resourceClientDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*apiClient).Client

	id := d.Id()
	name := strings.Split(id, "/")[1]
	serverName := strings.Split(id, "/")[0]

	_, err := client.Clients.Decommission(context.Background(), serverName, name)
	if err != nil {
		return err
	}

	return nil
}
