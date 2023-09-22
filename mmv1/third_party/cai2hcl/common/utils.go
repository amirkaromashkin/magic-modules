package common

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/GoogleCloudPlatform/terraform-google-conversion/v2/caiasset"
	hashicorpcty "github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/hcl/hcl/printer"
	"github.com/hashicorp/hcl/v2/hclwrite"
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

func Convert(assets []*caiasset.Asset, converterNames map[string]string, converterMap map[string]Converter) ([]byte, error) {
	// Group resources from the same tf resource type for convert.
	// tf -> cai has 1:N mappings occasionally
	groups := make(map[string][]*caiasset.Asset)
	for _, asset := range assets {
		name, ok := converterNames[asset.Type]
		if !ok {
			continue
		}
		groups[name] = append(groups[name], asset)
	}

	f := hclwrite.NewFile()
	rootBody := f.Body()
	for name, v := range groups {
		converter, ok := converterMap[name]
		if !ok {
			continue
		}
		items, err := converter.Convert(v)
		if err != nil {
			return nil, err
		}

		for _, resourceBlock := range items {
			hclBlock := rootBody.AppendNewBlock("resource", resourceBlock.Labels)
			if err := hclWriteBlock(resourceBlock.Value, hclBlock.Body()); err != nil {
				return nil, err
			}
		}
		if err != nil {
			return nil, err
		}
	}

	return printer.Format(f.Bytes())
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

func hclWriteBlock(val cty.Value, body *hclwrite.Body) error {
	if val.IsNull() {
		return nil
	}
	if !val.Type().IsObjectType() {
		return fmt.Errorf("expect object type only, but type = %s", val.Type().FriendlyName())
	}
	it := val.ElementIterator()
	for it.Next() {
		objKey, objVal := it.Element()
		if objVal.IsNull() {
			continue
		}
		objValType := objVal.Type()
		switch {
		case objValType.IsObjectType():
			newBlock := body.AppendNewBlock(objKey.AsString(), nil)
			if err := hclWriteBlock(objVal, newBlock.Body()); err != nil {
				return err
			}
		case objValType.IsCollectionType():
			if objVal.LengthInt() == 0 {
				continue
			}
			// Presumes map should not contain object type.
			if !objValType.IsMapType() && objValType.ElementType().IsObjectType() {
				listIterator := objVal.ElementIterator()
				for listIterator.Next() {
					_, listVal := listIterator.Element()
					subBlock := body.AppendNewBlock(objKey.AsString(), nil)
					if err := hclWriteBlock(listVal, subBlock.Body()); err != nil {
						return err
					}
				}
				continue
			}
			fallthrough
		default:
			if objValType.FriendlyName() == "string" && objVal.AsString() == "" {
				continue
			}
			body.SetAttributeValue(objKey.AsString(), objVal)
		}
	}
	return nil
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
