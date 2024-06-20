package errorDetails

import "emperror.dev/errors"

type ErrorDetails struct {
	ErrorCode    int32  `json:"errorCode"`
	ErrorMessage string `json:"errorMessage"`
}

func NewErrorDetails(errorCode int32, errorMessage string) *ErrorDetails {
	return &ErrorDetails{
		ErrorCode:    errorCode,
		ErrorMessage: errorMessage,
	}
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

func NewDetailFactory() DetailFactory {
	return DetailFactory{
		0: "unknown error",
	}
}

type DetailFactory map[int32]string

func (df DetailFactory) SetErrorDetails(errorCode int32, errorMessage string) {
	df[errorCode] = errorMessage
}

func (df DetailFactory) GetErrorDetails(errorCode int32) (*ErrorDetails, bool) {
	str, ok := df[errorCode]
	if !ok {
		return nil, false
	}
	return NewErrorDetails(errorCode, str), true
}

func (df DetailFactory) GetErrorDetailsList() []*ErrorDetails {
	var details = []*ErrorDetails{}
	for errorCode, errorMessage := range df {
		details = append(details, NewErrorDetails(errorCode, errorMessage))
	}
	return details
}

func (df DetailFactory) WithDetail(errorCode int32) error {
	detail, ok := df.GetErrorDetails(errorCode)
	if !ok {
		return nil
	}
	return errors.WithDetails(errors.New(""), detail)
}
