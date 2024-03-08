package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Provider will create the necessary terraform provider to talk to the
// Bitbucket APIs you should either specify Username and App Password, OAuth
// Client Credentials or a valid OAuth Access Token.
//
// See the Bitbucket authentication documentation for more:
// https://developer.atlassian.com/cloud/bitbucket/rest/intro/#authentication
func newProvider() *schema.Provider {
	return &schema.Provider{
		Schema:        map[string]*schema.Schema{},
		ConfigureFunc: nil,
		ResourcesMap:  map[string]*schema.Resource{},
		DataSourcesMap: map[string]*schema.Resource{
			"collection_ip_ranges": dataIPRanges(),
		},
	}
}
