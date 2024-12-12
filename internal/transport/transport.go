package transport

import (
	"context"
)

type SendImageParams struct {
	Ctx      context.Context
	Receiver string
	Message  string
	Image    []byte
	MimeType string
}

type Driver interface {
	SendWithImage(arg SendImageParams) error
}
