package credentials

import (
	"context"
	"maps"
	"os"
	"path"

	"github.com/BurntSushi/toml"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ function.Function = &Function{}

const DefaultPath = ".config/oxide/credentials.toml"

type credential struct {
	Host    string `toml:"host"     tfsdk:"host"`
	Token   string `toml:"token"    tfsdk:"token"`
	TokenID string `toml:"token_id" tfsdk:"token_id"`
	User    string `toml:"user"     tfsdk:"user"`
}

type credentials struct {
	Profiles map[string]credential `toml:"profile"`
}

func NewFunction() function.Function {
	return &Function{}
}

type Function struct{}

func (f *Function) Metadata(
	ctx context.Context,
	req function.MetadataRequest,
	resp *function.MetadataResponse,
) {
	resp.Name = "credentials"
}

func (f *Function) Definition(
	ctx context.Context,
	req function.DefinitionRequest,
	resp *function.DefinitionResponse,
) {
	resp.Definition = function.Definition{
		Summary: "Reads an Oxide credentials file.",
		MarkdownDescription: `
Reads Oxide API credentials from a local credentials file and returns a map of
credentials grouped by profile name. Refer to the [Oxide CLI
documentation](https://docs.oxide.computer/cli/guides/introduction) for more
information.
`,
		Parameters: []function.Parameter{
			function.StringParameter{
				Name:           "path",
				Description:    "Credentials file path. Defaults to `$HOME/.config/oxide/credentials.toml` if empty or `null`.",
				AllowNullValue: true,
			},
		},
		Return: function.MapReturn{
			ElementType: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"host":     types.StringType,
					"token":    types.StringType,
					"token_id": types.StringType,
					"user":     types.StringType,
				},
			},
		},
	}
}

func (f *Function) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	var pathInput types.String
	resp.Error = function.ConcatFuncErrors(resp.Error, req.Arguments.Get(ctx, &pathInput))

	pathStr := pathInput.ValueString()
	if pathStr == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			resp.Error = function.ConcatFuncErrors(resp.Error, function.NewFuncError(err.Error()))
			return
		}

		pathStr = path.Join(homeDir, DefaultPath)
	}

	var creds credentials
	if _, err := toml.DecodeFile(pathStr, &creds); err != nil {
		resp.Error = function.ConcatFuncErrors(resp.Error, function.NewFuncError(err.Error()))
		return
	}

	output := make(map[string]credential)
	maps.Copy(output, creds.Profiles)
	resp.Error = function.ConcatFuncErrors(resp.Error, resp.Result.Set(ctx, output))
}
