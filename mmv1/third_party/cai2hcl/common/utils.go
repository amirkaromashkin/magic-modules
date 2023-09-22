package common

import (
	"encoding/json"
	"fmt"
	"strings"

	hashicorpcty "github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	tpg_provider "github.com/hashicorp/terraform-provider-google-beta/google-beta/provider"
	tpg_transport "github.com/hashicorp/terraform-provider-google-beta/google-beta/transport"
	"github.com/zclconf/go-cty/cty"
	ctyjson "github.com/zclconf/go-cty/cty/json"
)

// Extracts named part from resource url.
func ParseFieldValue(url string, name string) string {
	fragments := strings.Split(url, "/")
	for ix, item := range fragments {
		if item == name && ix+1 < len(fragments) {
			return fragments[ix+1]
		}
	}
	return ""
}

// Decodes the map object into the target struct.
func DecodeJSON(data map[string]interface{}, v interface{}) error {
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	return nil
}

// Converts resource from untyped map format to TF JSON.
func MapToCtyValWithSchema(m map[string]interface{}, s map[string]*schema.Schema) (cty.Value, error) {
	// Normalize non-marshable properties manually.
	m = normalizeFlattenedMap(m).(map[string]interface{})

	b, err := json.Marshal(&m)
	if err != nil {
		return cty.NilVal, fmt.Errorf("error marshaling map as JSON: %v", err)
	}
	ty, err := hashicorpCtyTypeToZclconfCtyType(schema.InternalMap(s).CoreConfigSchema().ImpliedType())
	if err != nil {
		return cty.NilVal, fmt.Errorf("error casting type: %v", err)
	}
	ret, err := ctyjson.Unmarshal(b, ty)
	if err != nil {
		return cty.NilVal, fmt.Errorf("error unmarshaling JSON as cty.Value: %v", err)
	}
	return ret, nil
}

// Initializes map of converters.
func CreateConverterMap(converterFactories map[string]ConverterFactory) map[string]Converter {
	provider := tpg_provider.Provider()

	result := make(map[string]Converter, len(converterFactories))
	for name, factory := range converterFactories {
		result[name] = factory(name, provider.ResourcesMap[name].Schema)
	}

	return result
}

func NewConfig() *tpg_transport.Config {
	// Currently its not needed, but it may change in future.
	return &tpg_transport.Config{
		Project:   "",
		Zone:      "",
		Region:    "",
		UserAgent: "",
	}
}

func hashicorpCtyTypeToZclconfCtyType(t hashicorpcty.Type) (cty.Type, error) {
	b, err := json.Marshal(t)
	if err != nil {
		return cty.NilType, err
	}
	var ret cty.Type
	if err := json.Unmarshal(b, &ret); err != nil {
		return cty.NilType, err
	}
	return ret, nil
}

// Normalizes the output map by eliminating unmarshallable objects like schema.Set
func normalizeFlattenedMap(obj interface{}) interface{} {
	switch obj.(type) {
	case []interface{}:
		arr := obj.([]interface{})
		newArr := make([]interface{}, len(arr))
		for i := range arr {
			newArr[i] = normalizeFlattenedMap(arr[i])
		}

		return newArr
	case map[string]interface{}:
		mp := obj.(map[string]interface{})
		newMap := map[string]interface{}{}
		for key, value := range mp {
			newMap[key] = normalizeFlattenedMap(value)
		}
		return newMap
	case *schema.Set:
		return obj.(*schema.Set).List()
	default:
		return obj
	}
}
