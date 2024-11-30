package certificates

import "errors"

const (
	E_ERROR_BUSY        = 0xAA
	E_ERROR_FILE_EXISTS = 0x50
)

var (
	ErrBusy       = errors.New("The CSP handle is not NULL.")
	ErrFileExists = errors.New("The key already exists.")
)
