package azurerm

import (
	"fmt"
	"log"

	"github.com/Azure/azure-sdk-for-go/arm/automation"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

func resourceArmAutomationCredential() *schema.Resource {
	return &schema.Resource{
		Create: resourceArmAutomationCredentialCreateUpdate,
		Read:   resourceArmAutomationCredentialRead,
		Update: resourceArmAutomationCredentialCreateUpdate,
		Delete: resourceArmAutomationCredentialDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"resource_group_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"account_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"username": {
				Type:     schema.TypeString,
				Required: true,
			},

			"password": {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
			},
			"description": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceArmAutomationCredentialCreateUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).automationCredentialClient
	log.Printf("[INFO] preparing arguments for AzureRM Automation Credential creation.")

	name := d.Get("name").(string)
	resGroup := d.Get("resource_group_name").(string)
	accName := d.Get("account_name").(string)
	user := d.Get("username").(string)
	password := d.Get("password").(string)
	description := d.Get("description").(string)

	parameters := automation.CredentialCreateOrUpdateParameters{
		CredentialCreateOrUpdateProperties: &automation.CredentialCreateOrUpdateProperties{
			UserName:    &user,
			Password:    &password,
			Description: &description,
		},
		Name: &name,
	}

	_, err := client.CreateOrUpdate(resGroup, accName, name, parameters)
	if err != nil {
		return err
	}

	read, err := client.Get(resGroup, accName, name)
	if err != nil {
		return err
	}

	if read.ID == nil {
		return fmt.Errorf("Cannot read Automation Credential '%s' (resource group %s) ID", name, resGroup)
	}

	d.SetId(*read.ID)

	return resourceArmAutomationAccountRead(d, meta)
}

func resourceArmAutomationCredentialRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).automationCredentialClient
	id, err := parseAzureResourceID(d.Id())
	if err != nil {
		return err
	}
	resGroup := id.ResourceGroup
	accName := id.Path["automationAccounts"]
	name := id.Path["credentials"]

	resp, err := client.Get(resGroup, accName, name)

	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			d.SetId("")
			return nil
		}

		return fmt.Errorf("Error making Read request on AzureRM Automation Credential '%s': %+v", name, err)
	}

	d.Set("name", resp.Name)
	d.Set("resource_group_name", resGroup)
	d.Set("account_name", accName)
	if props := resp.CredentialProperties; props != nil {
		d.Set("username", props.UserName)
	}
	d.Set("description", resp.Description)

	return nil
}

func resourceArmAutomationCredentialDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).automationCredentialClient

	id, err := parseAzureResourceID(d.Id())
	if err != nil {
		return err
	}
	resGroup := id.ResourceGroup
	accName := id.Path["automationAccounts"]
	name := id.Path["credentials"]

	resp, err := client.Delete(resGroup, accName, name)

	if err != nil {
		if utils.ResponseWasNotFound(resp) {
			return nil
		}

		return fmt.Errorf("Error issuing AzureRM delete request for Automation Credential '%s': %+v", name, err)
	}

	return nil
}
