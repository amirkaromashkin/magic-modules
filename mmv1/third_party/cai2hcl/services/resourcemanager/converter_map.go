package resourcemanager

import (
	"github.com/GoogleCloudPlatform/terraform-google-conversion/v2/cai2hcl/common"
)

var ConverterNamesPerAssetType = map[string]string{
	ProjectAssetType:        "google_project",
	ProjectBillingAssetType: "google_project",
}

var ConverterNamesPerAssetRegex = map[string]string{}

var ConverterMap = common.CreateConverterMap(map[string]common.ConverterFactory{
	"google_project": NewProjectConverter,
})
