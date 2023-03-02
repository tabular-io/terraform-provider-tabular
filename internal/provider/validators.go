package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/tabular-io/terraform-provider-tabular/internal/tabular"
	"golang.org/x/exp/slices"
)

type privilegeListValidator struct{}

var (
	_ validator.List = &privilegeListValidator{}
)

func (p privilegeListValidator) Description(ctx context.Context) string {
	return "Validate privileges"
}

func (p privilegeListValidator) MarkdownDescription(ctx context.Context) string {
	return p.Description(ctx)
}

func (p privilegeListValidator) ValidateList(ctx context.Context, req validator.ListRequest, resp *validator.ListResponse) {
	privileges := req.ConfigValue.Elements()
	for i, priv := range privileges {
		if priv.IsUnknown() {
			// Can't resolve this now; we'll have to wait and see
			continue
		}
		if priv.IsNull() {
			resp.Diagnostics.AddAttributeError(req.Path, "Invalid privileges", "Cannot specify null privilege")
			return
		}
		privValue, ok := priv.(basetypes.StringValue)

		if !ok {
			resp.Diagnostics.AddAttributeError(req.Path.AtListIndex(i), "Failed while extracting value", "")
		}
		if !slices.Contains(tabular.DatabasePrivileges, privValue.ValueString()) {
			resp.Diagnostics.AddAttributeError(
				req.Path.AtListIndex(i),
				"Invalid Database privilege",
				fmt.Sprintf("%s is not a valid privilege. Valid privileges are %s", privValue.ValueString(), tabular.DatabasePrivileges),
			)
		}
	}

}
