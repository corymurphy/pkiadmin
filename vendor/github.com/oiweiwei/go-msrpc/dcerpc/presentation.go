package dcerpc

import (
	"sync/atomic"

	"github.com/oiweiwei/go-msrpc/ndr"
)

// The Presentation represents the data presentation
// context.
type Presentation struct {
	// The context identifier.
	id uint16
	// The list of abstract syntaxes. (only one supported).
	AbstractSyntax *SyntaxID
	// Selected transfer syntax.
	TransferSyntax *SyntaxID
	// Error.
	Error error
}

// TransferEncoding function returns the transfer encoding for the
// presentation context. The only supported encoding is NDR v2.0.
func (c *Presentation) TransferEncoding() func([]byte, ...any) ndr.NDR {
	return ndr.NDR20
}

func (c *transport) PresentationFromContextList(ps []*Presentation, results []*Result) *Feature {

	for i := 0; i < len(ps) && i < len(results); i++ {

		// if ps[i].TransferSyntax != nil {
		// the presentation has already been negotiated.
		// continue
		//}

		if results[i].HasError() {
			ps[i].Error = results[i]
			continue
		}

		ps[i].TransferSyntax, ps[i].Error = results[i].TransferSyntax, nil
	}

	if !c.IsBinded() && len(results) > len(ps) {
		if feature := results[len(results)-1]; feature.DefResult == NegotiateAck {
			// bind negotiate features.
			return (*Feature)(feature)
		}
	}

	return nil
}

func (c *transport) PresentationsToContextList(ps []*Presentation, transferSyntaxes []*SyntaxID) []*Context {

	ret := []*Context{}

	for _, p := range ps {

		// if p.TransferSyntax != nil {
		// this context has already been negotiated.
		//	continue
		//}

		// fmt.Println("abstract syntax", p.AbstractSyntax, p.id)

		id := p.id

		// RemoteUnknown2SyntaxUUID = &uuid.UUID{TimeLow: 0x143, TimeMid: 0x0, TimeHiAndVersion: 0x0, ClockSeqHiAndReserved: 0xc0, ClockSeqLow: 0x0, Node: [6]uint8{0x0, 0x0, 0x0, 0x0, 0x0, 0x46}}
		// Syntax ID
		// RemoteUnknown2SyntaxV0_0 = &dcerpc.SyntaxID{IfUUID: RemoteUnknown2SyntaxUUID, IfVersionMajor: 0, IfVersionMinor: 0}

		// CAMHACK
		// if p.AbstractSyntax.IfUUID.TimeLow == 0x143 {
		// 	id = 0
		// }

		// if p.AbstractSyntax.IfUUID.TimeLow == 0x99fcfec4 {
		// 	id = 0
		// }

		ret = append(ret, &Context{
			ContextID:        id,
			AbstractSyntax:   p.AbstractSyntax,
			TransferSyntaxes: transferSyntaxes,
		})
	}

	// CAMHACK
	if !c.IsBinded() && len(ret) > 0 {
		ret = append(ret, &Context{
			AbstractSyntax:   ret[len(ret)-1].AbstractSyntax,
			TransferSyntaxes: []*SyntaxID{BindFeatureSyntaxV1_0},
		})
	}

	return ret
}

var pContextID = new(atomic.Uint32)

func NewPresentationContextID() uint16 {
	return uint16(pContextID.Add(1))
}

func NewPresentation(abstractSyntax *SyntaxID) *Presentation {
	return &Presentation{id: NewPresentationContextID(), AbstractSyntax: abstractSyntax}
}

func (p *Presentation) ID() uint16 {
	return p.id
}
