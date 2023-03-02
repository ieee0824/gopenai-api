package api

import "errors"

var ErrUnauthorized = errors.New("Unauthorized")
var ErrUnknown = errors.New("Unkonown")

type Error struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Param   any    `json:"param"`
	Code    string `json:"code"`
}
