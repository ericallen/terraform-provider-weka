package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"net/http"
	"strconv"
	"time"
)

func resourceUserPolicy() *schema.Resource {
	return &schema.Resource{
		ReadContext:   resourceUserPolicyRead,
		CreateContext: resourceUserPolicyCreate,
		UpdateContext: resourceUserPolicyUpdate,
		DeleteContext: resourceUserPolicyDelete,
		Schema: map[string]*schema.Schema{
			"username": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"s3_policy_name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"last_updated": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func resourceUserPolicyRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return nil
}

func resourceUserPolicyDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	c := m.(*WekaClient)

	delDoc := make(map[string]interface{})
	delDoc["user_name"] = d.Get("username")

	url := c.makeRestEndpointURL("/s3/policies/detach")
	payload, err := json.Marshal(delDoc)

	if err != nil {
		return diag.FromErr(err)
	}

	req, err := http.NewRequest("POST", url.String(), bytes.NewBuffer(payload))

	if err != nil {
		return diag.FromErr(err)
	}

	_, err = c.makeRequest(req)

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return diags
}

func resourceUserPolicyUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// if the username changed, we have to detach the policy from the
	// user and _attach_ to the new user (i.e call Delete and Create)
	if d.HasChange("username") {
		diags = resourceUserPolicyDelete(ctx, d, m)

		if diags != nil && diags.HasError() {
			return diags
		}
		diags = resourceUserPolicyCreate(ctx, d, m)
		// ... and if the policy changed attach the new policy (i.e call Create)
	} else if d.HasChange("s3_policy_name") {
		diags = resourceUserPolicyCreate(ctx, d, m)
	}

	if diags != nil && diags.HasError() {
		return diags
	}

	d.Set("last_updated", time.Now().Format(time.RFC850))
	return diags
}

func resourceUserPolicyCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	c := m.(*WekaClient)

	createData := map[string]interface{}{
		"user_name":   d.Get("username").(string),
		"policy_name": d.Get("s3_policy_name").(string),
	}

	createBody, err := json.Marshal(createData)

	if err != nil {
		return diag.FromErr(err)
	}

	url := c.makeRestEndpointURL("/s3/policies/attach")
	req, err := http.NewRequest("POST", url.String(), bytes.NewBuffer(createBody))

	_, err = c.makeRequest(req)

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}
