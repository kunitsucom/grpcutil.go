package statusz

import (
	"fmt"

	"golang.org/x/xerrors"
	"google.golang.org/grpc/codes"
)

// Status
//
// nolint: errname
type Status struct {
	code  codes.Code
	err   error
	frame xerrors.Frame
}

// Error.
func Error(code codes.Code, err error) *Status {
	return &Status{
		code:  code,
		err:   err,
		frame: xerrors.Caller(1),
	}
}

func (e *Status) Error() string {
	return e.err.Error()
}

func (e *Status) Unwrap() error {
	return e.err
}

func (e *Status) Format(s fmt.State, v rune) {
	xerrors.FormatError(e, s, v)
}

func (e *Status) FormatError(p xerrors.Printer) error {
	p.Print(e.Error())
	e.frame.Format(p)
	return e.err
}
