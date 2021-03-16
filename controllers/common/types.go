package common

type Namespace string

func (s Namespace) Str() string {
	return string(s)
}

type Name string

func (s Name) Str() string {
	return string(s)
}

// TODO
//type NamespacedName struct {
//	namespace Namespace
//	name Name
//}


