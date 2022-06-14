package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/umich-vci/gospoc"
)

func init() {
	// Set descriptions to support markdown syntax, this will be used in document generation
	// and the language server.
	schema.DescriptionKind = schema.StringMarkdown

	// Customize the content of descriptions when output. For example you can add defaults on
	// to the exported descriptions if present.
	schema.SchemaDescriptionBuilder = func(s *schema.Schema) string {
		desc := s.Description
		if s.Default != nil {
			desc += fmt.Sprintf(" Defaults to `%v`.", s.Default)
		}
		return strings.TrimSpace(desc)
	}
}

func New(version string) func() *schema.Provider {
	return func() *schema.Provider {
		p := &schema.Provider{
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
		}

		p.ConfigureContextFunc = configure(version, p)

		return p
	}
}

type apiClient struct {
	Client gospoc.Client
}

func configure(version string, p *schema.Provider) func(context.Context, *schema.ResourceData) (interface{}, diag.Diagnostics) {
	return func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		userAgent := p.UserAgent("terraform-provider-spoc", version)
		username := d.Get("username").(string)
		password := d.Get("password").(string)
		endpoint := d.Get("spoc_endpoint").(string)
		urlScheme := d.Get("url_scheme").(string)
		sslVerify := d.Get("ssl_verify").(bool)

		config := new(gospoc.Config)
		config.Username = username
		config.Password = password
		config.OCHost = endpoint
		config.APIVersion = urlScheme
		config.SSLVerify = sslVerify

		client, err := gospoc.NewClient(config)
		if err != nil {
			return nil, diag.FromErr(err)
		}

		client.UserAgent = userAgent

		return &apiClient{Client: *client}, nil
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
