package image

import (
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/linode/linodego"
	"github.com/stretchr/testify/assert"
)

func TestParseImage(t *testing.T) {
	createdTime := &time.Time{}
	createdTimeFormatted := createdTime.Format(time.RFC3339)
	mockImage := linodego.Image{
		ID:           "linode/debian11",
		CreatedBy:    "linode",
		Capabilities: []string{},
		Label:        "Debian 11",
		Description:  "Example Image description.",
		Type:         "manual",
		Vendor:       "Debian",
		Status:       "available",
		Size:         2500,
		IsPublic:     true,
		Deprecated:   false,
		Created:      createdTime,
		Expiry:       nil,
	}

	var imageModel ImageModel
	imageModel.ParseImage(&mockImage)

	assert.Equal(t, types.StringValue("linode/debian11"), imageModel.ID)
	assert.Equal(t, types.StringValue("linode"), imageModel.CreatedBy)
	assert.Empty(t, imageModel.Capabilities)
	assert.Equal(t, types.StringValue("Debian 11"), imageModel.Label)
	assert.Equal(t, types.StringValue("Example Image description."), imageModel.Description)
	assert.Equal(t, types.StringValue("manual"), imageModel.Type)
	assert.Equal(t, types.StringValue("Debian"), imageModel.Vendor)
	assert.Equal(t, types.StringValue("available"), imageModel.Status)
	assert.Equal(t, types.Int64Value(2500), imageModel.Size)
	assert.Equal(t, types.BoolValue(true), imageModel.IsPublic)
	assert.Equal(t, types.BoolValue(false), imageModel.Deprecated)
	assert.Equal(t, imageModel.Created, types.StringValue(createdTimeFormatted))
	assert.Empty(t, imageModel.Expiry)
}
