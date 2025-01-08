package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type PaginatedIPRanges struct {
	Items     []IPRange `json:"items,omitempty"`
	SyncToken int       `json:"syncToken,omitempty"`
}

type IPRange struct {
	Network    string   `json:"network"`
	MaskLen    int      `json:"mask_len"`
	CIDR       string   `json:"cidr"`
	Mask       string   `json:"mask"`
	Regions    []string `json:"region"`
	Products   []string `json:"product"`
	Directions []string `json:"direction"`
	Perimeter  string   `json:"perimeter"`
}

func dataIPRanges() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataReadIPRanges,

		Schema: map[string]*schema.Schema{
			"ipv4_cidr_blocks": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"ipv6_cidr_blocks": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"cidr_blocks": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataReadIPRanges(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	req, err := http.Get("https://ip-ranges.atlassian.com/")
	if err != nil {
		return diag.FromErr(err)
	}

	if req.StatusCode == http.StatusNotFound {
		return diag.Errorf("IP whitelist not found")
	}

	body, readerr := io.ReadAll(req.Body)
	if readerr != nil {
		return diag.FromErr(readerr)
	}

	log.Printf("[DEBUG] IP Ranges Response JSON: %v", string(body))

	var pageIpRanges PaginatedIPRanges

	decodeerr := json.Unmarshal(body, &pageIpRanges)
	if decodeerr != nil {
		return diag.FromErr(decodeerr)
	}

	log.Printf("[DEBUG] IP Ranges Decoded: %#v", pageIpRanges)

	d.SetId(fmt.Sprintf("%d", pageIpRanges.SyncToken))
	iPRanges := flattenIPRanges(pageIpRanges.Items)
	d.Set("ipv4_cidr_blocks", iPRanges["ipv4"])
	d.Set("ipv6_cidr_blocks", iPRanges["ipv6"])
	d.Set("cidr_blocks", iPRanges["cird_blocks"])
	return nil
}

func flattenIPRanges(ranges []IPRange) (tfList map[string][]interface{}) {
	if len(ranges) == 0 {
		return nil
	}
	tfList = make(map[string][]interface{})
	
	for _, btRaw := range ranges {
		log.Printf("[DEBUG] IP Range Response Decoded: %#v", btRaw)
		if anyElementInList([]string{"ingress"},btRaw.Directions) {
			if stringInList([]string{"confluence","jira"},btRaw.Products) {
				ip := btRaw.Network
				if isIPv4(ip) {
					v4List := tfList["ipv4"]
					v4List = append(v4List, btRaw.CIDR)
					tfList["ipv4"] = v4List
				}
				if isIPv6(ip) {
					v6List := tfList["ipv6"]
					v6List = append(v6List, btRaw.CIDR)
					tfList["ipv6"] = v6List
				}
				cird_blocks := tfList["cird_blocks"]
				cird_blocks = append(cird_blocks, btRaw.CIDR)
				tfList["cird_blocks"] = cird_blocks
			}
		
	}

	return tfList
}
func isIPv4(ip string) bool {
	return net.ParseIP(ip) != nil && net.ParseIP(ip).To4() != nil
}

func isIPv6(ip string) bool {
	return net.ParseIP(ip) != nil && net.ParseIP(ip).To4() == nil
}
func anyElementInList(list1, list2 []string) bool {
	for _, item1 := range list1 {
		for _, item2 := range list2 {
			if item1 == item2 {
				return true
			}
		}
	}
	return false
}