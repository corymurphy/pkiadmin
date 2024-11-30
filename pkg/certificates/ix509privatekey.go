package certificates

import "unsafe"

type IX509PrivateKeyVtbl struct {
	ICertEnrollVtbl
	Create uintptr
}

type IX509PrivateKey struct {
	ICertEnroll
}

func (v *IX509PrivateKey) VTable() *IX509PrivateKeyVtbl {
	return (*IX509PrivateKeyVtbl)(unsafe.Pointer(v.RawVTable))
}
