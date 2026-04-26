package comparator

import (
	"strings"
	"uml_compare/domain"
	"uml_compare/src/matcher"
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
func (v *StandardMemberComparator) CompareAttributes(sol domain.SolutionProcessedNode, stu domain.ProcessedNode, typeMap map[string]string, report *domain.DiffReport) {
	stuAttrs := make([]domain.ProcessedAttribute, len(stu.Attributes))
	copy(stuAttrs, stu.Attributes)
	matchedStuAttrIdx := make(map[int]bool)

	for _, sAttr := range sol.Attributes {
		foundIdx := -1

		// 1. Try perfect match (Type + Name)
		for i, stAttr := range stuAttrs {
			if matchedStuAttrIdx[i] {
				continue
			}

			matchedType := false
			for _, t := range sAttr.Types {
				if v.typeAnalyzer.CompareTypes(t, stAttr.Type, typeMap) {
					matchedType = true
					break
				}
			}
			if matchedType {
				matchedName := false
				for _, n := range sAttr.Names {
					if strings.EqualFold(n, stAttr.Name) {
						matchedName = true
						break
					}
				}
				if matchedName {
					foundIdx = i
					break
				}
			}
		}

		// 2. Try name-only match (fuzzy)
		if foundIdx == -1 {
			for i, stAttr := range stuAttrs {
				if matchedStuAttrIdx[i] {
					continue
				}

				matchedName := false
				for _, n := range sAttr.Names {
					if v.fuzzyMatcher.Compare(n, stAttr.Name) >= 0.8 {
						matchedName = true
						break
					}
				}
				if matchedName {
					foundIdx = i
					break
				}
			}
		}

		if foundIdx != -1 {
			matchedStuAttrIdx[foundIdx] = true
			matchingStu := stuAttrs[foundIdx]
			issues := []string{}

			matchedType := false
			for _, t := range sAttr.Types {
				if v.typeAnalyzer.CompareTypes(t, matchingStu.Type, typeMap) {
					matchedType = true
					break
				}
			}

			if !matchedType {
				issues = append(issues, "Type mismatch (Sol: "+strings.Join(sAttr.Types, "|")+", Stu: "+matchingStu.Type+")")
			}
			if sAttr.Scope != matchingStu.Scope {
				issues = append(issues, "Scope mismatch ("+sAttr.Scope+" vs "+matchingStu.Scope+")")
			}
			if sAttr.Kind != matchingStu.Kind {
				issues = append(issues, "Kind mismatch ("+sAttr.Kind+" vs "+matchingStu.Kind+")")
			}

			// We need a pointer to store in DiffReport but we're ranging over values, so we make a local copy
			solAttrCopy := sAttr
			if len(issues) > 0 {
				report.WrongDetail.Attribute = append(report.WrongDetail.Attribute, domain.AttributeDiff{ParentClassName: sol.Name, Sol: &solAttrCopy, Stu: &matchingStu, Description: strings.Join(issues, ", ")})
			} else {
				report.CorrectDetail.Attribute = append(report.CorrectDetail.Attribute, domain.AttributeDiff{ParentClassName: sol.Name, Sol: &solAttrCopy, Stu: &matchingStu, Description: "Match"})
			}
		} else {
			solAttrCopy := sAttr
			report.MissingDetail.Attribute = append(report.MissingDetail.Attribute, domain.AttributeDiff{ParentClassName: sol.Name, Sol: &solAttrCopy, Stu: nil, Description: "Missing attribute (" + sAttr.Scope + " " + strings.Join(sAttr.Types, "|") + ")"})
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
func (v *StandardMemberComparator) CompareMethods(sol domain.SolutionProcessedNode, stu domain.ProcessedNode, typeMap map[string]string, report *domain.DiffReport) {
	solG, solS, solNormal := v.splitSolutionMethods(sol.Methods)

	stuMethods := make([]domain.ProcessedMethod, len(stu.Methods))
	copy(stuMethods, stu.Methods)
	matchedStuMethIdx := make(map[int]bool)

	// Normal Methods
	for _, sMethod := range solNormal {
		isCtor := v.isConstructor(sMethod, sol.Name)
		foundIdx := -1

		// 1. Try perfect match (Name + RetType + ParamCount)
		for i, stMethod := range stuMethods {
			if matchedStuMethIdx[i] {
				continue
			}

			matchedOutput := isCtor
			if !isCtor {
				for _, out := range sMethod.Outputs {
					if v.typeAnalyzer.CompareTypes(out, stMethod.Output, typeMap) {
						matchedOutput = true
						break
					}
				}
			}

			if matchedOutput || isCtor {
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
			for i, stMethod := range stuMethods {
				if matchedStuMethIdx[i] {
					continue
				}

				solPLen := len(sMethod.Inputs)
				stuPLen := len(stMethod.Inputs)
				paramCountMatch := (solPLen == stuPLen)
				if solPLen >= 2 && stuPLen >= 2 {
					diff := solPLen - stuPLen
					if diff < 0 {
						diff = -diff
					}
					if diff <= 1 {
						paramCountMatch = true
					}
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
			for i, stMethod := range stuMethods {
				if matchedStuMethIdx[i] {
					continue
				}
				matchedName := false
				for _, n := range sMethod.Names {
					if v.fuzzyMatcher.Compare(n, stMethod.Name) >= 0.8 {
						matchedName = true
						break
					}
				}
				if matchedName {
					foundIdx = i
					break
				}
			}
		}

		if foundIdx != -1 {
			matchedStuMethIdx[foundIdx] = true
			matchingStu := stuMethods[foundIdx]
			issues := []string{}

			matchedOutput := isCtor
			if !isCtor {
				for _, out := range sMethod.Outputs {
					if v.typeAnalyzer.CompareTypes(out, matchingStu.Output, typeMap) {
						matchedOutput = true
						break
					}
				}
			}

			if !matchedOutput {
				issues = append(issues, "Return type mismatch (Sol: "+strings.Join(sMethod.Outputs, "|")+", Stu: "+matchingStu.Output+")")
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
				if isCtor {
					matchedStuParams := make(map[int]bool)
					for _, sParam := range sMethod.Inputs {
						foundMatch := false
						for j, stuParam := range matchingStu.Inputs {
							if matchedStuParams[j] {
								continue
							}
							for _, solType := range sParam.Types {
								if v.typeAnalyzer.CompareTypes(solType, stuParam.Type, typeMap) {
									matchedStuParams[j] = true
									foundMatch = true
									break
								}
							}
							if foundMatch {
								break
							}
						}
						if !foundMatch {
							solTypeStr := strings.Join(sParam.Types, "|")
							issues = append(issues, "Param type '"+solTypeStr+"' not found")
						}
					}
				} else {
					for j := range sMethod.Inputs {
						matchedType := false
						for _, solType := range sMethod.Inputs[j].Types {
							if v.typeAnalyzer.CompareTypes(solType, matchingStu.Inputs[j].Type, typeMap) {
								matchedType = true
								break
							}
						}
						if !matchedType {
							issues = append(issues, "Param "+itoa(j+1)+" type mismatch")
							break
						}
					}
				}
			}

			solMethodCopy := sMethod
			if len(issues) > 0 {
				report.WrongDetail.Method = append(report.WrongDetail.Method, domain.MethodDiff{ParentClassName: sol.Name, Sol: &solMethodCopy, Stu: &matchingStu, Description: strings.Join(issues, ", ")})
			} else {
				report.CorrectDetail.Method = append(report.CorrectDetail.Method, domain.MethodDiff{ParentClassName: sol.Name, Sol: &solMethodCopy, Stu: &matchingStu, Description: "Match"})
			}
		} else {
			solMethodCopy := sMethod
			if isCtor {
				report.MissingDetail.Method = append(report.MissingDetail.Method, domain.MethodDiff{ParentClassName: sol.Name, Sol: &solMethodCopy, Stu: nil, Description: "Missing constructor"})
			} else {
				report.MissingDetail.Method = append(report.MissingDetail.Method, domain.MethodDiff{ParentClassName: sol.Name, Sol: &solMethodCopy, Stu: nil, Description: "Missing method (" + sMethod.Scope + " " + strings.Join(sMethod.Outputs, "|") + ")"})
			}
		}
	}

	var stuG, stuS, stuNormal []domain.ProcessedMethod
	for i, stMethod := range stuMethods {
		if !matchedStuMethIdx[i] {
			switch stMethod.Type {
			case "getter":
				stuG = append(stuG, stMethod)
			case "setter":
				stuS = append(stuS, stMethod)
			default:
				stuNormal = append(stuNormal, stMethod)
			}
		}
	}

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

	for i := range stuNormal {
		stMethod := &stuNormal[i]
		report.ExtraDetail.Method = append(report.ExtraDetail.Method, domain.MethodDiff{ParentClassName: stu.Name, Sol: nil, Stu: stMethod, Description: "Extra method (" + stMethod.Scope + " " + stMethod.Output + ")"})
	}
}

// splitSolutionMethods partitions methods into getters, setters, and others.
func (v *StandardMemberComparator) splitSolutionMethods(methods []domain.SolutionProcessedMethod) (g, s, normal []domain.SolutionProcessedMethod) {
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

// splitStudentMethods partitions methods into getters, setters, and others.
func (v *StandardMemberComparator) splitStudentMethods(methods []domain.ProcessedMethod) (g, s, normal []domain.ProcessedMethod) {
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
func (v *StandardMemberComparator) isConstructor(m domain.SolutionProcessedMethod, className string) bool {
	for _, n := range m.Names {
		if strings.EqualFold(n, className) || strings.EqualFold(n, "init") || strings.EqualFold(n, "<<create>>") {
			return true
		}
	}
	return false
}

// isStudentConstructor checks if a method is likely a constructor for a given class.
func (v *StandardMemberComparator) isStudentConstructor(m domain.ProcessedMethod, className string) bool {
	return strings.EqualFold(m.Name, className) || strings.EqualFold(m.Name, "init") || strings.EqualFold(m.Name, "<<create>>")
}

// matchMethodName performs fuzzy matching on method names with special logic for constructors.
func (v *StandardMemberComparator) matchMethodName(sol domain.SolutionProcessedMethod, stu domain.ProcessedMethod, solIsCtor bool, stuClassName string) bool {
	if solIsCtor {
		return v.isStudentConstructor(stu, stuClassName)
	}
	for _, n := range sol.Names {
		if v.fuzzyMatcher.Compare(n, stu.Name) >= 0.5 {
			return true
		}
	}
	return false
}

// itoa is a helper for integer to string conversion.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
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
