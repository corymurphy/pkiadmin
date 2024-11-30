package certificates

import (
	"unsafe"

	"github.com/go-ole/go-ole"
)

type ICertEnrollVtbl struct {
	ole.IDispatchVtbl
}

type ICertEnroll struct {
	ole.IDispatch
}

func (v *ICertEnroll) VTable() *ICertEnrollVtbl {
	return (*ICertEnrollVtbl)(unsafe.Pointer(v.RawVTable))
}
