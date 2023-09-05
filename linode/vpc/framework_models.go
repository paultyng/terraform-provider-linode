package vpc

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/linode/linodego"
	"github.com/linode/terraform-provider-linode/linode/helper/customtypes"
	"github.com/linode/terraform-provider-linode/linode/vpcsubnet"
)

type VPCModel struct {
	ID          types.Int64                        `tfsdk:"id"`
	Label       types.String                       `tfsdk:"label"`
	Description types.String                       `tfsdk:"description"`
	Region      types.String                       `tfsdk:"region"`
	Subnets     []vpcsubnet.VPCSubnetModel         `tfsdk:"subnets"`
	Created     customtypes.RFC3339TimeStringValue `tfsdk:"created"`
	Updated     customtypes.RFC3339TimeStringValue `tfsdk:"updated"`
}

func (d *VPCModel) parseVPC(
	ctx context.Context,
	vpc *linodego.VPC,
) diag.Diagnostics {
	d.Label = types.StringValue(vpc.Label)
	d.Description = types.StringValue(vpc.Description)
	d.Region = types.StringValue(vpc.Region)

	if vpc.Created != nil {
		d.Created = customtypes.RFC3339TimeStringValue{
			StringValue: types.StringValue(vpc.Created.Format(time.RFC3339)),
		}
	}

	if vpc.Updated != nil {
		d.Updated = customtypes.RFC3339TimeStringValue{
			StringValue: types.StringValue(vpc.Updated.Format(time.RFC3339)),
		}
	}

	subnets := make([]vpcsubnet.VPCSubnetModel, len(vpc.Subnets))

	for i, subnet := range vpc.Subnets {
		var vpcSubnet vpcsubnet.VPCSubnetModel

		diag := vpcSubnet.ParseVPCSubnet(ctx, &subnet)
		if diag != nil {
			return diag
		}

		subnets[i] = vpcSubnet
	}

	d.Subnets = subnets

	return nil
}
