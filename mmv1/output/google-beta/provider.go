package google

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-google-beta/version"

	googleoauth "golang.org/x/oauth2/google"
)

const TestEnvVar = "TF_ACC"

// Global MutexKV
var mutexKV = NewMutexKV()

// Provider returns a *schema.Provider.
func Provider() *schema.Provider {

	// The mtls service client gives the type of endpoint (mtls/regular)
	// at client creation. Since we use a shared client for requests we must
	// rewrite the endpoints to be mtls endpoints for the scenario where
	// mtls is enabled.
	if isMtls() {
		// if mtls is enabled switch all default endpoints to use the mtls endpoint
		for key, bp := range DefaultBasePaths {
			DefaultBasePaths[key] = getMtlsEndpoint(bp)
		}
	}

	provider := &schema.Provider{
		Schema: map[string]*schema.Schema{
			"credentials": {
				Type:          schema.TypeString,
				Optional:      true,
				ValidateFunc:  validateCredentials,
				ConflictsWith: []string{"access_token"},
			},

			"access_token": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"credentials"},
			},

			"impersonate_service_account": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"impersonate_service_account_delegates": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"project": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"billing_project": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"region": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"zone": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"scopes": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"batching": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"send_after": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validateNonNegativeDuration(),
						},
						"enable_batching": {
							Type:     schema.TypeBool,
							Optional: true,
						},
					},
				},
			},

			"user_project_override": {
				Type:     schema.TypeBool,
				Optional: true,
			},

			"request_timeout": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"request_reason": {
				Type:     schema.TypeString,
				Optional: true,
			},

			// Generated Products
			"compute_custom_endpoint": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateCustomEndpoint,
			},

			// Handwritten Products / Versioned / Atypical Entries
			CloudBillingCustomEndpointEntryKey:      CloudBillingCustomEndpointEntry,
			ComposerCustomEndpointEntryKey:          ComposerCustomEndpointEntry,
			ContainerCustomEndpointEntryKey:         ContainerCustomEndpointEntry,
			DataflowCustomEndpointEntryKey:          DataflowCustomEndpointEntry,
			IamCredentialsCustomEndpointEntryKey:    IamCredentialsCustomEndpointEntry,
			ResourceManagerV3CustomEndpointEntryKey: ResourceManagerV3CustomEndpointEntry,
			RuntimeConfigCustomEndpointEntryKey:     RuntimeConfigCustomEndpointEntry,
			IAMCustomEndpointEntryKey:               IAMCustomEndpointEntry,
			ServiceNetworkingCustomEndpointEntryKey: ServiceNetworkingCustomEndpointEntry,
			TagsLocationCustomEndpointEntryKey:      TagsLocationCustomEndpointEntry,

			// dcl
			ContainerAwsCustomEndpointEntryKey:   ContainerAwsCustomEndpointEntry,
			ContainerAzureCustomEndpointEntryKey: ContainerAzureCustomEndpointEntry,
		},

		ProviderMetaSchema: map[string]*schema.Schema{
			"module_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},

		DataSourcesMap: map[string]*schema.Resource{
			// ####### START datasources ###########
			"google_access_approval_folder_service_account":       DataSourceAccessApprovalFolderServiceAccount(),
			"google_access_approval_organization_service_account": DataSourceAccessApprovalOrganizationServiceAccount(),
			"google_access_approval_project_service_account":      DataSourceAccessApprovalProjectServiceAccount(),
			"google_active_folder":                                DataSourceGoogleActiveFolder(),
			"google_artifact_registry_repository":                 DataSourceArtifactRegistryRepository(),
			"google_app_engine_default_service_account":           DataSourceGoogleAppEngineDefaultServiceAccount(),
			"google_beyondcorp_app_connection":                    DataSourceGoogleBeyondcorpAppConnection(),
			"google_beyondcorp_app_connector":                     DataSourceGoogleBeyondcorpAppConnector(),
			"google_beyondcorp_app_gateway":                       DataSourceGoogleBeyondcorpAppGateway(),
			"google_billing_account":                              DataSourceGoogleBillingAccount(),
			"google_bigquery_default_service_account":             DataSourceGoogleBigqueryDefaultServiceAccount(),
			"google_cloudbuild_trigger":                           DataSourceGoogleCloudBuildTrigger(),
			"google_cloudfunctions_function":                      DataSourceGoogleCloudFunctionsFunction(),
			"google_cloudfunctions2_function":                     DataSourceGoogleCloudFunctions2Function(),
			"google_cloud_asset_resources_search_all":             DataSourceGoogleCloudAssetResourcesSearchAll(),
			"google_cloud_identity_groups":                        DataSourceGoogleCloudIdentityGroups(),
			"google_cloud_identity_group_memberships":             DataSourceGoogleCloudIdentityGroupMemberships(),
			"google_cloud_run_locations":                          DataSourceGoogleCloudRunLocations(),
			"google_cloud_run_service":                            DataSourceGoogleCloudRunService(),
			"google_composer_environment":                         DataSourceGoogleComposerEnvironment(),
			"google_composer_image_versions":                      DataSourceGoogleComposerImageVersions(),
			"google_compute_address":                              DataSourceGoogleComputeAddress(),
			"google_compute_addresses":                            DataSourceGoogleComputeAddresses(),
			"google_compute_backend_service":                      DataSourceGoogleComputeBackendService(),
			"google_compute_backend_bucket":                       DataSourceGoogleComputeBackendBucket(),
			"google_compute_default_service_account":              DataSourceGoogleComputeDefaultServiceAccount(),
			"google_compute_disk":                                 DataSourceGoogleComputeDisk(),
			"google_compute_forwarding_rule":                      DataSourceGoogleComputeForwardingRule(),
			"google_compute_global_address":                       DataSourceGoogleComputeGlobalAddress(),
			"google_compute_global_forwarding_rule":               DataSourceGoogleComputeGlobalForwardingRule(),
			"google_compute_ha_vpn_gateway":                       DataSourceGoogleComputeHaVpnGateway(),
			"google_compute_health_check":                         DataSourceGoogleComputeHealthCheck(),
			"google_compute_image":                                DataSourceGoogleComputeImage(),
			"google_compute_instance":                             DataSourceGoogleComputeInstance(),
			"google_compute_instance_group":                       DataSourceGoogleComputeInstanceGroup(),
			"google_compute_instance_group_manager":               DataSourceGoogleComputeInstanceGroupManager(),
			"google_compute_instance_serial_port":                 DataSourceGoogleComputeInstanceSerialPort(),
			"google_compute_instance_template":                    DataSourceGoogleComputeInstanceTemplate(),
			"google_compute_lb_ip_ranges":                         DataSourceGoogleComputeLbIpRanges(),
			"google_compute_network":                              DataSourceGoogleComputeNetwork(),
			"google_compute_network_endpoint_group":               DataSourceGoogleComputeNetworkEndpointGroup(),
			"google_compute_network_peering":                      DataSourceComputeNetworkPeering(),
			"google_compute_node_types":                           DataSourceGoogleComputeNodeTypes(),
			"google_compute_regions":                              DataSourceGoogleComputeRegions(),
			"google_compute_region_network_endpoint_group":        DataSourceGoogleComputeRegionNetworkEndpointGroup(),
			"google_compute_region_instance_group":                DataSourceGoogleComputeRegionInstanceGroup(),
			"google_compute_region_instance_template":             DataSourceGoogleComputeRegionInstanceTemplate(),
			"google_compute_region_ssl_certificate":               DataSourceGoogleRegionComputeSslCertificate(),
			"google_compute_resource_policy":                      DataSourceGoogleComputeResourcePolicy(),
			"google_compute_router":                               DataSourceGoogleComputeRouter(),
			"google_compute_router_nat":                           DataSourceGoogleComputeRouterNat(),
			"google_compute_router_status":                        DataSourceGoogleComputeRouterStatus(),
			"google_compute_snapshot":                             DataSourceGoogleComputeSnapshot(),
			"google_compute_ssl_certificate":                      DataSourceGoogleComputeSslCertificate(),
			"google_compute_ssl_policy":                           DataSourceGoogleComputeSslPolicy(),
			"google_compute_subnetwork":                           DataSourceGoogleComputeSubnetwork(),
			"google_compute_vpn_gateway":                          DataSourceGoogleComputeVpnGateway(),
			"google_compute_zones":                                DataSourceGoogleComputeZones(),
			"google_container_azure_versions":                     DataSourceGoogleContainerAzureVersions(),
			"google_container_aws_versions":                       DataSourceGoogleContainerAwsVersions(),
			"google_container_attached_versions":                  DataSourceGoogleContainerAttachedVersions(),
			"google_container_attached_install_manifest":          DataSourceGoogleContainerAttachedInstallManifest(),
			"google_container_cluster":                            DataSourceGoogleContainerCluster(),
			"google_container_engine_versions":                    DataSourceGoogleContainerEngineVersions(),
			"google_container_registry_image":                     DataSourceGoogleContainerImage(),
			"google_container_registry_repository":                DataSourceGoogleContainerRepo(),
			"google_dataproc_metastore_service":                   DataSourceDataprocMetastoreService(),
			"google_game_services_game_server_deployment_rollout": DataSourceGameServicesGameServerDeploymentRollout(),
			"google_iam_policy":                                   DataSourceGoogleIamPolicy(),
			"google_iam_role":                                     DataSourceGoogleIamRole(),
			"google_iam_testable_permissions":                     DataSourceGoogleIamTestablePermissions(),
			"google_iam_workload_identity_pool":                   DataSourceIAMBetaWorkloadIdentityPool(),
			"google_iam_workload_identity_pool_provider":          DataSourceIAMBetaWorkloadIdentityPoolProvider(),
			"google_iap_client":                                   DataSourceGoogleIapClient(),
			"google_kms_crypto_key":                               DataSourceGoogleKmsCryptoKey(),
			"google_kms_crypto_key_version":                       DataSourceGoogleKmsCryptoKeyVersion(),
			"google_kms_key_ring":                                 DataSourceGoogleKmsKeyRing(),
			"google_kms_secret":                                   DataSourceGoogleKmsSecret(),
			"google_kms_secret_ciphertext":                        DataSourceGoogleKmsSecretCiphertext(),
			"google_kms_secret_asymmetric":                        DataSourceGoogleKmsSecretAsymmetric(),
			"google_firebase_android_app":                         DataSourceGoogleFirebaseAndroidApp(),
			"google_firebase_apple_app":                           DataSourceGoogleFirebaseAppleApp(),
			"google_firebase_hosting_channel":                     DataSourceGoogleFirebaseHostingChannel(),
			"google_firebase_web_app":                             DataSourceGoogleFirebaseWebApp(),
			"google_folder":                                       DataSourceGoogleFolder(),
			"google_folders":                                      DataSourceGoogleFolders(),
			"google_folder_organization_policy":                   DataSourceGoogleFolderOrganizationPolicy(),
			"google_logging_project_cmek_settings":                DataSourceGoogleLoggingProjectCmekSettings(),
			"google_logging_sink":                                 DataSourceGoogleLoggingSink(),
			"google_monitoring_notification_channel":              DataSourceMonitoringNotificationChannel(),
			"google_monitoring_cluster_istio_service":             DataSourceMonitoringServiceClusterIstio(),
			"google_monitoring_istio_canonical_service":           DataSourceMonitoringIstioCanonicalService(),
			"google_monitoring_mesh_istio_service":                DataSourceMonitoringServiceMeshIstio(),
			"google_monitoring_app_engine_service":                DataSourceMonitoringServiceAppEngine(),
			"google_monitoring_uptime_check_ips":                  DataSourceGoogleMonitoringUptimeCheckIps(),
			"google_netblock_ip_ranges":                           DataSourceGoogleNetblockIpRanges(),
			"google_organization":                                 DataSourceGoogleOrganization(),
			"google_privateca_certificate_authority":              DataSourcePrivatecaCertificateAuthority(),
			"google_project":                                      DataSourceGoogleProject(),
			"google_projects":                                     DataSourceGoogleProjects(),
			"google_project_organization_policy":                  DataSourceGoogleProjectOrganizationPolicy(),
			"google_project_service":                              DataSourceGoogleProjectService(),
			"google_pubsub_subscription":                          DataSourceGooglePubsubSubscription(),
			"google_pubsub_topic":                                 DataSourceGooglePubsubTopic(),
			"google_runtimeconfig_config":                         DataSourceGoogleRuntimeconfigConfig(),
			"google_runtimeconfig_variable":                       DataSourceGoogleRuntimeconfigVariable(),
			"google_secret_manager_secret":                        DataSourceSecretManagerSecret(),
			"google_secret_manager_secret_version":                DataSourceSecretManagerSecretVersion(),
			"google_secret_manager_secret_version_access":         DataSourceSecretManagerSecretVersionAccess(),
			"google_service_account":                              DataSourceGoogleServiceAccount(),
			"google_service_account_access_token":                 DataSourceGoogleServiceAccountAccessToken(),
			"google_service_account_id_token":                     DataSourceGoogleServiceAccountIdToken(),
			"google_service_account_jwt":                          DataSourceGoogleServiceAccountJwt(),
			"google_service_account_key":                          DataSourceGoogleServiceAccountKey(),
			"google_sourcerepo_repository":                        DataSourceGoogleSourceRepoRepository(),
			"google_spanner_instance":                             DataSourceSpannerInstance(),
			"google_sql_ca_certs":                                 DataSourceGoogleSQLCaCerts(),
			"google_sql_backup_run":                               DataSourceSqlBackupRun(),
			"google_sql_databases":                                DataSourceSqlDatabases(),
			"google_sql_database":                                 DataSourceSqlDatabase(),
			"google_sql_database_instance":                        DataSourceSqlDatabaseInstance(),
			"google_sql_database_instances":                       DataSourceSqlDatabaseInstances(),
			"google_service_networking_peered_dns_domain":         DataSourceGoogleServiceNetworkingPeeredDNSDomain(),
			"google_storage_bucket":                               DataSourceGoogleStorageBucket(),
			"google_storage_bucket_object":                        DataSourceGoogleStorageBucketObject(),
			"google_storage_bucket_object_content":                DataSourceGoogleStorageBucketObjectContent(),
			"google_storage_object_signed_url":                    DataSourceGoogleSignedUrl(),
			"google_storage_project_service_account":              DataSourceGoogleStorageProjectServiceAccount(),
			"google_storage_transfer_project_service_account":     DataSourceGoogleStorageTransferProjectServiceAccount(),
			"google_tags_tag_key":                                 DataSourceGoogleTagsTagKey(),
			"google_tags_tag_value":                               DataSourceGoogleTagsTagValue(),
			"google_tpu_tensorflow_versions":                      DataSourceTpuTensorflowVersions(),
			"google_vpc_access_connector":                         DataSourceVPCAccessConnector(),
			"google_redis_instance":                               DataSourceGoogleRedisInstance(),
			// ####### END datasources ###########
		},
		ResourcesMap: ResourceMap(),
	}

	provider.ConfigureContextFunc = func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		return providerConfigure(ctx, d, provider)
	}

	ConfigureDCLProvider(provider)

	return provider
}

// Generated resources: 68
// Generated IAM resources: 30
// Total generated resources: 98
func ResourceMap() map[string]*schema.Resource {
	resourceMap, _ := ResourceMapWithErrors()
	return resourceMap
}

func ResourceMapWithErrors() (map[string]*schema.Resource, error) {
	return mergeResourceMaps(
		map[string]*schema.Resource{
			"google_compute_address":                                  ResourceComputeAddress(),
			"google_compute_autoscaler":                               ResourceComputeAutoscaler(),
			"google_compute_backend_bucket":                           ResourceComputeBackendBucket(),
			"google_compute_backend_bucket_iam_binding":               ResourceIamBinding(ComputeBackendBucketIamSchema, ComputeBackendBucketIamUpdaterProducer, ComputeBackendBucketIdParseFunc),
			"google_compute_backend_bucket_iam_member":                ResourceIamMember(ComputeBackendBucketIamSchema, ComputeBackendBucketIamUpdaterProducer, ComputeBackendBucketIdParseFunc),
			"google_compute_backend_bucket_iam_policy":                ResourceIamPolicy(ComputeBackendBucketIamSchema, ComputeBackendBucketIamUpdaterProducer, ComputeBackendBucketIdParseFunc),
			"google_compute_backend_bucket_signed_url_key":            ResourceComputeBackendBucketSignedUrlKey(),
			"google_compute_backend_service":                          ResourceComputeBackendService(),
			"google_compute_backend_service_iam_binding":              ResourceIamBinding(ComputeBackendServiceIamSchema, ComputeBackendServiceIamUpdaterProducer, ComputeBackendServiceIdParseFunc),
			"google_compute_backend_service_iam_member":               ResourceIamMember(ComputeBackendServiceIamSchema, ComputeBackendServiceIamUpdaterProducer, ComputeBackendServiceIdParseFunc),
			"google_compute_backend_service_iam_policy":               ResourceIamPolicy(ComputeBackendServiceIamSchema, ComputeBackendServiceIamUpdaterProducer, ComputeBackendServiceIdParseFunc),
			"google_compute_backend_service_signed_url_key":           ResourceComputeBackendServiceSignedUrlKey(),
			"google_compute_disk":                                     ResourceComputeDisk(),
			"google_compute_disk_iam_binding":                         ResourceIamBinding(ComputeDiskIamSchema, ComputeDiskIamUpdaterProducer, ComputeDiskIdParseFunc),
			"google_compute_disk_iam_member":                          ResourceIamMember(ComputeDiskIamSchema, ComputeDiskIamUpdaterProducer, ComputeDiskIdParseFunc),
			"google_compute_disk_iam_policy":                          ResourceIamPolicy(ComputeDiskIamSchema, ComputeDiskIamUpdaterProducer, ComputeDiskIdParseFunc),
			"google_compute_disk_resource_policy_attachment":          ResourceComputeDiskResourcePolicyAttachment(),
			"google_compute_external_vpn_gateway":                     ResourceComputeExternalVpnGateway(),
			"google_compute_firewall":                                 ResourceComputeFirewall(),
			"google_compute_forwarding_rule":                          ResourceComputeForwardingRule(),
			"google_compute_global_address":                           ResourceComputeGlobalAddress(),
			"google_compute_global_forwarding_rule":                   ResourceComputeGlobalForwardingRule(),
			"google_compute_global_network_endpoint":                  ResourceComputeGlobalNetworkEndpoint(),
			"google_compute_global_network_endpoint_group":            ResourceComputeGlobalNetworkEndpointGroup(),
			"google_compute_ha_vpn_gateway":                           ResourceComputeHaVpnGateway(),
			"google_compute_health_check":                             ResourceComputeHealthCheck(),
			"google_compute_http_health_check":                        ResourceComputeHttpHealthCheck(),
			"google_compute_https_health_check":                       ResourceComputeHttpsHealthCheck(),
			"google_compute_image":                                    ResourceComputeImage(),
			"google_compute_image_iam_binding":                        ResourceIamBinding(ComputeImageIamSchema, ComputeImageIamUpdaterProducer, ComputeImageIdParseFunc),
			"google_compute_image_iam_member":                         ResourceIamMember(ComputeImageIamSchema, ComputeImageIamUpdaterProducer, ComputeImageIdParseFunc),
			"google_compute_image_iam_policy":                         ResourceIamPolicy(ComputeImageIamSchema, ComputeImageIamUpdaterProducer, ComputeImageIdParseFunc),
			"google_compute_instance_iam_binding":                     ResourceIamBinding(ComputeInstanceIamSchema, ComputeInstanceIamUpdaterProducer, ComputeInstanceIdParseFunc),
			"google_compute_instance_iam_member":                      ResourceIamMember(ComputeInstanceIamSchema, ComputeInstanceIamUpdaterProducer, ComputeInstanceIdParseFunc),
			"google_compute_instance_iam_policy":                      ResourceIamPolicy(ComputeInstanceIamSchema, ComputeInstanceIamUpdaterProducer, ComputeInstanceIdParseFunc),
			"google_compute_instance_group_named_port":                ResourceComputeInstanceGroupNamedPort(),
			"google_compute_interconnect_attachment":                  ResourceComputeInterconnectAttachment(),
			"google_compute_machine_image":                            ResourceComputeMachineImage(),
			"google_compute_machine_image_iam_binding":                ResourceIamBinding(ComputeMachineImageIamSchema, ComputeMachineImageIamUpdaterProducer, ComputeMachineImageIdParseFunc),
			"google_compute_machine_image_iam_member":                 ResourceIamMember(ComputeMachineImageIamSchema, ComputeMachineImageIamUpdaterProducer, ComputeMachineImageIdParseFunc),
			"google_compute_machine_image_iam_policy":                 ResourceIamPolicy(ComputeMachineImageIamSchema, ComputeMachineImageIamUpdaterProducer, ComputeMachineImageIdParseFunc),
			"google_compute_managed_ssl_certificate":                  ResourceComputeManagedSslCertificate(),
			"google_compute_network":                                  ResourceComputeNetwork(),
			"google_compute_network_endpoint":                         ResourceComputeNetworkEndpoint(),
			"google_compute_network_endpoint_group":                   ResourceComputeNetworkEndpointGroup(),
			"google_compute_network_peering_routes_config":            ResourceComputeNetworkPeeringRoutesConfig(),
			"google_compute_node_group":                               ResourceComputeNodeGroup(),
			"google_compute_node_template":                            ResourceComputeNodeTemplate(),
			"google_compute_organization_security_policy":             ResourceComputeOrganizationSecurityPolicy(),
			"google_compute_organization_security_policy_association": ResourceComputeOrganizationSecurityPolicyAssociation(),
			"google_compute_organization_security_policy_rule":        ResourceComputeOrganizationSecurityPolicyRule(),
			"google_compute_packet_mirroring":                         ResourceComputePacketMirroring(),
			"google_compute_per_instance_config":                      ResourceComputePerInstanceConfig(),
			"google_compute_region_autoscaler":                        ResourceComputeRegionAutoscaler(),
			"google_compute_region_backend_service":                   ResourceComputeRegionBackendService(),
			"google_compute_region_backend_service_iam_binding":       ResourceIamBinding(ComputeRegionBackendServiceIamSchema, ComputeRegionBackendServiceIamUpdaterProducer, ComputeRegionBackendServiceIdParseFunc),
			"google_compute_region_backend_service_iam_member":        ResourceIamMember(ComputeRegionBackendServiceIamSchema, ComputeRegionBackendServiceIamUpdaterProducer, ComputeRegionBackendServiceIdParseFunc),
			"google_compute_region_backend_service_iam_policy":        ResourceIamPolicy(ComputeRegionBackendServiceIamSchema, ComputeRegionBackendServiceIamUpdaterProducer, ComputeRegionBackendServiceIdParseFunc),
			"google_compute_region_disk":                              ResourceComputeRegionDisk(),
			"google_compute_region_disk_iam_binding":                  ResourceIamBinding(ComputeRegionDiskIamSchema, ComputeRegionDiskIamUpdaterProducer, ComputeRegionDiskIdParseFunc),
			"google_compute_region_disk_iam_member":                   ResourceIamMember(ComputeRegionDiskIamSchema, ComputeRegionDiskIamUpdaterProducer, ComputeRegionDiskIdParseFunc),
			"google_compute_region_disk_iam_policy":                   ResourceIamPolicy(ComputeRegionDiskIamSchema, ComputeRegionDiskIamUpdaterProducer, ComputeRegionDiskIdParseFunc),
			"google_compute_region_disk_resource_policy_attachment":   ResourceComputeRegionDiskResourcePolicyAttachment(),
			"google_compute_region_health_check":                      ResourceComputeRegionHealthCheck(),
			"google_compute_region_network_endpoint_group":            ResourceComputeRegionNetworkEndpointGroup(),
			"google_compute_region_per_instance_config":               ResourceComputeRegionPerInstanceConfig(),
			"google_compute_region_ssl_certificate":                   ResourceComputeRegionSslCertificate(),
			"google_compute_region_ssl_policy":                        ResourceComputeRegionSslPolicy(),
			"google_compute_region_target_http_proxy":                 ResourceComputeRegionTargetHttpProxy(),
			"google_compute_region_target_https_proxy":                ResourceComputeRegionTargetHttpsProxy(),
			"google_compute_region_target_tcp_proxy":                  ResourceComputeRegionTargetTcpProxy(),
			"google_compute_region_url_map":                           ResourceComputeRegionUrlMap(),
			"google_compute_reservation":                              ResourceComputeReservation(),
			"google_compute_resource_policy":                          ResourceComputeResourcePolicy(),
			"google_compute_route":                                    ResourceComputeRoute(),
			"google_compute_router":                                   ResourceComputeRouter(),
			"google_compute_router_peer":                              ResourceComputeRouterBgpPeer(),
			"google_compute_router_nat":                               ResourceComputeRouterNat(),
			"google_compute_service_attachment":                       ResourceComputeServiceAttachment(),
			"google_compute_snapshot":                                 ResourceComputeSnapshot(),
			"google_compute_snapshot_iam_binding":                     ResourceIamBinding(ComputeSnapshotIamSchema, ComputeSnapshotIamUpdaterProducer, ComputeSnapshotIdParseFunc),
			"google_compute_snapshot_iam_member":                      ResourceIamMember(ComputeSnapshotIamSchema, ComputeSnapshotIamUpdaterProducer, ComputeSnapshotIdParseFunc),
			"google_compute_snapshot_iam_policy":                      ResourceIamPolicy(ComputeSnapshotIamSchema, ComputeSnapshotIamUpdaterProducer, ComputeSnapshotIdParseFunc),
			"google_compute_ssl_certificate":                          ResourceComputeSslCertificate(),
			"google_compute_ssl_policy":                               ResourceComputeSslPolicy(),
			"google_compute_subnetwork":                               ResourceComputeSubnetwork(),
			"google_compute_subnetwork_iam_binding":                   ResourceIamBinding(ComputeSubnetworkIamSchema, ComputeSubnetworkIamUpdaterProducer, ComputeSubnetworkIdParseFunc),
			"google_compute_subnetwork_iam_member":                    ResourceIamMember(ComputeSubnetworkIamSchema, ComputeSubnetworkIamUpdaterProducer, ComputeSubnetworkIdParseFunc),
			"google_compute_subnetwork_iam_policy":                    ResourceIamPolicy(ComputeSubnetworkIamSchema, ComputeSubnetworkIamUpdaterProducer, ComputeSubnetworkIdParseFunc),
			"google_compute_target_grpc_proxy":                        ResourceComputeTargetGrpcProxy(),
			"google_compute_target_http_proxy":                        ResourceComputeTargetHttpProxy(),
			"google_compute_target_https_proxy":                       ResourceComputeTargetHttpsProxy(),
			"google_compute_target_instance":                          ResourceComputeTargetInstance(),
			"google_compute_target_ssl_proxy":                         ResourceComputeTargetSslProxy(),
			"google_compute_target_tcp_proxy":                         ResourceComputeTargetTcpProxy(),
			"google_compute_url_map":                                  ResourceComputeUrlMap(),
			"google_compute_vpn_gateway":                              ResourceComputeVpnGateway(),
			"google_compute_vpn_tunnel":                               ResourceComputeVpnTunnel(),
		},
		map[string]*schema.Resource{
			// ####### START handwritten resources ###########
			"google_app_engine_application":                 ResourceAppEngineApplication(),
			"google_apigee_sharedflow":                      ResourceApigeeSharedFlow(),
			"google_apigee_sharedflow_deployment":           ResourceApigeeSharedFlowDeployment(),
			"google_apigee_flowhook":                        ResourceApigeeFlowhook(),
			"google_apigee_env_keystore_alias_pkcs12":       ResourceApigeeEnvKeystoreAliasPkcs12(),
			"google_apigee_keystores_aliases_key_cert_file": ResourceApigeeKeystoresAliasesKeyCertFile(),
			"google_bigquery_table":                         ResourceBigQueryTable(),
			"google_bigtable_gc_policy":                     ResourceBigtableGCPolicy(),
			"google_bigtable_instance":                      ResourceBigtableInstance(),
			"google_bigtable_table":                         ResourceBigtableTable(),
			"google_billing_subaccount":                     ResourceBillingSubaccount(),
			"google_cloudfunctions_function":                ResourceCloudFunctionsFunction(),
			"google_composer_environment":                   ResourceComposerEnvironment(),
			"google_compute_attached_disk":                  ResourceComputeAttachedDisk(),
			"google_compute_instance":                       ResourceComputeInstance(),
			"google_compute_instance_from_machine_image":    ResourceComputeInstanceFromMachineImage(),
			"google_compute_instance_from_template":         ResourceComputeInstanceFromTemplate(),
			"google_compute_instance_group":                 ResourceComputeInstanceGroup(),
			"google_compute_instance_group_manager":         ResourceComputeInstanceGroupManager(),
			"google_compute_instance_template":              ResourceComputeInstanceTemplate(),
			"google_compute_network_peering":                ResourceComputeNetworkPeering(),
			"google_compute_project_default_network_tier":   ResourceComputeProjectDefaultNetworkTier(),
			"google_compute_project_metadata":               ResourceComputeProjectMetadata(),
			"google_compute_project_metadata_item":          ResourceComputeProjectMetadataItem(),
			"google_compute_region_instance_group_manager":  ResourceComputeRegionInstanceGroupManager(),
			"google_compute_region_instance_template":       ResourceComputeRegionInstanceTemplate(),
			"google_compute_router_interface":               ResourceComputeRouterInterface(),
			"google_compute_security_policy":                ResourceComputeSecurityPolicy(),
			"google_compute_shared_vpc_host_project":        ResourceComputeSharedVpcHostProject(),
			"google_compute_shared_vpc_service_project":     ResourceComputeSharedVpcServiceProject(),
			"google_compute_target_pool":                    ResourceComputeTargetPool(),
			"google_container_cluster":                      ResourceContainerCluster(),
			"google_container_node_pool":                    ResourceContainerNodePool(),
			"google_container_registry":                     ResourceContainerRegistry(),
			"google_dataflow_job":                           ResourceDataflowJob(),
			"google_dataflow_flex_template_job":             ResourceDataflowFlexTemplateJob(),
			"google_dataproc_cluster":                       ResourceDataprocCluster(),
			"google_dataproc_job":                           ResourceDataprocJob(),
			"google_dialogflow_cx_version":                  ResourceDialogflowCXVersion(),
			"google_dialogflow_cx_environment":              ResourceDialogflowCXEnvironment(),
			"google_dns_record_set":                         ResourceDnsRecordSet(),
			"google_endpoints_service":                      ResourceEndpointsService(),
			"google_folder":                                 ResourceGoogleFolder(),
			"google_folder_organization_policy":             ResourceGoogleFolderOrganizationPolicy(),
			"google_logging_billing_account_sink":           ResourceLoggingBillingAccountSink(),
			"google_logging_billing_account_exclusion":      ResourceLoggingExclusion(BillingAccountLoggingExclusionSchema, NewBillingAccountLoggingExclusionUpdater, BillingAccountLoggingExclusionIdParseFunc),
			"google_logging_billing_account_bucket_config":  ResourceLoggingBillingAccountBucketConfig(),
			"google_logging_organization_sink":              ResourceLoggingOrganizationSink(),
			"google_logging_organization_exclusion":         ResourceLoggingExclusion(OrganizationLoggingExclusionSchema, NewOrganizationLoggingExclusionUpdater, OrganizationLoggingExclusionIdParseFunc),
			"google_logging_organization_bucket_config":     ResourceLoggingOrganizationBucketConfig(),
			"google_logging_folder_sink":                    ResourceLoggingFolderSink(),
			"google_logging_folder_exclusion":               ResourceLoggingExclusion(FolderLoggingExclusionSchema, NewFolderLoggingExclusionUpdater, FolderLoggingExclusionIdParseFunc),
			"google_logging_folder_bucket_config":           ResourceLoggingFolderBucketConfig(),
			"google_logging_project_sink":                   ResourceLoggingProjectSink(),
			"google_logging_project_exclusion":              ResourceLoggingExclusion(ProjectLoggingExclusionSchema, NewProjectLoggingExclusionUpdater, ProjectLoggingExclusionIdParseFunc),
			"google_logging_project_bucket_config":          ResourceLoggingProjectBucketConfig(),
			"google_monitoring_dashboard":                   ResourceMonitoringDashboard(),
			"google_project_service_identity":               ResourceProjectServiceIdentity(),
			"google_service_networking_connection":          ResourceServiceNetworkingConnection(),
			"google_sql_database_instance":                  ResourceSqlDatabaseInstance(),
			"google_sql_ssl_cert":                           ResourceSqlSslCert(),
			"google_sql_user":                               ResourceSqlUser(),
			"google_organization_iam_custom_role":           ResourceGoogleOrganizationIamCustomRole(),
			"google_organization_policy":                    ResourceGoogleOrganizationPolicy(),
			"google_project":                                ResourceGoogleProject(),
			"google_project_default_service_accounts":       ResourceGoogleProjectDefaultServiceAccounts(),
			"google_project_service":                        ResourceGoogleProjectService(),
			"google_project_iam_custom_role":                ResourceGoogleProjectIamCustomRole(),
			"google_project_organization_policy":            ResourceGoogleProjectOrganizationPolicy(),
			"google_project_usage_export_bucket":            ResourceProjectUsageBucket(),
			"google_runtimeconfig_config":                   ResourceRuntimeconfigConfig(),
			"google_runtimeconfig_variable":                 ResourceRuntimeconfigVariable(),
			"google_service_account":                        ResourceGoogleServiceAccount(),
			"google_service_account_key":                    ResourceGoogleServiceAccountKey(),
			"google_service_networking_peered_dns_domain":   ResourceGoogleServiceNetworkingPeeredDNSDomain(),
			"google_storage_bucket":                         ResourceStorageBucket(),
			"google_storage_bucket_acl":                     ResourceStorageBucketAcl(),
			"google_storage_bucket_object":                  ResourceStorageBucketObject(),
			"google_storage_object_acl":                     ResourceStorageObjectAcl(),
			"google_storage_default_object_acl":             ResourceStorageDefaultObjectAcl(),
			"google_storage_notification":                   ResourceStorageNotification(),
			"google_storage_transfer_job":                   ResourceStorageTransferJob(),
			"google_tags_location_tag_binding":              ResourceTagsLocationTagBinding(),
			// ####### END handwritten resources ###########
		},
		map[string]*schema.Resource{
			// ####### START non-generated IAM resources ###########
			"google_bigtable_instance_iam_binding":       ResourceIamBinding(IamBigtableInstanceSchema, NewBigtableInstanceUpdater, BigtableInstanceIdParseFunc),
			"google_bigtable_instance_iam_member":        ResourceIamMember(IamBigtableInstanceSchema, NewBigtableInstanceUpdater, BigtableInstanceIdParseFunc),
			"google_bigtable_instance_iam_policy":        ResourceIamPolicy(IamBigtableInstanceSchema, NewBigtableInstanceUpdater, BigtableInstanceIdParseFunc),
			"google_bigtable_table_iam_binding":          ResourceIamBinding(IamBigtableTableSchema, NewBigtableTableUpdater, BigtableTableIdParseFunc),
			"google_bigtable_table_iam_member":           ResourceIamMember(IamBigtableTableSchema, NewBigtableTableUpdater, BigtableTableIdParseFunc),
			"google_bigtable_table_iam_policy":           ResourceIamPolicy(IamBigtableTableSchema, NewBigtableTableUpdater, BigtableTableIdParseFunc),
			"google_bigquery_dataset_iam_binding":        ResourceIamBinding(IamBigqueryDatasetSchema, NewBigqueryDatasetIamUpdater, BigqueryDatasetIdParseFunc),
			"google_bigquery_dataset_iam_member":         ResourceIamMember(IamBigqueryDatasetSchema, NewBigqueryDatasetIamUpdater, BigqueryDatasetIdParseFunc),
			"google_bigquery_dataset_iam_policy":         ResourceIamPolicy(IamBigqueryDatasetSchema, NewBigqueryDatasetIamUpdater, BigqueryDatasetIdParseFunc),
			"google_billing_account_iam_binding":         ResourceIamBinding(IamBillingAccountSchema, NewBillingAccountIamUpdater, BillingAccountIdParseFunc),
			"google_billing_account_iam_member":          ResourceIamMember(IamBillingAccountSchema, NewBillingAccountIamUpdater, BillingAccountIdParseFunc),
			"google_billing_account_iam_policy":          ResourceIamPolicy(IamBillingAccountSchema, NewBillingAccountIamUpdater, BillingAccountIdParseFunc),
			"google_dataproc_cluster_iam_binding":        ResourceIamBinding(IamDataprocClusterSchema, NewDataprocClusterUpdater, DataprocClusterIdParseFunc),
			"google_dataproc_cluster_iam_member":         ResourceIamMember(IamDataprocClusterSchema, NewDataprocClusterUpdater, DataprocClusterIdParseFunc),
			"google_dataproc_cluster_iam_policy":         ResourceIamPolicy(IamDataprocClusterSchema, NewDataprocClusterUpdater, DataprocClusterIdParseFunc),
			"google_dataproc_job_iam_binding":            ResourceIamBinding(IamDataprocJobSchema, NewDataprocJobUpdater, DataprocJobIdParseFunc),
			"google_dataproc_job_iam_member":             ResourceIamMember(IamDataprocJobSchema, NewDataprocJobUpdater, DataprocJobIdParseFunc),
			"google_dataproc_job_iam_policy":             ResourceIamPolicy(IamDataprocJobSchema, NewDataprocJobUpdater, DataprocJobIdParseFunc),
			"google_folder_iam_binding":                  ResourceIamBinding(IamFolderSchema, NewFolderIamUpdater, FolderIdParseFunc),
			"google_folder_iam_member":                   ResourceIamMember(IamFolderSchema, NewFolderIamUpdater, FolderIdParseFunc),
			"google_folder_iam_policy":                   ResourceIamPolicy(IamFolderSchema, NewFolderIamUpdater, FolderIdParseFunc),
			"google_folder_iam_audit_config":             ResourceIamAuditConfig(IamFolderSchema, NewFolderIamUpdater, FolderIdParseFunc),
			"google_healthcare_dataset_iam_binding":      ResourceIamBindingWithBatching(IamHealthcareDatasetSchema, NewHealthcareDatasetIamUpdater, DatasetIdParseFunc, IamBatchingEnabled),
			"google_healthcare_dataset_iam_member":       ResourceIamMemberWithBatching(IamHealthcareDatasetSchema, NewHealthcareDatasetIamUpdater, DatasetIdParseFunc, IamBatchingEnabled),
			"google_healthcare_dataset_iam_policy":       ResourceIamPolicy(IamHealthcareDatasetSchema, NewHealthcareDatasetIamUpdater, DatasetIdParseFunc),
			"google_healthcare_dicom_store_iam_binding":  ResourceIamBindingWithBatching(IamHealthcareDicomStoreSchema, NewHealthcareDicomStoreIamUpdater, DicomStoreIdParseFunc, IamBatchingEnabled),
			"google_healthcare_dicom_store_iam_member":   ResourceIamMemberWithBatching(IamHealthcareDicomStoreSchema, NewHealthcareDicomStoreIamUpdater, DicomStoreIdParseFunc, IamBatchingEnabled),
			"google_healthcare_dicom_store_iam_policy":   ResourceIamPolicy(IamHealthcareDicomStoreSchema, NewHealthcareDicomStoreIamUpdater, DicomStoreIdParseFunc),
			"google_healthcare_fhir_store_iam_binding":   ResourceIamBindingWithBatching(IamHealthcareFhirStoreSchema, NewHealthcareFhirStoreIamUpdater, FhirStoreIdParseFunc, IamBatchingEnabled),
			"google_healthcare_fhir_store_iam_member":    ResourceIamMemberWithBatching(IamHealthcareFhirStoreSchema, NewHealthcareFhirStoreIamUpdater, FhirStoreIdParseFunc, IamBatchingEnabled),
			"google_healthcare_fhir_store_iam_policy":    ResourceIamPolicy(IamHealthcareFhirStoreSchema, NewHealthcareFhirStoreIamUpdater, FhirStoreIdParseFunc),
			"google_healthcare_hl7_v2_store_iam_binding": ResourceIamBindingWithBatching(IamHealthcareHl7V2StoreSchema, NewHealthcareHl7V2StoreIamUpdater, Hl7V2StoreIdParseFunc, IamBatchingEnabled),
			"google_healthcare_hl7_v2_store_iam_member":  ResourceIamMemberWithBatching(IamHealthcareHl7V2StoreSchema, NewHealthcareHl7V2StoreIamUpdater, Hl7V2StoreIdParseFunc, IamBatchingEnabled),
			"google_healthcare_hl7_v2_store_iam_policy":  ResourceIamPolicy(IamHealthcareHl7V2StoreSchema, NewHealthcareHl7V2StoreIamUpdater, Hl7V2StoreIdParseFunc),
			"google_kms_key_ring_iam_binding":            ResourceIamBinding(IamKmsKeyRingSchema, NewKmsKeyRingIamUpdater, KeyRingIdParseFunc),
			"google_kms_key_ring_iam_member":             ResourceIamMember(IamKmsKeyRingSchema, NewKmsKeyRingIamUpdater, KeyRingIdParseFunc),
			"google_kms_key_ring_iam_policy":             ResourceIamPolicy(IamKmsKeyRingSchema, NewKmsKeyRingIamUpdater, KeyRingIdParseFunc),
			"google_kms_crypto_key_iam_binding":          ResourceIamBinding(IamKmsCryptoKeySchema, NewKmsCryptoKeyIamUpdater, CryptoIdParseFunc),
			"google_kms_crypto_key_iam_member":           ResourceIamMember(IamKmsCryptoKeySchema, NewKmsCryptoKeyIamUpdater, CryptoIdParseFunc),
			"google_kms_crypto_key_iam_policy":           ResourceIamPolicy(IamKmsCryptoKeySchema, NewKmsCryptoKeyIamUpdater, CryptoIdParseFunc),
			"google_spanner_instance_iam_binding":        ResourceIamBinding(IamSpannerInstanceSchema, NewSpannerInstanceIamUpdater, SpannerInstanceIdParseFunc),
			"google_spanner_instance_iam_member":         ResourceIamMember(IamSpannerInstanceSchema, NewSpannerInstanceIamUpdater, SpannerInstanceIdParseFunc),
			"google_spanner_instance_iam_policy":         ResourceIamPolicy(IamSpannerInstanceSchema, NewSpannerInstanceIamUpdater, SpannerInstanceIdParseFunc),
			"google_spanner_database_iam_binding":        ResourceIamBinding(IamSpannerDatabaseSchema, NewSpannerDatabaseIamUpdater, SpannerDatabaseIdParseFunc),
			"google_spanner_database_iam_member":         ResourceIamMember(IamSpannerDatabaseSchema, NewSpannerDatabaseIamUpdater, SpannerDatabaseIdParseFunc),
			"google_spanner_database_iam_policy":         ResourceIamPolicy(IamSpannerDatabaseSchema, NewSpannerDatabaseIamUpdater, SpannerDatabaseIdParseFunc),
			"google_organization_iam_binding":            ResourceIamBinding(IamOrganizationSchema, NewOrganizationIamUpdater, OrgIdParseFunc),
			"google_organization_iam_member":             ResourceIamMember(IamOrganizationSchema, NewOrganizationIamUpdater, OrgIdParseFunc),
			"google_organization_iam_policy":             ResourceIamPolicy(IamOrganizationSchema, NewOrganizationIamUpdater, OrgIdParseFunc),
			"google_organization_iam_audit_config":       ResourceIamAuditConfig(IamOrganizationSchema, NewOrganizationIamUpdater, OrgIdParseFunc),
			"google_project_iam_policy":                  ResourceIamPolicy(IamProjectSchema, NewProjectIamUpdater, ProjectIdParseFunc),
			"google_project_iam_binding":                 ResourceIamBindingWithBatching(IamProjectSchema, NewProjectIamUpdater, ProjectIdParseFunc, IamBatchingEnabled),
			"google_project_iam_member":                  ResourceIamMemberWithBatching(IamProjectSchema, NewProjectIamUpdater, ProjectIdParseFunc, IamBatchingEnabled),
			"google_project_iam_audit_config":            ResourceIamAuditConfigWithBatching(IamProjectSchema, NewProjectIamUpdater, ProjectIdParseFunc, IamBatchingEnabled),
			"google_pubsub_subscription_iam_binding":     ResourceIamBinding(IamPubsubSubscriptionSchema, NewPubsubSubscriptionIamUpdater, PubsubSubscriptionIdParseFunc),
			"google_pubsub_subscription_iam_member":      ResourceIamMember(IamPubsubSubscriptionSchema, NewPubsubSubscriptionIamUpdater, PubsubSubscriptionIdParseFunc),
			"google_pubsub_subscription_iam_policy":      ResourceIamPolicy(IamPubsubSubscriptionSchema, NewPubsubSubscriptionIamUpdater, PubsubSubscriptionIdParseFunc),
			"google_service_account_iam_binding":         ResourceIamBinding(IamServiceAccountSchema, NewServiceAccountIamUpdater, ServiceAccountIdParseFunc),
			"google_service_account_iam_member":          ResourceIamMember(IamServiceAccountSchema, NewServiceAccountIamUpdater, ServiceAccountIdParseFunc),
			"google_service_account_iam_policy":          ResourceIamPolicy(IamServiceAccountSchema, NewServiceAccountIamUpdater, ServiceAccountIdParseFunc),
			// ####### END non-generated IAM resources ###########
		},
		dclResources,
	)
}

func providerConfigure(ctx context.Context, d *schema.ResourceData, p *schema.Provider) (interface{}, diag.Diagnostics) {
	HandleSDKDefaults(d)
	HandleDCLCustomEndpointDefaults(d)

	config := Config{
		Project:             d.Get("project").(string),
		Region:              d.Get("region").(string),
		Zone:                d.Get("zone").(string),
		UserProjectOverride: d.Get("user_project_override").(bool),
		BillingProject:      d.Get("billing_project").(string),
		UserAgent:           p.UserAgent("terraform-provider-google-beta", version.ProviderVersion),
	}

	// opt in extension for adding to the User-Agent header
	if ext := os.Getenv("GOOGLE_TERRAFORM_USERAGENT_EXTENSION"); ext != "" {
		ua := config.UserAgent
		config.UserAgent = fmt.Sprintf("%s %s", ua, ext)
	}

	if v, ok := d.GetOk("request_timeout"); ok {
		var err error
		config.RequestTimeout, err = time.ParseDuration(v.(string))
		if err != nil {
			return nil, diag.FromErr(err)
		}
	}

	if v, ok := d.GetOk("request_reason"); ok {
		config.RequestReason = v.(string)
	}

	// Check for primary credentials in config. Note that if neither is set, ADCs
	// will be used if available.
	if v, ok := d.GetOk("access_token"); ok {
		config.AccessToken = v.(string)
	}

	if v, ok := d.GetOk("credentials"); ok {
		config.Credentials = v.(string)
	}

	// only check environment variables if neither value was set in config- this
	// means config beats env var in all cases.
	if config.AccessToken == "" && config.Credentials == "" {
		config.Credentials = MultiEnvSearch([]string{
			"GOOGLE_CREDENTIALS",
			"GOOGLE_CLOUD_KEYFILE_JSON",
			"GCLOUD_KEYFILE_JSON",
		})

		config.AccessToken = MultiEnvSearch([]string{
			"GOOGLE_OAUTH_ACCESS_TOKEN",
		})
	}

	// Given that impersonate_service_account is a secondary auth method, it has
	// no conflicts to worry about. We pull the env var in a DefaultFunc.
	if v, ok := d.GetOk("impersonate_service_account"); ok {
		config.ImpersonateServiceAccount = v.(string)
	}

	delegates := d.Get("impersonate_service_account_delegates").([]interface{})
	if len(delegates) > 0 {
		config.ImpersonateServiceAccountDelegates = make([]string, len(delegates))
	}
	for i, delegate := range delegates {
		config.ImpersonateServiceAccountDelegates[i] = delegate.(string)
	}

	scopes := d.Get("scopes").([]interface{})
	if len(scopes) > 0 {
		config.Scopes = make([]string, len(scopes))
	}
	for i, scope := range scopes {
		config.Scopes[i] = scope.(string)
	}

	batchCfg, err := ExpandProviderBatchingConfig(d.Get("batching"))
	if err != nil {
		return nil, diag.FromErr(err)
	}
	config.BatchingConfig = batchCfg

	// Generated products
	config.ComputeBasePath = d.Get("compute_custom_endpoint").(string)

	// Handwritten Products / Versioned / Atypical Entries
	config.CloudBillingBasePath = d.Get(CloudBillingCustomEndpointEntryKey).(string)
	config.ComposerBasePath = d.Get(ComposerCustomEndpointEntryKey).(string)
	config.ContainerBasePath = d.Get(ContainerCustomEndpointEntryKey).(string)
	config.DataflowBasePath = d.Get(DataflowCustomEndpointEntryKey).(string)
	config.IamCredentialsBasePath = d.Get(IamCredentialsCustomEndpointEntryKey).(string)
	config.ResourceManagerV3BasePath = d.Get(ResourceManagerV3CustomEndpointEntryKey).(string)
	config.RuntimeConfigBasePath = d.Get(RuntimeConfigCustomEndpointEntryKey).(string)
	config.IAMBasePath = d.Get(IAMCustomEndpointEntryKey).(string)
	config.ServiceNetworkingBasePath = d.Get(ServiceNetworkingCustomEndpointEntryKey).(string)
	config.ServiceUsageBasePath = d.Get(ServiceUsageCustomEndpointEntryKey).(string)
	config.BigtableAdminBasePath = d.Get(BigtableAdminCustomEndpointEntryKey).(string)
	config.TagsLocationBasePath = d.Get(TagsLocationCustomEndpointEntryKey).(string)

	// dcl
	config.ContainerAwsBasePath = d.Get(ContainerAwsCustomEndpointEntryKey).(string)
	config.ContainerAzureBasePath = d.Get(ContainerAzureCustomEndpointEntryKey).(string)

	stopCtx, ok := schema.StopContext(ctx)
	if !ok {
		stopCtx = ctx
	}
	if err := config.LoadAndValidate(stopCtx); err != nil {
		return nil, diag.FromErr(err)
	}

	return ProviderDCLConfigure(d, &config), nil
}

func validateCredentials(v interface{}, k string) (warnings []string, errors []error) {
	if v == nil || v.(string) == "" {
		return
	}
	creds := v.(string)
	// if this is a path and we can stat it, assume it's ok
	if _, err := os.Stat(creds); err == nil {
		return
	}
	if _, err := googleoauth.CredentialsFromJSON(context.Background(), []byte(creds)); err != nil {
		errors = append(errors,
			fmt.Errorf("JSON credentials are not valid: %s", err))
	}

	return
}
