package builder

import (
	"fmt"
)

// GetBuilder returns the appropriate IModelBuilder based on the source type.
// Currently supported: "drawio".
// In the future, this can support "json", "yaml", "java", etc.
func GetBuilder(sourceType string) (IModelBuilder, error) {
	switch sourceType {
	case "drawio":
		return NewDrawioModelBuilder(), nil
	default:
		return nil, fmt.Errorf("GetBuilder: unsupported source type %q", sourceType)
	}
}
