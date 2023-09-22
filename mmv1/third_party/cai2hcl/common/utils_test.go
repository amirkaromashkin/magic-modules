package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTryGetConverterNameByAssetNameRegex(t *testing.T) {
	converterNamesPerRegex := map[string]string{
		"projects/(?P<project>[^/]+)/regions/(?P<region>[^/]+)/forwardingRules": "google_compute_forwarding_rule",
		"projects/(?P<project>[^/]+)/global/forwardingRules":                    "compute_global_forwarding_rule",
	}

	name, ok := tryGetConverterNameByAssetNameRegex("//compute.googleapis.com/projects/myproj/regions/us-central1/forwardingRules/test-1", converterNamesPerRegex)
	assert.True(t, ok)
	assert.Equal(t, name, "google_compute_forwarding_rule")

	name, ok = tryGetConverterNameByAssetNameRegex("projects/myproj/regions/us-central1/forwardingRules/test-1", converterNamesPerRegex)
	assert.True(t, ok)
	assert.Equal(t, name, "google_compute_forwarding_rule")

	name, ok = tryGetConverterNameByAssetNameRegex("//compute.googleapis.com/projects/myproj/global/forwardingRules/fr", converterNamesPerRegex)
	assert.True(t, ok)
	assert.Equal(t, name, "compute_global_forwarding_rule")

	name, ok = tryGetConverterNameByAssetNameRegex("projects/myproj/global/forwardingRules/fr", converterNamesPerRegex)
	assert.True(t, ok)
	assert.Equal(t, name, "compute_global_forwarding_rule")
}
