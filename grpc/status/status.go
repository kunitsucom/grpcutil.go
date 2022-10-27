package statusz

import (
	"fmt"

	"github.com/golang/protobuf/proto" //nolint:staticcheck
	"golang.org/x/xerrors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Status is a custom gRPC status with *google.golang.org/grpc/status.Status and error inside.
//
// nolint: errname
type Status struct {
	*status.Status
	err   error
	frame xerrors.Frame
}

// New returns *github.com/kunitsuinc/grpcutil.go/grpc/status.Status.
func New(c codes.Code, msg string, err error) *Status {
	return &Status{
		Status: status.New(c, msg),
		err:    err,
		frame:  xerrors.Caller(1),
	}
}

// New returns *github.com/kunitsuinc/grpcutil.go/grpc/status.Status with caller skip.
func NewWithCallerSkip(c codes.Code, msg string, err error, skip int) *Status {
	return &Status{
		Status: status.New(c, msg),
		err:    err,
		frame:  xerrors.Caller(skip),
	}
}

func (sz *Status) GRPCStatus() *status.Status {
	return sz.Status
}

func (sz *Status) WithDetails(details ...proto.Message) (*Status, error) {
	st, err := sz.Status.WithDetails(details...)
	if err != nil {
		return nil, fmt.Errorf("(*status.Status).WithDetails: %w", err)
	}

	sz.Status = st

	return sz, nil
}

func (sz *Status) Error() string {
	return sz.err.Error()
}

func (sz *Status) Unwrap() error {
	return sz.err
}

func (sz *Status) Format(s fmt.State, v rune) {
	xerrors.FormatError(sz, s, v)
}

func (sz *Status) FormatError(p xerrors.Printer) error {
	p.Print(sz.Error())
	sz.frame.Format(p)
	return sz.err
}
