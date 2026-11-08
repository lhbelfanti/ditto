package verification

import "errors"

var (
	ErrUnexpectedStatus   = errors.New("unexpected status")
	ErrCantSendRequest    = errors.New("can't send the request")
	ErrInvalidRequest     = errors.New("invalid request")
	ErrCantDecodeBody     = errors.New("can't decode the response body")
	ErrRequirementNotMet  = errors.New("requirement not met")
	ErrJourneyPanicked    = errors.New("journey panicked")
	ErrRouteWithoutMethod = errors.New("the route has no method")
	ErrUnknownParam       = errors.New("the route has no placeholder for the param")
	ErrMissingParam       = errors.New("the route has a placeholder without a param")
	ErrCantCreateDir      = errors.New("can't create the evidence directory")
	ErrCantMarshalReport  = errors.New("can't marshal the report")
	ErrCantWriteReport    = errors.New("can't write the report")
)
