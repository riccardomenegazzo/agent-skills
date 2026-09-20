package examples

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// ValidateCloudFormation checks the small set of structural invariants that
// fenced examples can verify without pretending to replace AWS service-side
// schema validation.
func ValidateCloudFormation(content string) error {
	var root map[string]any
	if err := yaml.Unmarshal([]byte(content), &root); err != nil {
		return fmt.Errorf("cloudformation: not valid YAML: %w", err)
	}

	resourcesRaw, ok := root["Resources"]
	if !ok {
		return fmt.Errorf("cloudformation: missing Resources")
	}
	resources, ok := resourcesRaw.(map[string]any)
	if !ok || len(resources) == 0 {
		return fmt.Errorf("cloudformation: Resources must be a non-empty mapping")
	}

	for logicalID, raw := range resources {
		resource, ok := raw.(map[string]any)
		if !ok {
			return fmt.Errorf("cloudformation: resource %q must be a mapping", logicalID)
		}
		resourceType, ok := resource["Type"].(string)
		if !ok || resourceType == "" {
			return fmt.Errorf("cloudformation: resource %q has invalid Type", logicalID)
		}
		if resourceType != "AWS::Lambda::Function" && resourceType != "AWS::Serverless::Function" {
			continue
		}

		propertiesRaw, ok := resource["Properties"]
		if !ok {
			return fmt.Errorf("cloudformation: Lambda resource %q is missing Properties", logicalID)
		}
		properties, ok := propertiesRaw.(map[string]any)
		if !ok {
			return fmt.Errorf("cloudformation: Lambda resource %q Properties must be a mapping", logicalID)
		}

		packageType, packageTypeKnown := properties["PackageType"].(string)
		if packageType == "" {
			packageType = "Zip"
			packageTypeKnown = true
		}
		if _, hasLayers := properties["Layers"]; packageTypeKnown && packageType == "Image" && hasLayers {
			return fmt.Errorf("cloudformation: Lambda resource %q uses PackageType Image and cannot use Layers", logicalID)
		}
	}

	return nil
}
