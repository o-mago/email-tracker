package erring

import "errors"

var (
	ErrExpected              = errors.New("expected")
	ErrEventInvalid          = NewAppError("event:invalid", "invalid event")
	ErrRequestInvalid        = NewAppError("request:invalid", "invalid request")
	ErrClientResponseInvalid = NewAppError("client-response:invalid", "client response invalid")
)
