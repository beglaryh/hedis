package redis_clone

type valueElement struct {
	v  any
	et elementType
}

type elementType int

const (
	ESTRING elementType = iota + 1
	ELIST
)
