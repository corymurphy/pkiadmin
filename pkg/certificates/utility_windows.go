//go:build windows

package certificates

import (
	"strings"

	"github.com/go-ole/go-ole"
)

func convertHresultToError(hr uintptr) (err error) {
	if hr != 0 {
		err = ole.NewError(hr)
		if strings.Contains(err.Error(), "FormatMessage failed") {
			switch hr {
			case E_ERROR_BUSY:
				err = ErrBusy
			case E_ERROR_FILE_EXISTS:
				err = ErrFileExists
			}
		}
	}
	return
}
