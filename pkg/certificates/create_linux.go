//go:build !windows

package certificates

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
	"os"

	"github.com/oiweiwei/go-msrpc/dcerpc"
	"github.com/oiweiwei/go-msrpc/msrpc/dcom"
	"github.com/oiweiwei/go-msrpc/msrpc/dcom/iobjectexporter/v0"
	"github.com/oiweiwei/go-msrpc/msrpc/dcom/iremunknown/v0"
	"github.com/oiweiwei/go-msrpc/msrpc/dcom/iremunknown2/v0"
	"github.com/oiweiwei/go-msrpc/ndr"

	"github.com/oiweiwei/go-msrpc/msrpc/dcom/iremotescmactivator/v0"
	wcce "github.com/oiweiwei/go-msrpc/msrpc/dcom/wcce"
	wccec "github.com/oiweiwei/go-msrpc/msrpc/dcom/wcce/client"
	"github.com/oiweiwei/go-msrpc/msrpc/dcom/wcce/icertrequestd/v0"
	"github.com/oiweiwei/go-msrpc/msrpc/dcom/wcce/icertrequestd2/v0"
	_ "github.com/oiweiwei/go-msrpc/msrpc/epm/epm/v3"
	"github.com/oiweiwei/go-msrpc/msrpc/erref/hresult"
	_ "github.com/oiweiwei/go-msrpc/msrpc/erref/ntstatus"
	"github.com/oiweiwei/go-msrpc/ssp/gssapi"
	"github.com/rs/zerolog"

	"github.com/joho/godotenv"
	"github.com/oiweiwei/go-msrpc/dcerpc/errors"

	"github.com/oiweiwei/go-msrpc/ssp"
	"github.com/oiweiwei/go-msrpc/ssp/credential"
)

func init() {

	godotenv.Load(".env")

	cred := credential.NewFromPassword(os.Getenv("USERNAME"), os.Getenv("PASSWORD"))

	fmt.Println("---------", os.Getenv("SERVER"), cred.DomainName(), cred.UserName(), "---------")
	gssapi.AddCredential(credential.NewFromPassword(os.Getenv("USERNAME"), os.Getenv("PASSWORD")))
	gssapi.AddMechanism(ssp.SPNEGO)
	gssapi.AddMechanism(ssp.NTLM)
	gssapi.AddMechanism(ssp.KRB5)

	errors.AddMapper(hresult.Mapper{})
}

var j = func(data any) string { b, _ := json.MarshalIndent(data, "", "  "); return string(b) }

var (
	_ = wcce.GoPackage
	_ = wccec.GoPackage
	_ = icertrequestd2.GoPackage
	_ = iremotescmactivator.GoPackage
	// _ = iremunknown2v0.GoPackage
	_ = iremunknown2.GoPackage
)

var (

	// d99e6e70-fc88-11d0-b498-00a0c90312f3

	// d99e6e74-fc88-11d0-b498-00a0c90312f3
	ActiveDirectoryCertificateServicesClassId = &dcom.ClassID{Data1: 0xd99e6e74, Data2: 0xfc88, Data3: 0x11d0, Data4: []byte{0xb4, 0x98, 0x00, 0xa0, 0xc9, 0x03, 0x12, 0xf3}}

	// 000001A2-0000-0000-C000-000000000046
	IActivationPropertiesInIID = &dcom.IID{Data1: 0x1a2, Data2: 0x00, Data3: 0x00, Data4: []byte{0xc0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

	// 00000338-0000-0000-c000-000000000046
	IActivationPropertiesInClassID = &dcom.ClassID{Data1: 0x338, Data2: 0x00, Data3: 0x00, Data4: []byte{0xc0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

	// 000001aa-0000-0000-c000-000000000046
	ScmRequestInfoClassID = &dcom.ClassID{Data1: 0x1aa, Data2: 0x00, Data3: 0x00, Data4: []byte{0xc0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

	// 000001a4-0000-0000-c000-000000000046
	ServerLocationinfoClassID = &dcom.ClassID{Data1: 0x1a4, Data2: 0x00, Data3: 0x00, Data4: []byte{0xc0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

	// 000001a6-0000-0000-c000-000000000046
	SecurityInfoClassID = &dcom.ClassID{Data1: 0x1a6, Data2: 0x00, Data3: 0x00, Data4: []byte{0xc0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

	// 000001b9-0000-0000-c000-000000000046
	SpecialSystemPropertiesClassID = &dcom.ClassID{Data1: 0x1b9, Data2: 0x00, Data3: 0x00, Data4: []byte{0xc0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

	// 000001ab-0000-0000-c000-000000000046
	InstantiationInfoClassID = &dcom.ClassID{Data1: 0x1ab, Data2: 0x00, Data3: 0x00, Data4: []byte{0xc0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

	// 000001a5-0000-0000-c000-000000000046
	ActivationContextInfoClassID = &dcom.ClassID{Data1: 0x1a5, Data2: 0x00, Data3: 0x00, Data4: []byte{0xc0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

	// 000001c0-0000-0000-C000-000000000046
	IContextIID = &dcom.IID{Data1: 0x1c0, Data2: 0x00, Data3: 0x00, Data4: []byte{0xc0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

	// 0000033b-0000-0000-c000-000000000046 CLSID_ContextMarshaler
	ContextMarshalerClassID = &dcom.ClassID{Data1: 0x33b, Data2: 0x00, Data3: 0x00, Data4: []byte{0xc0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}

	ZeroClassId = &dcom.ClassID{Data1: 0x00, Data2: 0x00, Data3: 0x00, Data4: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}}
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

type Marshable interface {
	ndr.Marshaler
	ndr.Unmarshaler
}

type ActivationProperty struct {
	Property Marshable
	ClassId  *dcom.ClassID
}

func NewActivationProperty(classId *dcom.ClassID, property Marshable) ActivationProperty {
	return ActivationProperty{
		Property: property,
		ClassId:  classId,
	}
}

func CertRequestPing() (result interface{}) {

	ctx := gssapi.NewSecurityContext(context.Background())

	logger := zerolog.New(os.Stdout).Level(zerolog.Disabled)

	cc, err := dcerpc.Dial(ctx, net.JoinHostPort(os.Getenv("SERVER"), "135"), dcerpc.WithLogger(logger))
	if err != nil {
		fmt.Fprintln(os.Stderr, err, "dial")
		return
	}
	defer cc.Close(ctx)

	oec, err := iobjectexporter.NewObjectExporterClient(ctx, cc, dcerpc.WithSign())
	if err != nil {
		fmt.Fprintln(os.Stderr, "new_object_exporter", err)
		return
	}

	srv, err := oec.ServerAlive2(ctx, &iobjectexporter.ServerAlive2Request{})
	if err != nil {
		fmt.Fprintln(os.Stderr, "server_alive2", err)
		return
	}

	cc, err = dcerpc.Dial(ctx, net.JoinHostPort(os.Getenv("SERVER"), "135"), dcerpc.WithLogger(logger))
	if err != nil {
		fmt.Fprintln(os.Stderr, err, "dial")
		return
	}
	defer cc.Close(ctx)

	irac, err := iremotescmactivator.NewRemoteSCMActivatorClient(ctx, cc, dcerpc.WithSign())
	if err != nil {
		fmt.Fprintln(os.Stderr, err, "new remote activator client")
		return
	}

	serialProperties := []byte{}

	clientContext := &dcom.ObjectReference{
		Signature: []byte{0x4d, 0x45, 0x4f, 0x57},
		Flags:     0x00000004,
		IID:       IContextIID,
		ObjectReference: &dcom.ObjectReference_ObjectReference{
			Value: &dcom.ObjectReference_Custom{
				Custom: &dcom.ObjectReferenceCustom{
					ClassID:             ContextMarshalerClassID,
					ObjectReferenceSize: uint32(len(make([]byte, 48))),
					ObjectData: []byte{
						0x01, 0x00, 0x01, 0x00, 0xc0, 0x96, 0x84, 0x26, 0xbe, 0x77, 0x8b, 0x4c, 0xac, 0x48, 0xfb, 0x05,
						0x46, 0x69, 0x88, 0xe7, 0x02, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
						0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00,
					},
				},
			},
		},
	}

	clientContextData, err := ndr.Marshal(clientContext, ndr.Opaque)

	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	properties := []ActivationProperty{
		NewActivationProperty(SpecialSystemPropertiesClassID,
			&dcom.SpecialPropertiesData{
				SessionID:         0xffffffff,
				DefaultAuthnLevel: 0x00000005,
				// DefaultAuthnLevel: 0x00000004,
				Flags:            0x00000002,
				OrigClassContext: 0x00100015,
			}),
		NewActivationProperty(InstantiationInfoClassID, &dcom.InstantiationInfoData{
			ClassID:          ActiveDirectoryCertificateServicesClassId,
			ClassContext:     0x00100015,
			IIDCount:         1,
			IID:              []*dcom.IID{icertrequestd.CertRequestDIID},
			ThisSize:         88,
			ClientCOMVersion: srv.COMVersion,
		}),
		NewActivationProperty(ActivationContextInfoClassID, &dcom.ActivationContextInfoData{
			IfdClientContext: &dcom.InterfacePointer{
				Data: clientContextData,
			},
		}),
		NewActivationProperty(SecurityInfoClassID, &dcom.SecurityInfoData{
			ServerInfo: &dcom.COMServerInfo{
				Name: os.Getenv("SERVER"),
			},
		}),
		NewActivationProperty(ServerLocationinfoClassID, &dcom.LocationInfoData{}),
		NewActivationProperty(ScmRequestInfoClassID, &dcom.SCMRequestInfoData{
			RemoteRequest: &dcom.CustomRemoteRequestSCMInfo{
				ClientImpLevel:                  0x00000002,
				RequestedProtocolSequencesCount: 1,
				RequestedProtocolSequences:      []uint16{7},
			},
		}),
	}

	header := &dcom.CustomHeader{
		InterfacesCount:    6,
		TotalSize:          uint32(len(serialProperties)),
		DestinationContext: 2,
		ClassIDs:           []*dcom.ClassID{},
	}

	// header

	for _, property := range properties {

		propSerial, err := ndr.MarshalWithTypeSerializationV1(property.Property)

		if err != nil {
			fmt.Println(err)
		}

		pad := (8 - len(propSerial)%8) % 8

		for i := 0; i < pad; i++ {
			propSerial = append(propSerial, 0x00)
		}

		header.Sizes = append(header.Sizes, uint32(len(propSerial)))
		header.ClassIDs = append(header.ClassIDs, property.ClassId)

		serialProperties = append(serialProperties, propSerial...)

	}

	headerSerial, err := ndr.MarshalWithTypeSerializationV1(header)

	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	pad := (8 - len(headerSerial)%8) % 8
	for i := 0; i < pad; i++ {
		headerSerial = append(headerSerial, 0x00)
	}

	header.HeaderSize = uint32(len(headerSerial))
	headerSerial, err = ndr.MarshalWithTypeSerializationV1(header)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	pad = (8 - len(headerSerial)%8) % 8
	for i := 0; i < pad; i++ {
		headerSerial = append(headerSerial, 0x00)
	}

	blob := []byte{}
	size := make([]byte, 4)
	binary.LittleEndian.PutUint32(size, uint32(len(serialProperties)+len(headerSerial)))

	blob = append(blob, size...)
	blob = append(blob, []byte{0x00, 0x00, 0x00, 0x00}...) // dwReserved
	blob = append(blob, headerSerial...)
	blob = append(blob, serialProperties...)

	or := &dcom.ObjectReference{
		Signature: []byte{0x4d, 0x45, 0x4f, 0x57},
		Flags:     0x00000004,
		IID:       IActivationPropertiesInIID,
		ObjectReference: &dcom.ObjectReference_ObjectReference{
			Value: &dcom.ObjectReference_Custom{
				Custom: &dcom.ObjectReferenceCustom{
					ClassID:             IActivationPropertiesInClassID,
					ObjectReferenceSize: uint32(len(blob) + 8),
					ObjectData:          blob,
				},
			},
		},
	}

	data, err := ndr.Marshal(or, ndr.Opaque)

	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	actPropertiesIn := &dcom.InterfacePointer{
		Data: data,
	}

	actPropertiesIn.MarshalNDR(ctx, ndr.NDR20(nil, ndr.Opaque))

	cid := NewGUID()

	req := &iremotescmactivator.RemoteCreateInstanceRequest{
		ORPCThis:        &dcom.ORPCThis{Version: srv.COMVersion, CID: cid},
		ActPropertiesIn: actPropertiesIn,
	}

	rac, err := irac.RemoteCreateInstance(ctx, req)

	if err != nil {
		fmt.Println("error remote activation instance", err)
		return
	}

	resp := rac.ActPropertiesOut.GetCustomObjectReference()

	header = &dcom.CustomHeader{}
	err = ndr.UnmarshalWithTypeSerializationV1(resp.ObjectData[8:], header)
	if err != nil {
		fmt.Println("error unmarshalling header", err)
		return
	}

	propertiesOutSize := header.Sizes[0]
	propertiesData := resp.ObjectData[header.HeaderSize+8 : header.HeaderSize+8+propertiesOutSize]
	propertiesOut := &dcom.PropertiesOutInfo{}
	err = ndr.UnmarshalWithTypeSerializationV1(propertiesData, propertiesOut)
	if err != nil {
		fmt.Println("error unmarshalling properties out", err)
		return
	}

	scmData := resp.ObjectData[header.HeaderSize+8+propertiesOutSize:]
	scm := &dcom.SCMReplyInfoData{}
	err = ndr.UnmarshalWithTypeSerializationV1(scmData, scm)
	if err != nil {
		fmt.Println("error unmarshalling scm reply", err)
		return
	}

	ctx = gssapi.NewSecurityContext(ctx)

	wcc, err := dcerpc.Dial(ctx, os.Getenv("SERVER"),
		dcerpc.WithLogger(logger),
		dcerpc.WithSeal(),
		scm.RemoteReply.OXIDBindings.EndpointsByProtocol("ncacn_ip_tcp")[1], // TODO: get the IP one
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, "dial_wmi_endpoint", err)
		return
	}
	defer wcc.Close(ctx)

	cli, err := iremunknown2.NewRemoteUnknown2Client(ctx, wcc, dcerpc.WithSeal())

	if err != nil {
		fmt.Println("error creating client", err)
		return
	}

	rqi, err := cli.RemoteUnknown().RemoteQueryInterface(ctx, &iremunknown.RemoteQueryInterfaceRequest{
		This:            &dcom.ORPCThis{Version: srv.COMVersion, Flags: 0x0, CID: cid},
		IPID:            propertiesOut.InterfaceData[0].IPID().GUID(),
		ReferencesCount: 5,
		IIDsCount:       1,
		IIDs:            []*dcom.IID{icertrequestd2.CertRequestD2IID},
	}, dcom.WithIPID(scm.RemoteReply.IPIDRemoteUnknown))
	if err != nil {
		fmt.Println("error remote query interface 2", err)
		return
	}

	certreq, err := icertrequestd2.NewCertRequestD2Client(ctx, wcc, dcom.WithIPID(rqi.QueryInterfaceResults[0].Std.IPID))

	if err != nil {
		fmt.Println("error creating cert request d2 client", err)
		return
	}

	pingres, err := certreq.Ping2(ctx, &icertrequestd2.Ping2Request{
		This: &dcom.ORPCThis{Version: srv.COMVersion, CID: cid},
	})

	if err != nil {
		fmt.Println("error pinging cert request d2", err)
		return
	}

	fmt.Println("ping response", pingres.Return)

	return ""
}
