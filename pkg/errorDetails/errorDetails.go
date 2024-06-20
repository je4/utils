package errorDetails

import (
	"context"
	"emperror.dev/emperror"
	"emperror.dev/errors"
	"google.golang.org/appengine/log"
)

var errorCodes = NewDetailFactory()

type ErrorDetails struct {
	ErrorCode    string `json:"errorCode"`
	ErrorMessage string `json:"errorMessage"`
}

func GetErrorStacktrace(err error) errors.StackTrace {
	type stackTracer interface {
		StackTrace() errors.StackTrace
	}

	var stack errors.StackTrace

	errors.UnwrapEach(err, func(err error) bool {
		e := emperror.ExposeStackTrace(err)
		st, ok := e.(stackTracer)
		if !ok {
			return true
		}

		stack = st.StackTrace()
		return true
	})

	if len(stack) > 2 {
		stack = stack[:len(stack)-2]
	}
	return stack
	// fmt.Printf("%+v", st[0:2]) // top two frames
}

func GetDetails(err error) []*ErrorDetails {
	var details []*ErrorDetails
	ds := errors.GetDetails(err)
	for _, d := range ds {
		if e, ok := d.(*ErrorDetails); ok {
			details = append(details, e)
		}
	}
	return details
}

func GetDetailCodes(err error) []string {
	var codes []string
	details := GetDetails(err)
	for _, d := range details {
		codes = append(codes, d.ErrorCode)
	}
	return codes
}

func WithDetail(err error, errorCode string) error {
	detail, ok := GetErrorDetails(errorCode)
	if !ok {
		var _err = errors.Errorf("unknown error code: %s", errorCode)
		log.Debugf(context.Background(), "%v%+v", _err, GetErrorStacktrace(err))
		emperror.Panic(_err)
	}
	return errors.WithDetails(err, detail)
}

func NewErrorDetails(errorCode string, errorMessage string) *ErrorDetails {
	return &ErrorDetails{
		ErrorCode:    errorCode,
		ErrorMessage: errorMessage,
	}
}

func NewDetailFactory() DetailFactory {
	return DetailFactory{
		"0000": "unknown error",
	}
}

type DetailFactory map[string]string

func SetErrorDetails(errorCode string, errorMessage string) {
	errorCodes[errorCode] = errorMessage
}

func GetErrorDetails(errorCode string) (*ErrorDetails, bool) {
	str, ok := errorCodes[errorCode]
	if !ok {
		return nil, false
	}
	return NewErrorDetails(errorCode, str), true
}

func GetErrorDetailsList() []*ErrorDetails {
	var details = []*ErrorDetails{}
	for errorCode, errorMessage := range errorCodes {
		details = append(details, NewErrorDetails(errorCode, errorMessage))
	}
	return details
}
