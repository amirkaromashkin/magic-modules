package google

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ProviderModel describes the provider data model.
type ProviderModel struct {
	Credentials                        types.String `tfsdk:"credentials"`
	AccessToken                        types.String `tfsdk:"access_token"`
	ImpersonateServiceAccount          types.String `tfsdk:"impersonate_service_account"`
	ImpersonateServiceAccountDelegates types.List   `tfsdk:"impersonate_service_account_delegates"`
	Project                            types.String `tfsdk:"project"`
	BillingProject                     types.String `tfsdk:"billing_project"`
	Region                             types.String `tfsdk:"region"`
	Zone                               types.String `tfsdk:"zone"`
	Scopes                             types.List   `tfsdk:"scopes"`
	Batching                           types.List   `tfsdk:"batching"`
	UserProjectOverride                types.Bool   `tfsdk:"user_project_override"`
	RequestTimeout                     types.String `tfsdk:"request_timeout"`
	RequestReason                      types.String `tfsdk:"request_reason"`

	// Generated Products
	ComputeCustomEndpoint types.String `tfsdk:"compute_custom_endpoint"`

	// Handwritten Products / Versioned / Atypical Entries
	CloudBillingCustomEndpoint      types.String `tfsdk:"cloud_billing_custom_endpoint"`
	ComposerCustomEndpoint          types.String `tfsdk:"composer_custom_endpoint"`
	ContainerCustomEndpoint         types.String `tfsdk:"container_custom_endpoint"`
	DataflowCustomEndpoint          types.String `tfsdk:"dataflow_custom_endpoint"`
	IamCredentialsCustomEndpoint    types.String `tfsdk:"iam_credentials_custom_endpoint"`
	ResourceManagerV3CustomEndpoint types.String `tfsdk:"resource_manager_v3_custom_endpoint"`
	RuntimeconfigCustomEndpoint     types.String `tfsdk:"runtimeconfig_custom_endpoint"`
	IAMCustomEndpoint               types.String `tfsdk:"iam_custom_endpoint"`
	ServiceNetworkingCustomEndpoint types.String `tfsdk:"service_networking_custom_endpoint"`
	TagsLocationCustomEndpoint      types.String `tfsdk:"tags_location_custom_endpoint"`

	// dcl
	ContainerAwsCustomEndpoint   types.String `tfsdk:"container_aws_custom_endpoint"`
	ContainerAzureCustomEndpoint types.String `tfsdk:"container_azure_custom_endpoint"`

	// dcl generated
	ApikeysCustomEndpoint              types.String `tfsdk:"apikeys_custom_endpoint"`
	AssuredWorkloadsCustomEndpoint     types.String `tfsdk:"assured_workloads_custom_endpoint"`
	CloudBuildWorkerPoolCustomEndpoint types.String `tfsdk:"cloud_build_worker_pool_custom_endpoint"`
	CloudDeployCustomEndpoint          types.String `tfsdk:"clouddeploy_custom_endpoint"`
	CloudResourceManagerCustomEndpoint types.String `tfsdk:"cloud_resource_manager_custom_endpoint"`
	EventarcCustomEndpoint             types.String `tfsdk:"eventarc_custom_endpoint"`
	FirebaserulesCustomEndpoint        types.String `tfsdk:"firebaserules_custom_endpoint"`
	NetworkConnectivityCustomEndpoint  types.String `tfsdk:"network_connectivity_custom_endpoint"`
	RecaptchaEnterpriseCustomEndpoint  types.String `tfsdk:"recaptcha_enterprise_custom_endpoint"`
	GkehubFeatureCustomEndpoint        types.String `tfsdk:"gkehub_feature_custom_endpoint"`
}

type ProviderBatching struct {
	SendAfter      types.String `tfsdk:"send_after"`
	EnableBatching types.Bool   `tfsdk:"enable_batching"`
}

// ProviderMetaModel describes the provider meta model
type ProviderMetaModel struct {
	ModuleName types.String `tfsdk:"module_name"`
}
