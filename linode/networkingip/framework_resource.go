package networkingip

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/linode/linodego"
	"github.com/linode/terraform-provider-linode/v2/linode/helper"
)

func NewResource() resource.Resource {
	return &Resource{
		BaseResource: helper.NewBaseResource(
			helper.BaseResourceConfig{
				Name:   "linode_networking_ip",
				IDType: types.StringType,
				Schema: &frameworkResourceSchema,
			},
		),
	}
}

type Resource struct {
	helper.BaseResource
}

func (r *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Debug(ctx, "Create linode_networking_ip")
	var plan NetworkingIPModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client := r.Meta.Client

	createOpts := linodego.AllocateReserveIPOptions{
		Type:   plan.Type.ValueString(),
		Public: plan.Public.ValueBool(),
	}

	if !plan.LinodeID.IsNull() {
		createOpts.LinodeID = int(plan.LinodeID.ValueInt64())
	}
	if !plan.Reserved.IsNull() {
		createOpts.Reserved = plan.Reserved.ValueBool()
	}
	if !plan.Region.IsNull() {
		createOpts.Region = plan.Region.ValueString()
	}

	ip, err := client.AllocateReserveIP(ctx, createOpts)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating IP Address",
			fmt.Sprintf("Could not create IP address: %s", err),
		)
		return
	}

	plan.FlattenIPAddress(ctx, ip, false)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Debug(ctx, "Read linode_networking_ip")
	var state NetworkingIPModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client := r.Meta.Client

	ip, err := client.GetIPAddress(ctx, state.ID.ValueString())
	if err != nil {
		if lerr, ok := err.(*linodego.Error); ok && lerr.Code == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error reading IP Address",
			fmt.Sprintf("Could not read IP address %s: %s", state.ID.ValueString(), err),
		)
		return
	}

	state.FlattenIPAddress(ctx, ip, false)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	tflog.Debug(ctx, "Update linode_networking_ip")
	var plan, state NetworkingIPModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client := r.Meta.Client

	var reservedValue *bool
	if plan.Reserved != state.Reserved {
		value := plan.Reserved.ValueBoolPointer()
		reservedValue = value
	}

	updateOpts := linodego.IPAddressUpdateOptions{
		Reserved: reservedValue,
	}

	ip, err := client.UpdateIPAddress(ctx, state.Address.ValueString(), updateOpts)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Update IP Address",
			fmt.Sprintf("Could not update reserved status of IP address: %s", err),
		)
		return
	}

	plan.FlattenIPAddress(ctx, ip, false)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Debug(ctx, "Delete linode_networking_ip")
	var state NetworkingIPModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client := r.Meta.Client

	// Regular assigned ephemeral IP address
	if !state.Reserved.ValueBool() {
		// This is a regular ephemeral IP address
		linodeID := helper.FrameworkSafeInt64ToInt(state.LinodeID.ValueInt64(), &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}

		// Check if this is the only public IP on the Linode
		ips, err := client.GetInstanceIPAddresses(ctx, linodeID)
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to List IPs",
				fmt.Sprintf("failed to list IP addresses for Linode (%d): %s", linodeID, err.Error()),
			)
			return
		}

		// If this is the only public IP, skip deletion to avoid the "must have at least one public IP" error
		if len(ips.IPv4.Public) == 1 {
			resp.Diagnostics.AddWarning(
				"Cannot Delete Last IP",
				"Linode must have at least one public IP address. The last IP cannot be deleted.",
			)
			return
		}

		// Proceed with deleting the IP if it's not the only one
		err = client.DeleteInstanceIPAddress(ctx, linodeID, state.Address.ValueString())
		if err != nil {
			if lErr, ok := err.(*linodego.Error); (ok && lErr.Code != 404) || !ok {
				resp.Diagnostics.AddError(
					"Failed to Delete IP",
					fmt.Sprintf(
						"failed to delete instance (%d) ip (%s): %s",
						linodeID, state.Address.ValueString(), err.Error(),
					),
				)
			}
		}
	} else {
		// Reserved IP address
		// If the IP is currently assigned (reserved but used)
		if state.LinodeID.ValueInt64() != 0 {
			// It's an assigned reserved IP, we can delete it regardless of being the only IP
			linodeID := helper.FrameworkSafeInt64ToInt(state.LinodeID.ValueInt64(), &resp.Diagnostics)
			if resp.Diagnostics.HasError() {
				return
			}

			// Delete the reserved IP (this will turn it into an ephemeral IP if it's the only IP)
			err := client.DeleteReservedIPAddress(ctx, state.Address.ValueString())
			if err != nil {
				resp.Diagnostics.AddError(
					"Failed to Delete Assigned Reserved IP",
					fmt.Sprintf(
						"failed to delete assigned reserved ip (%s) from linode (%d): %s",
						state.Address.ValueString(), linodeID, err.Error(),
					),
				)
			}
			return
		} else {
			// Reserved IP (unassigned) that needs to be deleted
			// If it's a reserved IP but it is not assigned to a Linode, proceed with deletion
			err := client.DeleteReservedIPAddress(ctx, state.Address.ValueString())
			if err != nil {
				resp.Diagnostics.AddError(
					"Failed to Delete Reserved IP",
					fmt.Sprintf(
						"failed to delete reserved ip (%s): %s",
						state.Address.ValueString(), err.Error(),
					),
				)
			}
		}
	}
}