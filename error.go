package fuzzy

import "errors"

var (
	ErrMissingArguments      = errors.New("missing arguments")
	ErrUndefinedVariable     = errors.New("undefined variable")
	ErrValueNotFound         = errors.New("value not found")
	ErrUndefinedTerm         = errors.New("undefined term")
	ErrVariableAlreadyExists = errors.New("variable already exists")
	ErrTermAlreadyExists     = errors.New("term already exists")
)
