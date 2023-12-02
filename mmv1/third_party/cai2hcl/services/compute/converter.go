package compute

import (
	"github.com/GoogleCloudPlatform/terraform-google-conversion/v5/cai2hcl/common"
	"github.com/GoogleCloudPlatform/terraform-google-conversion/v5/cai2hcl/common/matchers"
)

var forwardingRuleMatchers = []matchers.ConverterMatcher{
	matchers.ByAssetName(ComputeForwardingRuleAssetNameRegex, "google_compute_forwarding_rule"),
	matchers.ByAssetName(ComputeGlobalForwardingRuleAssetNameRegex, "google_compute_global_forwarding_rule"),
}

var backendServiceMatchers = []matchers.ConverterMatcher{
	matchers.ByAssetName(ComputeRegionBackendServiceAssetNameRegex, "google_compute_region_backend_service"),
	matchers.ByAssetName(ComputeBackendServiceAssetNameRegex, "google_compute_backend_service"),
}

var healthCheckMatchers = []matchers.ConverterMatcher{
	matchers.ByAssetName(ComputeRegionHealthCheckAssetNameRegex, "google_compute_region_health_check"),
	matchers.ByAssetName(ComputeHealthCheckAssetNameRegex, "google_compute_health_check"),
}

var UberConverter = common.UberConverter{
	ConverterByAssetType: map[string]string{
		ComputeInstanceAssetType: "google_compute_instance",
	},
	ConverterMatchersByAssetType: map[string][]matchers.ConverterMatcher{
		ComputeRegionBackendServiceAssetType: backendServiceMatchers,
		ComputeBackendServiceAssetType:       backendServiceMatchers,

		ComputeForwardingRuleAssetType:       forwardingRuleMatchers,
		ComputeGlobalForwardingRuleAssetType: forwardingRuleMatchers,

		ComputeRegionHealthCheckAssetType: healthCheckMatchers,
		ComputeHealthCheckAssetType:       healthCheckMatchers,
	},
	Converters: common.CreateConverterMap(map[string]common.ConverterFactory{
		"google_compute_instance":               NewComputeInstanceConverter,
		"google_compute_region_backend_service": NewComputeRegionBackendServiceConverter,
		"google_compute_backend_service":        NewComputeBackendServiceConverter,
		"google_compute_forwarding_rule":        NewComputeForwardingRuleConverter,
		"google_compute_global_forwarding_rule": NewComputeGlobalForwardingRuleConverter,
		"google_compute_region_health_check":    NewComputeRegionHealthCheckConverter,
		"google_compute_health_check":           NewComputeHealthCheckConverter,
	}),
}
