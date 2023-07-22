package google

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"golang.org/x/oauth2"
	googleoauth "golang.org/x/oauth2/google"

	"google.golang.org/api/option"
	"google.golang.org/api/transport"
	"google.golang.org/grpc"

	"github.com/hashicorp/go-cleanhttp"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/logging"

	grpc_logrus "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	"github.com/sirupsen/logrus"
)

// Provider methods

// LoadAndValidateFramework handles the bulk of configuring the provider
// it is pulled out so that we can manually call this from our testing provider as well
func (p *frameworkProvider) LoadAndValidateFramework(ctx context.Context, data ProviderModel, tfVersion string, diags *diag.Diagnostics) {
	// Set defaults if needed
	p.HandleDefaults(ctx, &data, diags)
	if diags.HasError() {
		return
	}

	p.context = ctx

	// Handle User Agent string
	p.userAgent = CompileUserAgentString(ctx, "terraform-provider-google-beta", tfVersion, p.version)
	// opt in extension for adding to the User-Agent header
	if ext := os.Getenv("GOOGLE_TERRAFORM_USERAGENT_EXTENSION"); ext != "" {
		ua := p.userAgent
		p.userAgent = fmt.Sprintf("%s %s", ua, ext)
	}

	// Set up client configuration
	p.SetupClient(ctx, data, diags)
	if diags.HasError() {
		return
	}

	// gRPC Logging setup
	p.SetupGrpcLogging()

	// Handle Batching Config
	batchingConfig := GetBatchingConfig(ctx, data.Batching, diags)
	if diags.HasError() {
		return
	}

	// Setup Base Paths for clients
	// Generated products
	p.ComputeBasePath = data.ComputeCustomEndpoint.ValueString()

	p.context = ctx
	p.region = data.Region
	p.zone = data.Zone
	p.pollInterval = 10 * time.Second
	p.project = data.Project
	p.requestBatcherServiceUsage = NewRequestBatcher("Service Usage", ctx, batchingConfig)
	p.requestBatcherIam = NewRequestBatcher("IAM", ctx, batchingConfig)
}

// HandleDefaults will handle all the defaults necessary in the provider
func (p *frameworkProvider) HandleDefaults(ctx context.Context, data *ProviderModel, diags *diag.Diagnostics) {
	if data.AccessToken.IsNull() && data.Credentials.IsNull() {
		credentials := MultiEnvDefault([]string{
			"GOOGLE_CREDENTIALS",
			"GOOGLE_CLOUD_KEYFILE_JSON",
			"GCLOUD_KEYFILE_JSON",
		}, nil)

		if credentials != nil {
			data.Credentials = types.StringValue(credentials.(string))
		}

		accessToken := MultiEnvDefault([]string{
			"GOOGLE_OAUTH_ACCESS_TOKEN",
		}, nil)

		if accessToken != nil {
			data.AccessToken = types.StringValue(accessToken.(string))
		}
	}

	if data.ImpersonateServiceAccount.IsNull() && os.Getenv("GOOGLE_IMPERSONATE_SERVICE_ACCOUNT") != "" {
		data.ImpersonateServiceAccount = types.StringValue(os.Getenv("GOOGLE_IMPERSONATE_SERVICE_ACCOUNT"))
	}

	if data.Project.IsNull() {
		project := MultiEnvDefault([]string{
			"GOOGLE_PROJECT",
			"GOOGLE_CLOUD_PROJECT",
			"GCLOUD_PROJECT",
			"CLOUDSDK_CORE_PROJECT",
		}, nil)
		if project != nil {
			data.Project = types.StringValue(project.(string))
		}
	}

	if data.BillingProject.IsNull() && os.Getenv("GOOGLE_BILLING_PROJECT") != "" {
		data.BillingProject = types.StringValue(os.Getenv("GOOGLE_BILLING_PROJECT"))
	}

	if data.Region.IsNull() {
		region := MultiEnvDefault([]string{
			"GOOGLE_REGION",
			"GCLOUD_REGION",
			"CLOUDSDK_COMPUTE_REGION",
		}, nil)

		if region != nil {
			data.Region = types.StringValue(region.(string))
		}
	}

	if data.Zone.IsNull() {
		zone := MultiEnvDefault([]string{
			"GOOGLE_ZONE",
			"GCLOUD_ZONE",
			"CLOUDSDK_COMPUTE_ZONE",
		}, nil)

		if zone != nil {
			data.Zone = types.StringValue(zone.(string))
		}
	}

	if len(data.Scopes.Elements()) == 0 {
		var d diag.Diagnostics
		data.Scopes, d = types.ListValueFrom(ctx, types.StringType, defaultClientScopes)
		diags.Append(d...)
		if diags.HasError() {
			return
		}
	}

	if !data.Batching.IsNull() {
		var pbConfigs []ProviderBatching
		d := data.Batching.ElementsAs(ctx, &pbConfigs, true)
		diags.Append(d...)
		if diags.HasError() {
			return
		}

		if pbConfigs[0].SendAfter.IsNull() {
			pbConfigs[0].SendAfter = types.StringValue("10s")
		}

		if pbConfigs[0].EnableBatching.IsNull() {
			pbConfigs[0].EnableBatching = types.BoolValue(true)
		}

		data.Batching, d = types.ListValueFrom(ctx, types.ObjectType{}, pbConfigs)
	}

	if data.UserProjectOverride.IsNull() && os.Getenv("USER_PROJECT_OVERRIDE") != "" {
		override, err := strconv.ParseBool(os.Getenv("USER_PROJECT_OVERRIDE"))
		if err != nil {
			diags.AddError(
				"error parsing environment variable `USER_PROJECT_OVERRIDE` into bool", err.Error())
		}
		data.UserProjectOverride = types.BoolValue(override)
	}

	if data.RequestReason.IsNull() && os.Getenv("CLOUDSDK_CORE_REQUEST_REASON") != "" {
		data.RequestReason = types.StringValue(os.Getenv("CLOUDSDK_CORE_REQUEST_REASON"))
	}

	if data.RequestTimeout.IsNull() {
		data.RequestTimeout = types.StringValue("120s")
	}

	// Generated Products
	if data.ComputeCustomEndpoint.IsNull() {
		customEndpoint := MultiEnvDefault([]string{
			"GOOGLE_COMPUTE_CUSTOM_ENDPOINT",
		}, DefaultBasePaths[ComputeBasePathKey])
		if customEndpoint != nil {
			data.ComputeCustomEndpoint = types.StringValue(customEndpoint.(string))
		}
	}

	// Handwritten Products / Versioned / Atypical Entries
	if data.CloudBillingCustomEndpoint.IsNull() {
		customEndpoint := MultiEnvDefault([]string{
			"GOOGLE_CLOUD_BILLING_CUSTOM_ENDPOINT",
		}, DefaultBasePaths["cloud_billing_custom_endpoint"])
		if customEndpoint != nil {
			data.CloudBillingCustomEndpoint = types.StringValue(customEndpoint.(string))
		}
	}

	if data.ComposerCustomEndpoint.IsNull() {
		customEndpoint := MultiEnvDefault([]string{
			"GOOGLE_COMPOSER_CUSTOM_ENDPOINT",
		}, DefaultBasePaths[ComposerBasePathKey])
		if customEndpoint != nil {
			data.ComposerCustomEndpoint = types.StringValue(customEndpoint.(string))
		}
	}

	if data.ContainerCustomEndpoint.IsNull() {
		customEndpoint := MultiEnvDefault([]string{
			"GOOGLE_CONTAINER_CUSTOM_ENDPOINT",
		}, DefaultBasePaths[ContainerBasePathKey])
		if customEndpoint != nil {
			data.ContainerCustomEndpoint = types.StringValue(customEndpoint.(string))
		}
	}

	if data.DataflowCustomEndpoint.IsNull() {
		customEndpoint := MultiEnvDefault([]string{
			"GOOGLE_DATAFLOW_CUSTOM_ENDPOINT",
		}, DefaultBasePaths[DataflowBasePathKey])
		if customEndpoint != nil {
			data.DataflowCustomEndpoint = types.StringValue(customEndpoint.(string))
		}
	}

	if data.IamCredentialsCustomEndpoint.IsNull() {
		customEndpoint := MultiEnvDefault([]string{
			"GOOGLE_IAM_CREDENTIALS_CUSTOM_ENDPOINT",
		}, DefaultBasePaths[IamCredentialsBasePathKey])
		if customEndpoint != nil {
			data.IamCredentialsCustomEndpoint = types.StringValue(customEndpoint.(string))
		}
	}

	if data.ResourceManagerV3CustomEndpoint.IsNull() {
		customEndpoint := MultiEnvDefault([]string{
			"GOOGLE_RESOURCE_MANAGER_V3_CUSTOM_ENDPOINT",
		}, DefaultBasePaths[ResourceManagerV3BasePathKey])
		if customEndpoint != nil {
			data.ResourceManagerV3CustomEndpoint = types.StringValue(customEndpoint.(string))
		}
	}

	if data.RuntimeConfigCustomEndpoint.IsNull() {
		customEndpoint := MultiEnvDefault([]string{
			"GOOGLE_RUNTIMECONFIG_CUSTOM_ENDPOINT",
		}, DefaultBasePaths[RuntimeConfigBasePathKey])
		if customEndpoint != nil {
			data.RuntimeConfigCustomEndpoint = types.StringValue(customEndpoint.(string))
		}
	}

	if data.IAMCustomEndpoint.IsNull() {
		customEndpoint := MultiEnvDefault([]string{
			"GOOGLE_IAM_CUSTOM_ENDPOINT",
		}, DefaultBasePaths[IAMBasePathKey])
		if customEndpoint != nil {
			data.IAMCustomEndpoint = types.StringValue(customEndpoint.(string))
		}
	}

	if data.ServiceNetworkingCustomEndpoint.IsNull() {
		customEndpoint := MultiEnvDefault([]string{
			"GOOGLE_SERVICE_NETWORKING_CUSTOM_ENDPOINT",
		}, DefaultBasePaths[ServiceNetworkingBasePathKey])
		if customEndpoint != nil {
			data.ServiceNetworkingCustomEndpoint = types.StringValue(customEndpoint.(string))
		}
	}

	if data.TagsLocationCustomEndpoint.IsNull() {
		customEndpoint := MultiEnvDefault([]string{
			"GOOGLE_TAGS_LOCATION_CUSTOM_ENDPOINT",
		}, DefaultBasePaths[TagsLocationBasePathKey])
		if customEndpoint != nil {
			data.TagsLocationCustomEndpoint = types.StringValue(customEndpoint.(string))
		}
	}

	// dcl
	if data.ContainerAwsCustomEndpoint.IsNull() {
		customEndpoint := MultiEnvDefault([]string{
			"GOOGLE_CONTAINERAWS_CUSTOM_ENDPOINT",
		}, DefaultBasePaths[ContainerAwsBasePathKey])
		if customEndpoint != nil {
			data.ContainerAwsCustomEndpoint = types.StringValue(customEndpoint.(string))
		}
	}

	if data.ContainerAzureCustomEndpoint.IsNull() {
		customEndpoint := MultiEnvDefault([]string{
			"GOOGLE_CONTAINERAZURE_CUSTOM_ENDPOINT",
		}, DefaultBasePaths[ContainerAzureBasePathKey])
		if customEndpoint != nil {
			data.ContainerAzureCustomEndpoint = types.StringValue(customEndpoint.(string))
		}
	}

	// DCL generated defaults
	if data.ApikeysCustomEndpoint.IsNull() {
		customEndpoint := MultiEnvDefault([]string{
			"GOOGLE_APIKEYS_CUSTOM_ENDPOINT",
		}, "")
		if customEndpoint != nil {
			data.ApikeysCustomEndpoint = types.StringValue(customEndpoint.(string))
		}
	}

	if data.AssuredWorkloadsCustomEndpoint.IsNull() {
		customEndpoint := MultiEnvDefault([]string{
			"GOOGLE_ASSURED_WORKLOADS_CUSTOM_ENDPOINT",
		}, "")
		if customEndpoint != nil {
			data.AssuredWorkloadsCustomEndpoint = types.StringValue(customEndpoint.(string))
		}
	}

	if data.CloudBuildWorkerPoolCustomEndpoint.IsNull() {
		customEndpoint := MultiEnvDefault([]string{
			"GOOGLE_CLOUD_BUILD_WORKER_POOL_CUSTOM_ENDPOINT",
		}, "")
		if customEndpoint != nil {
			data.CloudBuildWorkerPoolCustomEndpoint = types.StringValue(customEndpoint.(string))
		}
	}

	if data.CloudDeployCustomEndpoint.IsNull() {
		customEndpoint := MultiEnvDefault([]string{
			"GOOGLE_CLOUDDEPLOY_CUSTOM_ENDPOINT",
		}, "")
		if customEndpoint != nil {
			data.CloudDeployCustomEndpoint = types.StringValue(customEndpoint.(string))
		}
	}

	if data.CloudResourceManagerCustomEndpoint.IsNull() {
		customEndpoint := MultiEnvDefault([]string{
			"GOOGLE_CLOUD_RESOURCE_MANAGER_CUSTOM_ENDPOINT",
		}, "")
		if customEndpoint != nil {
			data.CloudResourceManagerCustomEndpoint = types.StringValue(customEndpoint.(string))
		}
	}

	if data.DataplexCustomEndpoint.IsNull() {
		customEndpoint := MultiEnvDefault([]string{
			"GOOGLE_DATAPLEX_CUSTOM_ENDPOINT",
		}, "")
		if customEndpoint != nil {
			data.DataplexCustomEndpoint = types.StringValue(customEndpoint.(string))
		}
	}

	if data.EventarcCustomEndpoint.IsNull() {
		customEndpoint := MultiEnvDefault([]string{
			"GOOGLE_EVENTARC_CUSTOM_ENDPOINT",
		}, "")
		if customEndpoint != nil {
			data.EventarcCustomEndpoint = types.StringValue(customEndpoint.(string))
		}
	}

	if data.FirebaserulesCustomEndpoint.IsNull() {
		customEndpoint := MultiEnvDefault([]string{
			"GOOGLE_FIREBASERULES_CUSTOM_ENDPOINT",
		}, "")
		if customEndpoint != nil {
			data.FirebaserulesCustomEndpoint = types.StringValue(customEndpoint.(string))
		}
	}

	if data.NetworkConnectivityCustomEndpoint.IsNull() {
		customEndpoint := MultiEnvDefault([]string{
			"GOOGLE_NETWORK_CONNECTIVITY_CUSTOM_ENDPOINT",
		}, "")
		if customEndpoint != nil {
			data.NetworkConnectivityCustomEndpoint = types.StringValue(customEndpoint.(string))
		}
	}

	if data.RecaptchaEnterpriseCustomEndpoint.IsNull() {
		customEndpoint := MultiEnvDefault([]string{
			"GOOGLE_RECAPTCHA_ENTERPRISE_CUSTOM_ENDPOINT",
		}, "")
		if customEndpoint != nil {
			data.RecaptchaEnterpriseCustomEndpoint = types.StringValue(customEndpoint.(string))
		}
	}
}

func (p *frameworkProvider) SetupClient(ctx context.Context, data ProviderModel, diags *diag.Diagnostics) {
	tokenSource := GetTokenSource(ctx, data, false, diags)
	if diags.HasError() {
		return
	}

	cleanCtx := context.WithValue(ctx, oauth2.HTTPClient, cleanhttp.DefaultClient())

	// 1. MTLS TRANSPORT/CLIENT - sets up proper auth headers
	client, _, err := transport.NewHTTPClient(cleanCtx, option.WithTokenSource(tokenSource))
	if err != nil {
		diags.AddError("error creating new http client", err.Error())
		return
	}

	// Userinfo is fetched before request logging is enabled to reduce additional noise.
	p.logGoogleIdentities(ctx, data, diags)
	if diags.HasError() {
		return
	}

	// 2. Logging Transport - ensure we log HTTP requests to GCP APIs.
	loggingTransport := logging.NewTransport("Google", client.Transport)

	// 3. Retry Transport - retries common temporary errors
	// Keep order for wrapping logging so we log each retried request as well.
	// This value should be used if needed to create shallow copies with additional retry predicates.
	// See ClientWithAdditionalRetries
	retryTransport := NewTransportWithDefaultRetries(loggingTransport)

	// 4. Header Transport - outer wrapper to inject additional headers we want to apply
	// before making requests
	headerTransport := newTransportWithHeaders(retryTransport)
	if !data.RequestReason.IsNull() {
		headerTransport.Set("X-Goog-Request-Reason", data.RequestReason.ValueString())
	}

	// Ensure $userProject is set for all HTTP requests using the client if specified by the provider config
	// See https://cloud.google.com/apis/docs/system-parameters
	if data.UserProjectOverride.ValueBool() && !data.BillingProject.IsNull() {
		headerTransport.Set("X-Goog-User-Project", data.BillingProject.ValueString())
	}

	// Set final transport value.
	client.Transport = headerTransport

	// This timeout is a timeout per HTTP request, not per logical operation.
	timeout, err := time.ParseDuration(data.RequestTimeout.ValueString())
	if err != nil {
		diags.AddError("error parsing request timeout", err.Error())
	}
	client.Timeout = timeout

	p.tokenSource = tokenSource
	p.client = client
}

func (p *frameworkProvider) SetupGrpcLogging() {
	logger := logrus.StandardLogger()

	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetFormatter(&Formatter{
		TimestampFormat: "2006/01/02 15:04:05",
		LogFormat:       "%time% [%lvl%] %msg% \n",
	})

	alwaysLoggingDeciderClient := func(ctx context.Context, fullMethodName string) bool { return true }
	grpc_logrus.ReplaceGrpcLogger(logrus.NewEntry(logger))

	p.gRPCLoggingOptions = append(
		p.gRPCLoggingOptions, option.WithGRPCDialOption(grpc.WithUnaryInterceptor(
			grpc_logrus.PayloadUnaryClientInterceptor(logrus.NewEntry(logger), alwaysLoggingDeciderClient))),
		option.WithGRPCDialOption(grpc.WithStreamInterceptor(
			grpc_logrus.PayloadStreamClientInterceptor(logrus.NewEntry(logger), alwaysLoggingDeciderClient))),
	)
}

func (p *frameworkProvider) logGoogleIdentities(ctx context.Context, data ProviderModel, diags *diag.Diagnostics) {
	// GetCurrentUserEmailFramework doesn't pass an error back from logGoogleIdentities, so we want
	// a separate diagnostics here
	var d diag.Diagnostics

	if data.ImpersonateServiceAccount.IsNull() {

		tokenSource := GetTokenSource(ctx, data, true, diags)
		if diags.HasError() {
			return
		}

		p.client = oauth2.NewClient(ctx, tokenSource) // p.client isn't initialised fully when this code is called.

		email := GetCurrentUserEmailFramework(p, p.userAgent, &d)
		if d.HasError() {
			tflog.Info(ctx, "error retrieving userinfo for your provider credentials. have you enabled the 'https://www.googleapis.com/auth/userinfo.email' scope?")
		}

		tflog.Info(ctx, fmt.Sprintf("Terraform is using this identity: %s", email))
		return
	}

	// Drop Impersonated ClientOption from OAuth2 TokenSource to infer original identity
	tokenSource := GetTokenSource(ctx, data, true, diags)
	if diags.HasError() {
		return
	}

	p.client = oauth2.NewClient(ctx, tokenSource) // p.client isn't initialised fully when this code is called.
	email := GetCurrentUserEmailFramework(p, p.userAgent, &d)
	if d.HasError() {
		tflog.Info(ctx, "error retrieving userinfo for your provider credentials. have you enabled the 'https://www.googleapis.com/auth/userinfo.email' scope?")
	}

	tflog.Info(ctx, fmt.Sprintf("Terraform is configured with service account impersonation, original identity: %s, impersonated identity: %s", email, data.ImpersonateServiceAccount.ValueString()))

	// Add the Impersonated ClientOption back in to the OAuth2 TokenSource
	tokenSource = GetTokenSource(ctx, data, false, diags)
	if diags.HasError() {
		return
	}

	p.client = oauth2.NewClient(ctx, tokenSource) // p.client isn't initialised fully when this code is called.

	return
}

// Configuration helpers

// GetTokenSource gets token source based on the Google Credentials configured.
// If initialCredentialsOnly is true, don't follow the impersonation settings and return the initial set of creds.
func GetTokenSource(ctx context.Context, data ProviderModel, initialCredentialsOnly bool, diags *diag.Diagnostics) oauth2.TokenSource {
	creds := GetCredentials(ctx, data, initialCredentialsOnly, diags)

	return creds.TokenSource
}

// GetCredentials gets credentials with a given scope (clientScopes).
// If initialCredentialsOnly is true, don't follow the impersonation
// settings and return the initial set of creds instead.
func GetCredentials(ctx context.Context, data ProviderModel, initialCredentialsOnly bool, diags *diag.Diagnostics) googleoauth.Credentials {
	var clientScopes []string
	var delegates []string

	d := data.Scopes.ElementsAs(ctx, &clientScopes, false)
	diags.Append(d...)
	if diags.HasError() {
		return googleoauth.Credentials{}
	}

	d = data.ImpersonateServiceAccountDelegates.ElementsAs(ctx, &delegates, false)
	diags.Append(d...)
	if diags.HasError() {
		return googleoauth.Credentials{}
	}

	if !data.AccessToken.IsNull() {
		contents, _, err := pathOrContents(data.AccessToken.ValueString())
		if err != nil {
			diags.AddError("error loading access token", err.Error())
			return googleoauth.Credentials{}
		}

		token := &oauth2.Token{AccessToken: contents}
		if !data.ImpersonateServiceAccount.IsNull() && !initialCredentialsOnly {
			opts := []option.ClientOption{option.WithTokenSource(oauth2.StaticTokenSource(token)), option.ImpersonateCredentials(data.ImpersonateServiceAccount.ValueString(), delegates...), option.WithScopes(clientScopes...)}
			creds, err := transport.Creds(context.TODO(), opts...)
			if err != nil {
				diags.AddError("error impersonating credentials", err.Error())
				return googleoauth.Credentials{}
			}
			return *creds
		}

		tflog.Info(ctx, "Authenticating using configured Google JSON 'access_token'...")
		tflog.Info(ctx, fmt.Sprintf("  -- Scopes: %s", clientScopes))
		return googleoauth.Credentials{
			TokenSource: staticTokenSource{oauth2.StaticTokenSource(token)},
		}
	}

	if !data.Credentials.IsNull() {
		contents, _, err := pathOrContents(data.Credentials.ValueString())
		if err != nil {
			diags.AddError(fmt.Sprintf("error loading credentials: %s", err), err.Error())
			return googleoauth.Credentials{}
		}

		if !data.ImpersonateServiceAccount.IsNull() && !initialCredentialsOnly {
			opts := []option.ClientOption{option.WithCredentialsJSON([]byte(contents)), option.ImpersonateCredentials(data.ImpersonateServiceAccount.ValueString(), delegates...), option.WithScopes(clientScopes...)}
			creds, err := transport.Creds(context.TODO(), opts...)
			if err != nil {
				diags.AddError("error impersonating credentials", err.Error())
				return googleoauth.Credentials{}
			}
			return *creds
		}

		creds, err := googleoauth.CredentialsFromJSON(ctx, []byte(contents), clientScopes...)
		if err != nil {
			diags.AddError("unable to parse credentials", err.Error())
			return googleoauth.Credentials{}
		}

		tflog.Info(ctx, "Authenticating using configured Google JSON 'credentials'...")
		tflog.Info(ctx, fmt.Sprintf("  -- Scopes: %s", clientScopes))
		return *creds
	}

	if !data.ImpersonateServiceAccount.IsNull() && !initialCredentialsOnly {
		opts := option.ImpersonateCredentials(data.ImpersonateServiceAccount.ValueString(), delegates...)
		creds, err := transport.Creds(context.TODO(), opts, option.WithScopes(clientScopes...))
		if err != nil {
			diags.AddError("error impersonating credentials", err.Error())
			return googleoauth.Credentials{}
		}

		return *creds
	}

	tflog.Info(ctx, "Authenticating using DefaultClient...")
	tflog.Info(ctx, fmt.Sprintf("  -- Scopes: %s", clientScopes))
	defaultTS, err := googleoauth.DefaultTokenSource(context.Background(), clientScopes...)
	if err != nil {
		diags.AddError(fmt.Sprintf("Attempted to load application default credentials since neither `credentials` nor `access_token` was set in the provider block.  "+
			"No credentials loaded. To use your gcloud credentials, run 'gcloud auth application-default login'"), err.Error())
		return googleoauth.Credentials{}
	}

	return googleoauth.Credentials{
		TokenSource: defaultTS,
	}
}

// GetBatchingConfig returns the batching config object given the
// provider configuration set for batching
func GetBatchingConfig(ctx context.Context, data types.List, diags *diag.Diagnostics) *batchingConfig {
	bc := &batchingConfig{
		SendAfter:      time.Second * DefaultBatchSendIntervalSec,
		EnableBatching: true,
	}

	if data.IsNull() {
		return bc
	}

	var pbConfigs []ProviderBatching
	d := data.ElementsAs(ctx, &pbConfigs, true)
	diags.Append(d...)
	if diags.HasError() {
		return bc
	}

	sendAfter, err := time.ParseDuration(pbConfigs[0].SendAfter.ValueString())
	if err != nil {
		diags.AddError("error parsing send after time duration", err.Error())
		return bc
	}

	bc.SendAfter = sendAfter

	if !pbConfigs[0].EnableBatching.IsNull() {
		bc.EnableBatching = pbConfigs[0].EnableBatching.ValueBool()
	}

	return bc
}
