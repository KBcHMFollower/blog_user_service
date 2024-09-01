package handlers_dep

type Validator interface {
	Struct(s any) error
}
