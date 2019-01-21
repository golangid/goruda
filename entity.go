package goruda

import "time"

type DomainData struct {
	StructName  string
	TimeStamp   time.Time
	Attributes  []Attribute
	Imports     map[string]Import
	Packagename string
}

type Import struct {
	Alias string
	Path  string
}

type Attribute struct {
	Name string
	Type string
}
