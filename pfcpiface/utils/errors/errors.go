package errors

import (
	"errors"
	"fmt"
)

var (
	errNotFound         = errors.New("not found")
	errInvalidArgument  = errors.New("invalid argument")
	errInvalidOperation = errors.New("invalid operation")
	errFailed           = errors.New("failed")
	//errUnsupported      = errors.New("unsupported")
)

type PFCPErr struct {
	Message string
	Value   interface{}
	Err     error
}

type NotFoundParamErr struct {
	Message        string
	ParameterName  string
	ParameterValue interface{}
	Err            error
}

func (err *NotFoundParamErr) Error() string {
	if err.Err != nil {
		return err.Err.Error()
	}

	return err.Message
}

func (err *PFCPErr) Error() string {
	if err.Err != nil {
		return err.Err.Error()
	}

	return err.Message
}

func ErrUnsupported(message string, value interface{}) *PFCPErr {
	msg := fmt.Sprintf("Unsupported Error; Msg: %s. Value: %v", message, value)
	return &PFCPErr{
		Message: msg,
		Value:   value,
	}
}

func ErrNotFound(message string) error {
	msg := fmt.Sprintf("NotFound Error: %s", message)
	return &PFCPErr{
		Message: msg,
	}
}

func ErrNotFoundWithParam(message string, paramName string, paramValue interface{}) error {
	msg := fmt.Sprintf("NotFound Error with parameters: %s ", message)
	return &NotFoundParamErr{
		Message:        msg,
		ParameterName:  paramName,
		ParameterValue: paramValue,
	}

	//return fmt.Errorf("%s %w with %s=%v", what, errNotFound, paramName, paramValue)
}

func ErrInvalidOperation(operation interface{}) error {
	return fmt.Errorf("%w: %v", errInvalidOperation, operation)
}

func ErrInvalidArgument(name string, value interface{}) error {
	return fmt.Errorf("%w '%s': %v", errInvalidArgument, name, value)
}

func ErrInvalidArgumentWithReason(name string, value interface{}, reason string) error {
	return fmt.Errorf("%w '%s'=%v (%s)", errInvalidArgument, name, value, reason)
}

func ErrOperationFailed(operation interface{}) error {
	return fmt.Errorf("%v %w", operation, errFailed)
}

func ErrOperationFailedWithReason(operation interface{}, reason string) error {
	return fmt.Errorf("%v %w due to: : %s", operation, errFailed, reason)
}

func ErrOperationFailedWithParam(operation interface{}, paramName string, paramValue interface{}) error {
	return fmt.Errorf("'%v' %w for %s=%v", operation, errFailed, paramName, paramValue)
}
