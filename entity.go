package goruda

import "time"

type DomainData struct {
	StructName  string
	TimeStamp   time.Time
	Attributes  []Attribute
	Imports     map[string]Import
	Packagename string
}

func (d DomainData) IsStructPolymorph() bool {
	for _, attribute := range d.Attributes {
		if attribute.Name == "" {
			return true
		}
	}
	return false
}

type Import struct {
	Alias string
	Path  string
}

type Attribute struct {
	Name string
	Type string
}
