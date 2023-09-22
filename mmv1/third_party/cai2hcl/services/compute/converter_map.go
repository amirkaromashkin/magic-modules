package compute

import (
	"github.com/GoogleCloudPlatform/terraform-google-conversion/v2/cai2hcl/common"
)

var ConverterNames = map[string]string{
	// Custom converters
	ComputeInstanceAssetType: "google_compute_instance",
	// Generated converters
	ComputeForwardingRuleAssetType:       "google_compute_forwarding_rule",
	ComputeGlobalForwardingRuleAssetType: "google_compute_global_forwarding_rule",
	ComputeRegionBackendServiceAssetType: "google_compute_region_backend_service",
	ComputeBackendServiceAssetType:       "google_compute_backend_service",
	ComputeHealthCheckAssetType:          "google_compute_health_check",
}

var ConverterMap = common.CreateConverterMap(map[string]common.ConverterFactory{
	// Custom converters
	"google_compute_instance": NewComputeInstanceConverter,
	// Generated converters
	"google_compute_forwarding_rule":        NewComputeForwardingRuleConverter,
	"google_compute_global_forwarding_rule": NewComputeGlobalForwardingRuleConverter,
	"google_compute_region_backend_service": NewComputeRegionBackendServiceConverter,
	"google_compute_backend_service":        NewComputeBackendServiceConverter,
	"google_compute_health_check":           NewComputeHealthCheckConverter,
})

var TestsFolder = "./testdata"

var TestsMap = map[string][]string{
	// Custom converters
	"google_compute_instance": {"full_compute_instance", "compute_instance_iam"},
	// Generated converters
	"google_compute_forwarding_rule":        {"full_compute_forwarding_rule"},
	"google_compute_global_forwarding_rule": {"full_compute_global_forwarding_rule"},
	"google_compute_region_backend_service": {"full_compute_region_backend_service"},
	"google_compute_backend_service":        {"full_compute_backend_service"},
	"google_compute_health_check":           {"full_compute_health_check"},
}
