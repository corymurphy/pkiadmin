package adcs

import (
	"crypto/rand"
	"encoding/binary"

	"github.com/oiweiwei/go-msrpc/msrpc/dcom"
	"golang.org/x/text/encoding/unicode"
)

func NewGUID() *dcom.CID {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}

	// Version 4 UUID - Set version bits
	b[6] = (b[6] & 0x0f) | 0x40 // Version 4
	b[8] = (b[8] & 0x3f) | 0x80 // Variant RFC4122

	return &dcom.CID{
		Data1: binary.LittleEndian.Uint32(b[0:4]),
		Data2: binary.LittleEndian.Uint16(b[4:6]),
		Data3: binary.LittleEndian.Uint16(b[6:8]),
		Data4: b[8:16], // byte array stays as-is
	}
}

func DecodeCertTransportBlob(data []byte) (string, error) {
	decoder := unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM).NewDecoder()
	bytes := make([]byte, len(data)*2)
	for i, b := range data {
		binary.LittleEndian.PutUint16(bytes[i*2:], uint16(b))
	}
	decoded, err := decoder.String(string(bytes))
	if err != nil {
		return "", err
	}
	return decoded, nil
}
