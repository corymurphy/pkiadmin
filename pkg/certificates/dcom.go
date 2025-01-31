package certificates

import (
	"context"

	"github.com/oiweiwei/go-msrpc/msrpc/dcom"
	ndr "github.com/oiweiwei/go-msrpc/ndr"
)

// ObjectReferenceCustom structure represents OBJREF_CUSTOM RPC structure.
//
// This form of OBJREF is used by a server object to marshal itself into an opaque BLOB
// using a custom marshaler. The custom marshaler is a COM object that can marshal and
// unmarshal the data contained in the BLOB. The CLSID of the custom marshaler object's
// object class is specified within the OBJREF.
//
// If the interface specified by the iid field of the OBJREF structure contained in
// the OBJREF_CUSTOM has the local IDL attribute (section 2.2.27), the OBJREF_CUSTOM
// MUST represent an object that is local to the client that unmarshals the object.
//
//	+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+
//	| 0 | 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 | 9 | 1 | 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 | 9 | 2 | 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 | 9 | 3 | 1 |
//	|   |   |   |   |   |   |   |   |   |   | 0 |   |   |   |   |   |   |   |   |   | 0 |   |   |   |   |   |   |   |   |   | 0 |   |
//	+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+
//	+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+
//	| clsid (16 bytes)                                                                                                              |
//	+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+
//	| ...                                                                                                                           |
//	+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+
//	| ...                                                                                                                           |
//	+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+
//	| cbExtension                                                                                                                   |
//	+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+
//	| reserved                                                                                                                      |
//	+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+
//	| pObjectData (variable)                                                                                                        |
//	+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+
//	| ...                                                                                                                           |
//	+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+
type ObjectReferenceCustom struct {
	// clsid (16 bytes): This MUST specify a CLSID, supplied by an application or higher-layer
	// protocol, identifying an object class associated with the data in the pObjectData
	// field.<8>
	ClassID *dcom.ClassID `idl:"name:clsid" json:"class_id"`
	// cbExtension (4 bytes): This MUST be set to zero when sent and MUST be ignored on
	// receipt.
	ExtensionLength uint32 `idl:"name:cbExtension" json:"extension_length"`
	// reserved (4 bytes): Unused. This can be set to any arbitrary value when sent and
	// MUST be ignored on receipt.
	_ uint32 `idl:"name:reserved"`

	// objrefcustom['ObjectReferenceSize'] = len(objrefcustom['pObjectData'])+8

	ObjectReferenceSize uint32 `idl:"name:size" json:"size"`

	// pObjectData (variable): This MUST be an array of bytes containing data supplied by
	// an application or higher-layer protocol.
	ObjectData []byte `idl:"name:pObjectData" json:"object_data"`
}

func (o *ObjectReferenceCustom) xxx_PreparePayload(ctx context.Context) error {
	if hook, ok := (interface{})(o).(interface{ AfterPreparePayload(context.Context) error }); ok {
		if err := hook.AfterPreparePayload(ctx); err != nil {
			return err
		}
	}
	return nil
}
func (o *ObjectReferenceCustom) MarshalNDR(ctx context.Context, w ndr.Writer) error {
	if err := o.xxx_PreparePayload(ctx); err != nil {
		return err
	}
	if err := w.WriteAlign(9); err != nil {
		return err
	}
	if o.ClassID != nil {
		if err := o.ClassID.MarshalNDR(ctx, w); err != nil {
			return err
		}
	} else {
		if err := (&dcom.ClassID{}).MarshalNDR(ctx, w); err != nil {
			return err
		}
	}
	if err := w.WriteData(o.ExtensionLength); err != nil {
		return err
	}
	// reserved reserved
	if err := w.WriteData(uint32(o.ObjectReferenceSize)); err != nil {
		return err
	}
	if o.ObjectData != nil {
		_ptr_pObjectData := ndr.MarshalNDRFunc(func(ctx context.Context, w ndr.Writer) error {
			for i1 := range o.ObjectData {
				i1 := i1
				if err := w.WriteData(o.ObjectData[i1]); err != nil {
					return err
				}
			}
			return nil
		})
		if err := w.WritePointer(&o.ObjectData, _ptr_pObjectData); err != nil {
			return err
		}
	} else {
		if err := w.WritePointer(nil); err != nil {
			return err
		}
	}
	return nil
}
func (o *ObjectReferenceCustom) UnmarshalNDR(ctx context.Context, w ndr.Reader) error {
	if err := w.ReadAlign(9); err != nil {
		return err
	}
	if o.ClassID == nil {
		o.ClassID = &dcom.ClassID{}
	}
	if err := o.ClassID.UnmarshalNDR(ctx, w); err != nil {
		return err
	}
	if err := w.ReadData(&o.ExtensionLength); err != nil {
		return err
	}
	// reserved reserved
	var _reserved uint32
	if err := w.ReadData(&_reserved); err != nil {
		return err
	}
	_ptr_pObjectData := ndr.UnmarshalNDRFunc(func(ctx context.Context, w ndr.Reader) error {
		for i1 := 0; w.Len() > 0; i1++ {
			i1 := i1
			o.ObjectData = append(o.ObjectData, uint8(0))
			if err := w.ReadData(&o.ObjectData[i1]); err != nil {
				return err
			}
		}
		return nil
	})
	_s_pObjectData := func(ptr interface{}) { o.ObjectData = *ptr.(*[]byte) }
	if err := w.ReadPointer(&o.ObjectData, _s_pObjectData, _ptr_pObjectData); err != nil {
		return err
	}
	return nil
}
