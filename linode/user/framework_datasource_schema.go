package user

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var linodeUserGrantsGlobalObjectType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"account_access":        types.StringType,
		"add_domains":           types.BoolType,
		"add_databases":         types.BoolType,
		"add_firewalls":         types.BoolType,
		"add_images":            types.BoolType,
		"add_linodes":           types.BoolType,
		"add_longview":          types.BoolType,
		"add_nodebalancers":     types.BoolType,
		"add_placement_groups":  types.BoolType,
		"add_stackscripts":      types.BoolType,
		"add_volumes":           types.BoolType,
		"add_vpcs":              types.BoolType,
		"cancel_account":        types.BoolType,
		"longview_subscription": types.BoolType,
	},
}

var linodeUserGrantsEntityObjectType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"id":          types.Int64Type,
		"permissions": types.StringType,
		"label":       types.StringType,
	},
}

var linodeUserGrantsEntitySet = schema.SetAttribute{
	Description: "A set containing all of the user's active grants.",
	Computed:    true,
	Optional:    true,
	ElementType: linodeUserGrantsEntityObjectType,
}

var UserAttributes = map[string]schema.Attribute{
	"username": schema.StringAttribute{
		Description: "This User's username. This is used for logging in, and may also be displayed alongside " +
			"actions the User performs (for example, in Events or public StackScripts).",
		Required: true,
	},
	"ssh_keys": schema.ListAttribute{
		Description: "A list of SSH Key labels added by this User. These are the keys that will be deployed " +
			"if this User is included in the authorized_users field of a create Linode, rebuild Linode, or " +
			"create Disk request.",
		Computed:    true,
		ElementType: types.StringType,
	},
	"email": schema.StringAttribute{
		Description: "The email address for this User, for account management communications, and may be used " +
			"for other communications as configured.",
		Computed: true,
	},
	"restricted": schema.BoolAttribute{
		Description: "If true, this User must be granted access to perform actions or access entities on this Account.",
		Computed:    true,
	},
	"user_type": schema.StringAttribute{
		Description: "The type of this user.",
		Computed:    true,
	},
	"global_grants": schema.ListAttribute{
		Description: "A structure containing the Account-level grants a User has.",
		Computed:    true,
		ElementType: linodeUserGrantsGlobalObjectType,
	},
	"database_grant":        linodeUserGrantsEntitySet,
	"domain_grant":          linodeUserGrantsEntitySet,
	"firewall_grant":        linodeUserGrantsEntitySet,
	"image_grant":           linodeUserGrantsEntitySet,
	"linode_grant":          linodeUserGrantsEntitySet,
	"longview_grant":        linodeUserGrantsEntitySet,
	"nodebalancer_grant":    linodeUserGrantsEntitySet,
	"placement_group_grant": linodeUserGrantsEntitySet,
	"stackscript_grant":     linodeUserGrantsEntitySet,
	"volume_grant":          linodeUserGrantsEntitySet,
	"vpc_grant":             linodeUserGrantsEntitySet,
	"id": schema.StringAttribute{
		Description: "Unique identifier for this DataSource.",
		Computed:    true,
	},
	"password_created": schema.StringAttribute{
		Description: "The date and time when this User’s current password was created." +
			"User passwords are first created during the Account sign-up process, " +
			"and updated using the Reset Password webpage." +
			"null if this User has not created a password yet.",
		Computed: true,
	},
	"tfa_enabled": schema.BoolAttribute{
		Description: "A boolean value indicating if the User has Two Factor Authentication (TFA) enabled.",
		Computed:    true,
	},
	"verified_phone_number": schema.StringAttribute{
		Description: "The phone number verified for this User Profile with the Phone Number Verify command." +
			"null if this User Profile has no verified phone number.",
		Computed: true,
	},
}

var frameworkDatasourceSchema = schema.Schema{
	Attributes: UserAttributes,
}
