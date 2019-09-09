package main

import (
	"github.com/hashicorp/terraform/plugin"
	"github.com/umich-vci/terraform-provider-spoc/spoc"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: spoc.Provider})
}
