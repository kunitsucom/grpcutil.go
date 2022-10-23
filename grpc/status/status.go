package statusz

import (
	"fmt"

	"golang.org/x/xerrors"
	"google.golang.org/grpc/status"
)

// Status is a custom gRPC status with *google.golang.org/grpc/status.Status and error inside.
//
// nolint: errname
type Status struct {
	grpcStatus *status.Status
	err        error
	frame      xerrors.Frame
}

// New returns *github.com/kunitsuinc/grpcutil.go/grpc/status.Status.
func New(grpcStatus *status.Status, err error) *Status {
	return &Status{
		grpcStatus: grpcStatus,
		err:        err,
		frame:      xerrors.Caller(1),
	}
}

func NewWithCallerSkip(grpcStatus *status.Status, err error, skip int) *Status {
	return &Status{
		grpcStatus: grpcStatus,
		err:        err,
		frame:      xerrors.Caller(skip),
	}
}

func (e *Status) GRPCStatus() *status.Status {
	return e.grpcStatus
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
