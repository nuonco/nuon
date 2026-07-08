package provider

import "github.com/hashicorp/terraform-plugin-framework/types"

// phoneHomeResourceModel is the Terraform shape for the stack_phone_home
// resource. The reported run outputs are carried opaquely as a JSON-encoded
// string in payload; the provider injects request_type (from the resource
// lifecycle) and phone_home_type before POSTing.
type phoneHomeResourceModel struct {
	ID            types.String `tfsdk:"id"`
	InstallID     types.String `tfsdk:"install_id"`
	PhoneHomeID   types.String `tfsdk:"phone_home_id"`
	PhoneHomeType types.String `tfsdk:"phone_home_type"`
	Payload       types.String `tfsdk:"payload"`
}
