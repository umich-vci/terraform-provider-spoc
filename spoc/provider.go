package spoc

import (
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	config := &Config{
		Username:     d.Get("username").(string),
		Password:     d.Get("password").(string),
		SPOCEndpoint: d.Get("spoc_endpoint").(string),
		URLScheme:    d.Get("url_scheme").(string),
		SSLVerify:    d.Get("ssl_verify").(bool),
	}

	return config, nil
}

// Provider returns a terraform resource provider
func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"username": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("SPOC_USERNAME", nil),
				Description: "A Spectrum Protect Operations Center username.",
			},
			"password": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("SPOC_PASSWORD", nil),
				Description: "The Spectrum Protect Operations Center password.",
			},
			"spoc_endpoint": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("SPOC_ENDPOINT", nil),
				Description: "The Spectrum Protect Operations Center endpoint hostname/port",
			},
			"url_scheme": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The Spectrum Protect Operations Center REST API Schema",
			},
			"ssl_verify": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Perform SSL verification",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"spoc_client": resourceClient(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"spoc_client": dataSourceClient(),
		},
		ConfigureFunc: providerConfigure,
	}
}

func parseYesNoBool(val string) (bool, error) {
	switch val {
	case "Yes",
		"yes",
		"YES":
		return true, nil
	case "No",
		"no",
		"NO":
		return false, nil
	default:
		return false, fmt.Errorf("Unable to parse %s with parseYesNoBool", val)
	}
}
