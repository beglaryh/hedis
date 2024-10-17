package store

import (
	"strings"

	"github.com/beglaryh/gocommon/collection/list/linkedlist"
)

type valueElement struct {
	v  any
	et elementType
}

type elementType int

const (
	ESTRING elementType = iota + 1
	ELIST
)

func (v valueElement) GetValue() any {
	return v.v
}

func (t elementType) IsString() bool {
	return t == ESTRING
}

func (t elementType) IsList() bool {
	return t == ELIST
}

func (v valueElement) String() string {
	if v.et.IsString() {
		return v.v.(string)
	}
	l := v.v.(linkedlist.LinkedList[string])
	var sb strings.Builder
	count := 0
	sb.WriteString("[")
	l.Stream().ForEach(func(s string) {
		sb.WriteString(s)
		if count != l.Size()-1 {
			sb.WriteString(",")
		}
		count += 1
	})

	sb.WriteString("]")
	return sb.String()
}
