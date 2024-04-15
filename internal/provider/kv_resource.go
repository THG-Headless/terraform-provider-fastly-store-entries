package provider

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &KVStoreitemResource{}
var _ resource.ResourceWithImportState = &KVStoreitemResource{}

func NewKvStoreitemResource() resource.Resource {
	return &KVStoreitemResource{}
}

// KVStoreitemResource defines the resource implementation.
type KVStoreitemResource struct {
	client  *http.Client
	baseUrl string
	apiKey  string
}

// KVStoreitemResourceModel describes the resource data model.
type KVStoreitemResourceModel struct {
	StoreId  types.String `tfsdk:"store_id"`
	Key      types.String `tfsdk:"key"`
	Value    types.String `tfsdk:"value"`
	Metadata types.String `tfsdk:"metadata"`
}

type KVStoreAPIQueryParameters struct {
	failIfKeyExists bool
}

type KVStoreAPIHeaders struct {
	metadata string
}

func (r *KVStoreitemResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kv"
}

func (r *KVStoreitemResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "An item within a KV store.",

		Attributes: map[string]schema.Attribute{
			"store_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The ID of the store where the item will be contained. KV store names have a maximum length of 255 characters and may contain letters, numbers, dashes (-), underscores (_), and periods (.)",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"key": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The Key identifier for the KV store value. The maximum length is 1024 UTF-8 bytes.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"value": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The value which will be inserted. This value can have a maximum size of 25 MB.",
			},
			"metadata": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "An arbitrary data field which can contain up to 2000B of data",
			},
		},
	}
}

func (r *KVStoreitemResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	resourceData, ok := req.ProviderData.(*ConfiguredData)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *ConfiguredData, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = resourceData.client
	r.baseUrl = resourceData.baseUrl
	r.apiKey = resourceData.apiKey
}

func (r *KVStoreitemResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data KVStoreitemResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.updateKVStoreitem(
		ctx,
		data.StoreId.ValueString(),
		data.Key.ValueString(),
		data.Value.ValueString(),
		KVStoreAPIHeaders{
			metadata: data.Metadata.ValueString(),
		},
		KVStoreAPIQueryParameters{
			failIfKeyExists: true,
		},
	)

	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to create KV Store item",
			"An error occurred while executing the creation. "+
				"If unexpected, please report this issue to the provider developers.\n\n"+
				"JSON Error: "+err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *KVStoreitemResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data KVStoreitemResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	kvValue, err := r.getKVStoreitem(
		data.StoreId.ValueString(),
		data.Key.ValueString(),
	)

	if err != nil {
		resp.Diagnostics.AddError("Failed to get KV Store item", err.Error())
		return
	}

	data.Value = basetypes.NewStringValue(kvValue)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *KVStoreitemResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state KVStoreitemResourceModel

	// Read Terraform plan data into the model - new data
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	// Read Terraform plan data into the model - old data
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if plan.Value.Equal(state.Value) && plan.Metadata.Equal(state.Metadata) {
		resp.Diagnostics.AddWarning(
			"Skipping KV item Update.",
			"The old and new KV item values and metadata are the same. This would result in no physical change to the item.",
		)
	}

	err := r.updateKVStoreitem(
		ctx,
		plan.StoreId.ValueString(),
		plan.Key.ValueString(),
		plan.Value.ValueString(),
		KVStoreAPIHeaders{
			metadata: plan.Metadata.ValueString(),
		},
		KVStoreAPIQueryParameters{
			failIfKeyExists: false,
		},
	)

	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to update KV Store item",
			"An error occurred while executing the update. "+
				"If unexpected, please report this issue to the provider developers.\n\n"+
				"JSON Error: "+err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *KVStoreitemResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data KVStoreitemResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.deleteKVStoreitem(
		data.StoreId.ValueString(),
		data.Key.ValueString(),
	)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Delete Resource",
			"An unexpected error occurred while executing the request. "+
				"Please report this issue to the provider developers.\n\n"+
				"JSON Error: "+err.Error(),
		)
		return
	}
}

func (r *KVStoreitemResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *KVStoreitemResource) updateKVStoreitem(
	ctx context.Context,
	storeId string,
	key string,
	value string,
	headers KVStoreAPIHeaders,
	queryParameters KVStoreAPIQueryParameters,
) error {
	httpReq, err := http.NewRequest(
		http.MethodPut,
		fmt.Sprintf("%s/resources/stores/kv/%s/keys/%s", r.baseUrl, storeId, key),
		bytes.NewBuffer([]byte(value)),
	)

	if err != nil {
		return &KVStoreitemError{
			shortMessage: "Client Error",
			detail:       fmt.Sprintf("Unable to create http request, got error: %s", err),
		}
	}

	httpReq.Header.Add("Fastly-Key", r.apiKey)

	if headers.metadata != "" {
		httpReq.Header.Add("metadata", headers.metadata)
	}

	q := httpReq.URL.Query()
	if queryParameters.failIfKeyExists {
		q.Add("add", "true")
	}

	httpReq.URL.RawQuery = q.Encode()

	httpRes, err := r.client.Do(httpReq)

	if err != nil {
		return &KVStoreitemError{
			shortMessage: "HTTP Error",
			detail:       fmt.Sprintf("There has been an error with the http request, got error: %s", err),
		}
	}

	if httpRes.StatusCode == 412 {
		return &KVStoreitemError{
			shortMessage: "Key Conflict",
			detail:       "This key already exists within the KV store.",
		}
	}

	if httpRes.StatusCode == 404 {
		return &KVStoreitemError{
			shortMessage: "No KV Store Found",
			detail:       fmt.Sprintf("The KV store %s cannot be found within your account", storeId),
		}
	}

	if httpRes.StatusCode != 200 {
		defer httpRes.Body.Close()
		body, _ := io.ReadAll(httpRes.Body)
		tflog.Error(ctx, fmt.Sprintf("Body: %s", body))
		return &KVStoreitemError{
			shortMessage: "Unexpected Fastly API Response",
			detail:       fmt.Sprintf("The KV item Creation Request returned a non-200 response of %s.", httpRes.Status),
		}
	}

	return nil
}

func (r *KVStoreitemResource) deleteKVStoreitem(
	storeId string,
	key string,
) error {
	httpReq, err := http.NewRequest(
		http.MethodDelete,
		fmt.Sprintf("%s/resources/stores/kv/%s/keys/%s", r.baseUrl, storeId, key),
		nil,
	)

	if err != nil {
		return &KVStoreitemError{
			shortMessage: "Client Error",
			detail:       fmt.Sprintf("Unable to create http request, got error: %s", err),
		}
	}
	httpReq.Header.Add("Fastly-Key", r.apiKey)

	httpRes, err := r.client.Do(httpReq)

	if err != nil {
		return &KVStoreitemError{
			shortMessage: "HTTP Error",
			detail:       fmt.Sprintf("There has been an error with the http request, got error: %s", err),
		}
	}

	if httpRes.StatusCode == 404 {
		return &KVStoreitemError{
			shortMessage: "The key cannot be found within the KV store",
			detail:       fmt.Sprintf("Either the KV store %s cannot be found within your account or the key %s does not exist within the KV Store", storeId, key),
		}
	}

	if httpRes.StatusCode != 204 {
		return &KVStoreitemError{
			shortMessage: "Unexpected Fastly API Response",
			detail:       fmt.Sprintf("The KV item Creation Request returned a non-200 response of %s.", httpRes.Status),
		}
	}

	return nil
}

func (r *KVStoreitemResource) getKVStoreitem(
	storeId string,
	key string,
) (string, error) {
	httpReq, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("%s/resources/stores/kv/%s/keys/%s", r.baseUrl, storeId, key),
		nil,
	)

	if err != nil {
		return "", &KVStoreitemError{
			shortMessage: "Client Error",
			detail:       fmt.Sprintf("Unable to create http request, got error: %s", err),
		}
	}
	httpReq.Header.Add("Fastly-Key", r.apiKey)
	resp, err := r.client.Do(httpReq)

	if err != nil {
		return "", &KVStoreitemError{
			shortMessage: "HTTP Error",
			detail:       fmt.Sprintf("There has been an error with the http request, got error: %s", err),
		}
	}

	if resp.StatusCode == 404 {
		return "", &KVStoreitemError{
			shortMessage: "The key cannot be found within the KV store",
			detail:       fmt.Sprintf("Either the KV store %s cannot be found within your account or the key %s does not exist within the KV Store", storeId, key),
		}
	}

	if resp.StatusCode != 200 {
		return "", &KVStoreitemError{
			shortMessage: "Unexpected Fastly API Response",
			detail:       fmt.Sprintf("The KV item Creation Request returned a non-200 response of %s.", resp.Status),
		}
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", &KVStoreitemError{
			shortMessage: "Body Read Error",
			detail:       "Unable to read response body",
		}
	}

	return string(body[:]), nil
}
