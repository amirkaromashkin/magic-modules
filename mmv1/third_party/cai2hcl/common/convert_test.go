package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindMatchingConverterName(t *testing.T) {
	assetNameRegexpConverterPairs := []RegexpNamePair{
		{AssetNameRegexp: "projects/(?P<project>[^/]+)/regions/(?P<region>[^/]+)/forwardingRules", ConverterName: "google_compute_forwarding_rule"},
		{AssetNameRegexp: "projects/(?P<project>[^/]+)/global/forwardingRules", ConverterName: "compute_global_forwarding_rule"},
	}

	name, ok := findMatchingConverterName("//compute.googleapis.com/projects/myproj/regions/us-central1/forwardingRules/test-1", assetNameRegexpConverterPairs)
	assert.True(t, ok)
	assert.Equal(t, name, "google_compute_forwarding_rule")

	name, ok = findMatchingConverterName("projects/myproj/regions/us-central1/forwardingRules/test-1", assetNameRegexpConverterPairs)
	assert.True(t, ok)
	assert.Equal(t, name, "google_compute_forwarding_rule")

	name, ok = findMatchingConverterName("//compute.googleapis.com/projects/myproj/global/forwardingRules/fr", assetNameRegexpConverterPairs)
	assert.True(t, ok)
	assert.Equal(t, name, "compute_global_forwarding_rule")

	name, ok = findMatchingConverterName("projects/myproj/global/forwardingRules/fr", assetNameRegexpConverterPairs)
	assert.True(t, ok)
	assert.Equal(t, name, "compute_global_forwarding_rule")
}
