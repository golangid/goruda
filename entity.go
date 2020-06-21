package goruda

import (
	"strings"
	"time"
)

type DomainData struct {
	StructName  string
	TimeStamp   time.Time
	Attributes  []Attribute
	Imports     map[string]Import
	Packagename string
	IsPolymorph bool
	SliceData   SliceData
}

type SliceData struct {
	Type string
}

type Import struct {
	Alias string
	Path  string
}

type Attribute struct {
	Name       string
	Type       string
	IsRequired bool
}

func (a Attribute) IsInteger() bool {
	if strings.Contains(strings.ToLower(a.Type), "int") {
		return true
	}
	return false
}

func (a Attribute) GetBitNumber() string {
	if strings.Contains(a.Type, "32") {
		return "32"
	} else if strings.Contains(a.Type, "32") {
		return "64"
	}
	return ""
}

func (a Attribute) IsFloat() bool {
	if strings.Contains(strings.ToLower(a.Type), "float") {
		return true
	}
	return false
}

func (a Attribute) IsCustomType() bool {
	switch a.Type {
	case "string", "int", "int32", "int64", "float", "float32", "float64", "interface{}", "map[string]interface{}":
		return false
	}
	return true
}

type ListOfAttributes struct {
	Attributes  Attributes
	ReturnValue Attributes
}

type Attributes []Attribute

func (a Attributes) GetLastIndex() int {
	if len(a) < 1 {
		return 0
	}
	return len(a) - 1
}

type AbstractionData struct {
	TimeStamp   time.Time
	PackageName string
	Name        string
	Methods     map[string]ListOfAttributes
}

type HTTPMethods struct {
	Path        string
	MethodsName string
	Data        ListOfAttributes
}

type HTTPData struct {
	TimeStamp   time.Time
	PackageName string
	ServiceName string
	Methods     map[string]HTTPMethods
}
