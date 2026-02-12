package persona

import "sasswall/internal/classify"

type Selector struct {
	defaultPersona string
	dynamic        bool
	mapping        map[string]string
}

func New(defaultPersona string, dynamic bool, mapping map[string]string) *Selector {
	if defaultPersona == "" {
		defaultPersona = "sister"
	}
	m := map[string]string{}
	for k, v := range mapping {
		m[k] = v
	}
	return &Selector{defaultPersona: defaultPersona, dynamic: dynamic, mapping: m}
}

func (s *Selector) Pick(category classify.Category) string {
	if !s.dynamic {
		return s.defaultPersona
	}
	if v, ok := s.mapping[string(category)]; ok && v != "" {
		return v
	}
	return s.defaultPersona
}
