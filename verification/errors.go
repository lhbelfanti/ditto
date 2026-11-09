package verification

import "errors"

var (
	ErrUnexpectedStatus   = errors.New("unexpected status")
	ErrCantSendRequest    = errors.New("can't send the request")
	ErrCantDecodeBody     = errors.New("can't decode the response body")
	ErrUnexpectedData     = errors.New("unexpected data")
	ErrCheckFailed        = errors.New("check failed")
	ErrUnexpectedResponse = errors.New("unexpected response")
	ErrJourneyPanicked    = errors.New("journey panicked")
	ErrCantCreateDir      = errors.New("can't create the evidence directory")
	ErrCantMarshalReport  = errors.New("can't marshal the report")
	ErrCantWriteReport    = errors.New("can't write the report")
	ErrCantListCards      = errors.New("can't list the feature cards")
	ErrCantReadCard       = errors.New("can't read the feature card")
	ErrCantWriteCard      = errors.New("can't write the feature card")
	ErrCardNotFound       = errors.New("no feature card has a table for the journey")
)
