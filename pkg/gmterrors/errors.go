// Package gmterrors provides error handling for the gorm-multitenancy package.
package gmterrors

// Error is an error that contains a scheme and an error.
type Error struct {
	scheme string
	err    error
}

// Error returns the error message.
func (e *Error) Error() string {
	baseMsg := "gorm-multitenancy"
	if e.scheme != "" {
		baseMsg += "/" + e.scheme
	}
	return baseMsg + ": " + e.err.Error()
}

// Unwrap returns the original error.
func (e *Error) Unwrap() error {
	return e.err
}

// NewWithScheme returns a new error with the given scheme and error.
func NewWithScheme(scheme string, err error) *Error {
	return &Error{scheme: scheme, err: err}
}

// New returns a new error with the given error.
func New(err error) *Error {
	return &Error{err: err}
}
