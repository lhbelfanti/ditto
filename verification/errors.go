package verification

import "errors"

var (
	ErrUnexpectedStatus  = errors.New("unexpected status")
	ErrCantSendRequest   = errors.New("can't send the request")
	ErrCantDecodeBody    = errors.New("can't decode the response body")
	ErrRequirementNotMet = errors.New("requirement not met")
	ErrJourneyPanicked   = errors.New("journey panicked")
	ErrCantCreateDir     = errors.New("can't create the evidence directory")
	ErrCantMarshalReport = errors.New("can't marshal the report")
	ErrCantWriteReport   = errors.New("can't write the report")
)
