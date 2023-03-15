package validators

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/tabular-io/terraform-provider-tabular/internal/tabular"
	"golang.org/x/exp/slices"
)

type PrivilegeSetValidator struct{}

var (
	_ validator.Set = &PrivilegeSetValidator{}
)

func (p PrivilegeSetValidator) Description(ctx context.Context) string {
	return "Validate privileges"
}

func (p PrivilegeSetValidator) MarkdownDescription(ctx context.Context) string {
	return p.Description(ctx)
}

func (p PrivilegeSetValidator) ValidateSet(ctx context.Context, req validator.SetRequest, resp *validator.SetResponse) {
	privileges := req.ConfigValue.Elements()
	for _, priv := range privileges {
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
			resp.Diagnostics.AddAttributeError(req.Path.AtSetValue(priv), "Failed while extracting value", "")
		}
		if !slices.Contains(tabular.DatabasePrivileges, privValue.ValueString()) {
			resp.Diagnostics.AddAttributeError(
				req.Path.AtSetValue(priv),
				"Invalid Database privilege",
				fmt.Sprintf("%s is not a valid privilege. Valid privileges are %s", privValue.ValueString(), tabular.DatabasePrivileges),
			)
		}
	}

}
