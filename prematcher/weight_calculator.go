package prematcher

import "strings"

// WeightCalculator implements IWeightCalculator.
type WeightCalculator struct {
	detector ITypeDetector
}

// NewWeightCalculator creates a new instance of WeightCalculator.
func NewWeightCalculator(detector ITypeDetector) *WeightCalculator {
	return &WeightCalculator{
		detector: detector,
	}
}

// Calculate uses bitwise shifting to pack structural info into a single uint32.
func (c *WeightCalculator) Calculate(
	nodeType string,
	hasInheritance bool,
	numInterfaces int,
	numMethods int,
	numAttributes int,
	numRelated int,
	numCustomTypes int,
	numStaticMembers int,
) uint32 {
	var weight uint32 = 0

	// 1. Loại Class (Bit 29-31)
	var typeVal uint32 = 0
	lowerType := strings.ToLower(nodeType)
	if (strings.Contains(lowerType, "class") || lowerType == "default") && !strings.Contains(lowerType, "abstract") {
		typeVal = 1
	} else if strings.Contains(lowerType, "interface") {
		typeVal = 2
	} else if strings.Contains(lowerType, "abstract") {
		typeVal = 3
	} else if c.detector.IsEnumType(nodeType) {
		typeVal = 4
	}
	weight |= (typeVal & 0x7) << 29

	// 2. Thừa kế (Bit 28)
	if hasInheritance {
		weight |= (1 & 0x1) << 28
	}

	// 3. Số lượng Interface (Bit 24-27)
	weight |= (minU32(uint32(numInterfaces), 15) & 0xF) << 24

	// 4. Số lượng Method (Bit 18-23)
	weight |= (minU32(uint32(numMethods), 63) & 0x3F) << 18

	// 5. Số lượng Attribute (Bit 13-17)
	weight |= (minU32(uint32(numAttributes), 31) & 0x1F) << 13

	// 6. Số lượng Class liên quan (Bit 9-12)
	weight |= (minU32(uint32(numRelated), 15) & 0xF) << 9

	// 7. Số lượng Custom Type (Bit 6-8)
	weight |= (minU32(uint32(numCustomTypes), 7) & 0x7) << 6

	// 8. Số lượng Static members (Bit 2-5)
	weight |= (minU32(uint32(numStaticMembers), 15) & 0xF) << 2

	return weight
}
