package goruda

import (
	"time"
)

type DomainData struct {
	StructName  string
	TimeStamp   time.Time
	Attributes  []Attribute
	Imports     map[string]Import
	Packagename string
	IsPolymorph bool
}

type Import struct {
	Alias string
	Path  string
}

type Attribute struct {
	Name string
	Type string
}

type ListOfAttributes struct {
	Attributes  Attributes
	ReturnValue Attributes
}

type Attributes []Attribute

func (l Attributes) GetLastIndex() int {
	if len(l) < 1 {
		return 0
	}
	return len(l) - 1
}

type AbstractionData struct {
	TimeStamp   time.Time
	PackageName string
	Name        string
	Methods     map[string]ListOfAttributes
}
