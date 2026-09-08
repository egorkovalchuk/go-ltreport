package influx

// Методы для типа ScenarioDinamic
func (p *ScenarioDinamic) SetApplication(NameTest string) {
	p.NameTest = NameTest
}
func (p *ScenarioDinamic) SetThread(NameThread string) {
	p.NameThread = NameThread
}

func (p *ScenarioDinamic) SeField(YF []YField) {
	p.Field = append(p.Field, YF...)
}

func AddMap(m map[string]map[string]KeyField, TestName, ErrorField string, val KeyField) map[string]map[string]KeyField {
	mm, ok := m[TestName]
	if !ok {
		mm = make(map[string]KeyField)
		m[TestName] = mm
	}
	mm[ErrorField] = val

	return m
}

func AddMapS(m map[string]map[string]ScenarioDinamic, TestName, NameThread string, val ScenarioDinamic) map[string]map[string]ScenarioDinamic {
	mm, ok := m[TestName]
	if !ok {
		mm = make(map[string]ScenarioDinamic)
		m[TestName] = mm
	}
	mm[NameThread] = val

	return m
}
