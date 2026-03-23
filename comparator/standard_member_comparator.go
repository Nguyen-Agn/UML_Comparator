package comparator

import (
	"strings"
	"uml_compare/domain"
	"uml_compare/matcher"
)

// StandardMemberComparator implements IMemberComparator for class-level attribute and method comparison.
type StandardMemberComparator struct {
	fuzzyMatcher matcher.IFuzzyMatcher
	typeAnalyzer ITypeAnalyzer
}

var _ IMemberComparator = (*StandardMemberComparator)(nil)

// NewStandardMemberComparator creates a new instance of StandardMemberComparator.
func NewStandardMemberComparator(fz matcher.IFuzzyMatcher, ta ITypeAnalyzer) *StandardMemberComparator {
	return &StandardMemberComparator{
		fuzzyMatcher: fz,
		typeAnalyzer: ta,
	}
}

// CompareAttributes identifies differences in attributes between solution and student nodes.
func (v *StandardMemberComparator) CompareAttributes(sol, stu domain.ProcessedNode, typeMap map[string]string, report *domain.DiffReport) {
	stuAttrs := make([]domain.ProcessedAttribute, len(stu.Attributes))
	copy(stuAttrs, stu.Attributes)
	matchedStuAttrIdx := make(map[int]bool)

	for _, sAttr := range sol.Attributes {
		foundIdx := -1
		
		// 1. Try perfect match (Type + Name)
		for i, stAttr := range stuAttrs {
			if matchedStuAttrIdx[i] { continue }
			if v.typeAnalyzer.CompareTypes(sAttr.Type, stAttr.Type, typeMap) {
				if strings.EqualFold(sAttr.Name, stAttr.Name) {
					foundIdx = i
					break
				}
			}
		}
		
		// 2. Try name-only match (fuzzy)
		if foundIdx == -1 {
			for i, stAttr := range stuAttrs {
				if matchedStuAttrIdx[i] { continue }
				if v.fuzzyMatcher.Compare(sAttr.Name, stAttr.Name) >= 0.8 {
					foundIdx = i
					break
				}
			}
		}

		if foundIdx != -1 {
			matchedStuAttrIdx[foundIdx] = true
			matchingStu := stuAttrs[foundIdx]
			issues := []string{}
			
			if !v.typeAnalyzer.CompareTypes(sAttr.Type, matchingStu.Type, typeMap) {
				issues = append(issues, "Type mismatch (Sol: "+sAttr.Type+", Stu: "+matchingStu.Type+")")
			}
			if sAttr.Scope != matchingStu.Scope {
				issues = append(issues, "Scope mismatch ("+sAttr.Scope+" vs "+matchingStu.Scope+")")
			}
			if sAttr.Kind != matchingStu.Kind {
				issues = append(issues, "Kind mismatch ("+sAttr.Kind+" vs "+matchingStu.Kind+")")
			}

			if len(issues) > 0 {
				report.WrongDetail.Attribute = append(report.WrongDetail.Attribute, domain.AttributeDiff{ParentClassName: sol.Name, Sol: &sAttr, Stu: &matchingStu, Description: strings.Join(issues, ", ")})
			} else {
				report.CorrectDetail.Attribute = append(report.CorrectDetail.Attribute, domain.AttributeDiff{ParentClassName: sol.Name, Sol: &sAttr, Stu: &matchingStu, Description: "Match"})
			}
		} else {
			report.MissingDetail.Attribute = append(report.MissingDetail.Attribute, domain.AttributeDiff{ParentClassName: sol.Name, Sol: &sAttr, Stu: nil, Description: "Missing attribute (" + sAttr.Scope + " " + sAttr.Type + ")"})
		}
	}

	for i := range stuAttrs {
		stAttr := &stuAttrs[i]
		if !matchedStuAttrIdx[i] {
			report.ExtraDetail.Attribute = append(report.ExtraDetail.Attribute, domain.AttributeDiff{ParentClassName: stu.Name, Sol: nil, Stu: stAttr, Description: "Extra attribute (" + stAttr.Scope + " " + stAttr.Type + ")"})
		}
	}
}

// CompareMethods identifies differences in methods between solution and student nodes.
func (v *StandardMemberComparator) CompareMethods(sol, stu domain.ProcessedNode, typeMap map[string]string, report *domain.DiffReport) {
	solG, solS, solNormal := v.splitMethods(sol.Methods)
	stuG, stuS, stuNormal := v.splitMethods(stu.Methods)

	// Getter/Setter Count logic
	if (sol.Shortcut&1) == 0 && (stu.Shortcut&1) == 0 {
		if len(solG) != len(stuG) {
			report.WrongDetail.Method = append(report.WrongDetail.Method, domain.MethodDiff{ParentClassName: sol.Name, Sol: nil, Stu: nil, Description: strings.Join([]string{"Expected", itoa(len(solG)), "getter(s), got", itoa(len(stuG))}, " ")})
		} else if len(solG) > 0 {
			report.CorrectDetail.Method = append(report.CorrectDetail.Method, domain.MethodDiff{ParentClassName: sol.Name, Sol: nil, Stu: nil, Description: itoa(len(solG)) + " getter(s) match"})
		}
	}
	if (sol.Shortcut&2) == 0 && (stu.Shortcut&2) == 0 {
		if len(solS) != len(stuS) {
			report.WrongDetail.Method = append(report.WrongDetail.Method, domain.MethodDiff{ParentClassName: sol.Name, Sol: nil, Stu: nil, Description: strings.Join([]string{"Expected", itoa(len(solS)), "setter(s), got", itoa(len(stuS))}, " ")})
		} else if len(solS) > 0 {
			report.CorrectDetail.Method = append(report.CorrectDetail.Method, domain.MethodDiff{ParentClassName: sol.Name, Sol: nil, Stu: nil, Description: itoa(len(solS)) + " setter(s) match"})
		}
	}

	// Normal Methods
	matchedStuMethIdx := make(map[int]bool)
	for _, sMethod := range solNormal {
		isCtor := v.isConstructor(sMethod, sol.Name)
		foundIdx := -1
		
		// 1. Try perfect match (Name + RetType + ParamCount)
		for i, stMethod := range stuNormal {
			if matchedStuMethIdx[i] { continue }
			if v.typeAnalyzer.CompareTypes(sMethod.Output, stMethod.Output, typeMap) || isCtor {
				if len(sMethod.Inputs) == len(stMethod.Inputs) {
					if v.matchMethodName(sMethod, stMethod, isCtor, stu.Name) {
						foundIdx = i
						break
					}
				}
			}
		}

		// 2. Try signature-ish match (Name + ParamCount +-1 rule)
		if foundIdx == -1 {
			for i, stMethod := range stuNormal {
				if matchedStuMethIdx[i] { continue }
				
				solPLen := len(sMethod.Inputs)
				stuPLen := len(stMethod.Inputs)
				paramCountMatch := (solPLen == stuPLen)
				if solPLen >= 2 && stuPLen >= 2 {
					diff := solPLen - stuPLen
					if diff < 0 { diff = -diff }
					if diff <= 1 { paramCountMatch = true }
				}

				if paramCountMatch {
					if v.matchMethodName(sMethod, stMethod, isCtor, stu.Name) {
						foundIdx = i
						break
					}
				}
			}
		}

		// 3. Try fuzzy match
		if foundIdx == -1 {
			for i, stMethod := range stuNormal {
				if matchedStuMethIdx[i] { continue }
				if v.fuzzyMatcher.Compare(sMethod.Name, stMethod.Name) >= 0.8 {
					foundIdx = i
					break
				}
			}
		}

		if foundIdx != -1 {
			matchedStuMethIdx[foundIdx] = true
			matchingStu := stuNormal[foundIdx]
			issues := []string{}
			
			if !isCtor && !v.typeAnalyzer.CompareTypes(sMethod.Output, matchingStu.Output, typeMap) {
				issues = append(issues, "Return type mismatch (Sol: "+sMethod.Output+", Stu: "+matchingStu.Output+")")
			}
			if sMethod.Scope != matchingStu.Scope {
				issues = append(issues, "Scope mismatch ("+sMethod.Scope+" vs "+matchingStu.Scope+")")
			}
			if sMethod.Kind != matchingStu.Kind {
				issues = append(issues, "Kind mismatch ("+sMethod.Kind+" vs "+matchingStu.Kind+")")
			}
			if len(sMethod.Inputs) != len(matchingStu.Inputs) {
				issues = append(issues, "Param count mismatch ("+itoa(len(sMethod.Inputs))+" vs "+itoa(len(matchingStu.Inputs))+")")
			} else {
				for j := range sMethod.Inputs {
					if !v.typeAnalyzer.CompareTypes(sMethod.Inputs[j].Type, matchingStu.Inputs[j].Type, typeMap) {
						issues = append(issues, "Param "+itoa(j+1)+" type mismatch")
						break
					}
				}
			}

			if len(issues) > 0 {
				report.WrongDetail.Method = append(report.WrongDetail.Method, domain.MethodDiff{ParentClassName: sol.Name, Sol: &sMethod, Stu: &matchingStu, Description: strings.Join(issues, ", ")})
			} else {
				report.CorrectDetail.Method = append(report.CorrectDetail.Method, domain.MethodDiff{ParentClassName: sol.Name, Sol: &sMethod, Stu: &matchingStu, Description: "Match"})
			}
		} else {
			if isCtor {
				report.MissingDetail.Method = append(report.MissingDetail.Method, domain.MethodDiff{ParentClassName: sol.Name, Sol: &sMethod, Stu: nil, Description: "Missing constructor"})
			} else {
				report.MissingDetail.Method = append(report.MissingDetail.Method, domain.MethodDiff{ParentClassName: sol.Name, Sol: &sMethod, Stu: nil, Description: "Missing method (" + sMethod.Scope + " " + sMethod.Output + ")"})
			}
		}
	}

	for i := range stuNormal {
		stMethod := &stuNormal[i]
		if !matchedStuMethIdx[i] {
			report.ExtraDetail.Method = append(report.ExtraDetail.Method, domain.MethodDiff{ParentClassName: stu.Name, Sol: nil, Stu: stMethod, Description: "Extra method (" + stMethod.Scope + " " + stMethod.Output + ")"})
		}
	}
}

// splitMethods partitions methods into getters, setters, and others.
func (v *StandardMemberComparator) splitMethods(methods []domain.ProcessedMethod) (g, s, normal []domain.ProcessedMethod) {
	for _, m := range methods {
		switch m.Type {
		case "getter":
			g = append(g, m)
		case "setter":
			s = append(s, m)
		default:
			normal = append(normal, m)
		}
	}
	return
}

// isConstructor checks if a method is likely a constructor for a given class.
func (v *StandardMemberComparator) isConstructor(m domain.ProcessedMethod, className string) bool {
	return strings.EqualFold(m.Name, className) || strings.EqualFold(m.Name, "init") || strings.EqualFold(m.Name, "<<create>>")
}

// matchMethodName performs fuzzy matching on method names with special logic for constructors.
func (v *StandardMemberComparator) matchMethodName(sol, stu domain.ProcessedMethod, solIsCtor bool, stuClassName string) bool {
	if solIsCtor {
		return v.isConstructor(stu, stuClassName)
	}
	return v.fuzzyMatcher.Compare(sol.Name, stu.Name) >= 0.5
}

// itoa is a helper for integer to string conversion.
func itoa(n int) string {
	if n == 0 { return "0" }
	neg := n < 0
	if neg { n = -n }
	buf := [20]byte{}
	pos := len(buf)
	for n > 0 {
		pos--
		buf[pos] = byte(n%10) + '0'
		n /= 10
	}
	if neg {
		pos--
		buf[pos] = '-'
	}
	return string(buf[pos:])
}
