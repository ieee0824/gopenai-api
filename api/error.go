package api

import (
	"golang.org/x/xerrors"
)

var ErrUnauthorized = xerrors.New("Unauthorized")
var ErrUnknown = xerrors.New("Unkonown")
var ErrStatusBadGateway = xerrors.New("Bad Gateway")

type Error struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Param   any    `json:"param"`
	Code    string `json:"code"`
}
