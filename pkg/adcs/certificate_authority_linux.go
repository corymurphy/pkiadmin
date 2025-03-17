//go:build !windows

package adcs

import (
	"context"
	"encoding/pem"
	"fmt"
	"net"
	"strings"

	"github.com/oiweiwei/go-msrpc/dcerpc"
	"github.com/oiweiwei/go-msrpc/msrpc/dcom"
	"github.com/oiweiwei/go-msrpc/msrpc/dcom/iobjectexporter/v0"
	"github.com/oiweiwei/go-msrpc/msrpc/dcom/iremotescmactivator/v0"
	"github.com/oiweiwei/go-msrpc/msrpc/dcom/iremunknown/v0"
	"github.com/oiweiwei/go-msrpc/msrpc/dcom/iremunknown2/v0"
	"github.com/oiweiwei/go-msrpc/msrpc/dcom/wcce"
	"github.com/oiweiwei/go-msrpc/msrpc/dcom/wcce/icertrequestd/v0"
	"github.com/oiweiwei/go-msrpc/msrpc/dcom/wcce/icertrequestd2/v0"
	"github.com/oiweiwei/go-msrpc/ssp"
	"github.com/oiweiwei/go-msrpc/ssp/credential"
	"github.com/oiweiwei/go-msrpc/ssp/gssapi"
	"golang.org/x/text/encoding/unicode"

	"github.com/oiweiwei/go-msrpc/msrpc/dcom/csra/icertadmind2/v0"
	_ "github.com/oiweiwei/go-msrpc/msrpc/dcom/csra/icertadmind2/v0"
)

type AdcsCertificateAuthority struct {
	Name     string
	Server   string
	Username string
	Password string
	Port     string
}

var (
	// d99e6e74-fc88-11d0-b498-00a0c90312f3
	ActiveDirectoryCertificateServicesClassId = &dcom.ClassID{Data1: 0xd99e6e74, Data2: 0xfc88, Data3: 0x11d0, Data4: []byte{0xb4, 0x98, 0x00, 0xa0, 0xc9, 0x03, 0x12, 0xf3}}

	//d99e6e73-fc88-11d0-b498-00a0c90312f3
	ActiveDirectoryCertificateServicesAdminClassId = &dcom.ClassID{Data1: 0xd99e6e73, Data2: 0xfc88, Data3: 0x11d0, Data4: []byte{0xb4, 0x98, 0x00, 0xa0, 0xc9, 0x03, 0x12, 0xf3}}
	SessionIDAnyLoginSession                       = uint32(0xffffffff)

	CR_PROP_TEMPLATES int32 = 0x0000001D
)

func (ca *AdcsCertificateAuthority) Ping() (err error) {

	ctx := gssapi.NewSecurityContext(context.Background())

	cc, err := dcerpc.Dial(ctx, net.JoinHostPort(ca.Server, "135"),
		dcerpc.WithMechanism(ssp.SPNEGO),
		dcerpc.WithMechanism(ssp.NTLM),
		dcerpc.WithCredentials(credential.NewFromPassword(ca.Username, ca.Password)))
	if err != nil {
		return fmt.Errorf("dial %v", err)
	}
	defer cc.Close(ctx)

	oec, err := iobjectexporter.NewObjectExporterClient(ctx, cc, dcerpc.WithSign())
	if err != nil {
		return fmt.Errorf("new_object_exporter %v", err)
	}

	srv, err := oec.ServerAlive2(ctx, &iobjectexporter.ServerAlive2Request{})
	if err != nil {
		return fmt.Errorf("server_alive2 %v", err)
	}

	irac, err := iremotescmactivator.NewRemoteSCMActivatorClient(ctx, cc, dcerpc.WithSign())
	if err != nil {
		return fmt.Errorf("new remote activator client %v", err)
	}

	clientContext := &dcom.Context{
		MajorVersion: 1,
		MinVersion:   1,
		ContextID:    NewGUID().GUID(),
		Flags:        uint32(dcom.ContextMarshalFlagByValue),
		Frozen:       true,
	}

	ifdClientContext, err := clientContext.InterfacePointer()
	if err != nil {
		return fmt.Errorf("error creating client context interface pointer %v", err)
	}

	activationProps := &dcom.ActivationProperties{
		DestinationContext: 2,
		Properties: []dcom.ActivationProperty{
			&dcom.InstantiationInfoData{
				ClassID:          ActiveDirectoryCertificateServicesClassId,
				ClassContext:     0x00100015,
				IID:              []*dcom.IID{icertrequestd.CertRequestDIID},
				ClientCOMVersion: srv.COMVersion,
			},
			&dcom.SCMRequestInfoData{
				RemoteRequest: &dcom.CustomRemoteRequestSCMInfo{
					ClientImpLevel:             0x00000002,
					RequestedProtocolSequences: []uint16{7},
				},
			},
			&dcom.SecurityInfoData{
				ServerInfo: &dcom.COMServerInfo{
					Name: ca.Server,
				},
			},
			&dcom.SpecialPropertiesData{
				SessionID:         SessionIDAnyLoginSession,
				DefaultAuthnLevel: uint32(dcerpc.AuthLevelPktIntegrity),
				Flags:             uint32(dcom.SPDFlagsUseDefaultAuthnLevel), // undocumented flag used by certutil?
				OrigClassContext:  0x00100015,
			},
			&dcom.ActivationContextInfoData{
				IfdClientContext: ifdClientContext,
			},
			&dcom.LocationInfoData{},
		},
	}

	actPropertiesIn, err := activationProps.ActivationPropertiesIn()
	if err != nil {
		return fmt.Errorf("error activation properties in: %v", err)
	}

	this := &dcom.ORPCThis{Version: srv.COMVersion, CID: NewGUID()}

	req := &iremotescmactivator.RemoteCreateInstanceRequest{
		ORPCThis:        this,
		ActPropertiesIn: actPropertiesIn,
	}

	rac, err := irac.RemoteCreateInstance(ctx, req)
	if err != nil {
		return fmt.Errorf("remote activation instance error: %v", err)
	}

	if err = activationProps.Parse(rac.ActPropertiesOut); err != nil {
		return fmt.Errorf("rerror parsing activation properties in: %v", err)
	}

	scm, propertiesOut := activationProps.SCMReplyInfoData(), activationProps.PropertiesOutInfo()
	if scm == nil || propertiesOut == nil {
		return fmt.Errorf("no scmReply / propertiesOut info data: %v", err)
	}

	ctx = gssapi.NewSecurityContext(ctx)

	wcc, err := dcerpc.Dial(ctx, ca.Server,
		dcerpc.WithMechanism(ssp.SPNEGO),
		dcerpc.WithMechanism(ssp.NTLM),
		dcerpc.WithCredentials(credential.NewFromPassword(ca.Username, ca.Password)),
		dcerpc.WithSeal(),
		scm.RemoteReply.OXIDBindings.EndpointsByProtocol("ncacn_ip_tcp")[1], // TODO: get the IP one
	)
	if err != nil {
		return fmt.Errorf("dial_wmi_endpoint: %v", err)
	}
	defer wcc.Close(ctx)

	cli, err := iremunknown2.NewRemoteUnknown2Client(ctx, wcc, dcerpc.WithSeal())
	if err != nil {
		return fmt.Errorf("error creating client %v", err)

	}

	rqi, err := cli.RemoteUnknown().RemoteQueryInterface(ctx, &iremunknown.RemoteQueryInterfaceRequest{
		This:            this,
		IPID:            propertiesOut.InterfaceData[0].IPID().GUID(),
		ReferencesCount: 5,
		IIDs:            []*dcom.IID{icertrequestd2.CertRequestD2IID},
	}, dcom.WithIPID(scm.RemoteReply.IPIDRemoteUnknown))
	if err != nil {
		return fmt.Errorf("error remote query interface 2 %v", err)
	}

	certreq, err := icertrequestd2.NewCertRequestD2Client(ctx, wcc, dcom.WithIPID(rqi.QueryInterfaceResults[0].Std.IPID))
	if err != nil {
		return fmt.Errorf("error creating cert request d2 client %v", err)
	}

	pingres, err := certreq.Ping2(ctx, &icertrequestd2.Ping2Request{This: this, Authority: ca.Name})
	if err != nil || pingres.Return != 0 {
		return fmt.Errorf("error pinging cert request d2 %v", err)
	}

	return err
}

func (ca *AdcsCertificateAuthority) Request(ctx context.Context, request CertificateSigningRequest) (response CertificateSigningResponse, err error) {

	ctx = gssapi.NewSecurityContext(ctx)

	cc, err := dcerpc.Dial(ctx, net.JoinHostPort(ca.Server, ca.Port),
		dcerpc.WithMechanism(ssp.SPNEGO),
		dcerpc.WithMechanism(ssp.NTLM),
		dcerpc.WithCredentials(credential.NewFromPassword(ca.Username, ca.Password)))
	if err != nil {
		return response, fmt.Errorf("dial %v", err)
	}
	defer cc.Close(ctx)

	oec, err := iobjectexporter.NewObjectExporterClient(ctx, cc, dcerpc.WithSign())
	if err != nil {
		return response, fmt.Errorf("new_object_exporter %v", err)
	}

	srv, err := oec.ServerAlive2(ctx, &iobjectexporter.ServerAlive2Request{})
	if err != nil {
		return response, fmt.Errorf("server_alive2 %v", err)
	}

	irac, err := iremotescmactivator.NewRemoteSCMActivatorClient(ctx, cc, dcerpc.WithSign())
	if err != nil {
		return response, fmt.Errorf("new remote activator client %v", err)
	}

	clientContext := &dcom.Context{
		MajorVersion: 1,
		MinVersion:   1,
		ContextID:    NewGUID().GUID(),
		Flags:        uint32(dcom.ContextMarshalFlagByValue),
		Frozen:       true,
	}

	ifdClientContext, err := clientContext.InterfacePointer()
	if err != nil {
		return response, fmt.Errorf("error creating client context interface pointer %v", err)
	}

	activationProps := &dcom.ActivationProperties{
		DestinationContext: 2,
		Properties: []dcom.ActivationProperty{
			&dcom.InstantiationInfoData{
				ClassID:          ActiveDirectoryCertificateServicesClassId,
				ClassContext:     0x00100015,
				IID:              []*dcom.IID{icertrequestd.CertRequestDIID},
				ClientCOMVersion: srv.COMVersion,
			},
			&dcom.SCMRequestInfoData{
				RemoteRequest: &dcom.CustomRemoteRequestSCMInfo{
					ClientImpLevel:             0x00000002,
					RequestedProtocolSequences: []uint16{7},
				},
			},
			&dcom.SecurityInfoData{
				ServerInfo: &dcom.COMServerInfo{
					Name: ca.Server,
				},
			},
			&dcom.SpecialPropertiesData{
				SessionID:         SessionIDAnyLoginSession,
				DefaultAuthnLevel: uint32(dcerpc.AuthLevelPktIntegrity),
				Flags:             uint32(dcom.SPDFlagsUseDefaultAuthnLevel), // undocumented flag used by certutil?
				OrigClassContext:  0x00100015,
			},
			&dcom.ActivationContextInfoData{
				IfdClientContext: ifdClientContext,
			},
			&dcom.LocationInfoData{},
		},
	}

	actPropertiesIn, err := activationProps.ActivationPropertiesIn()
	if err != nil {
		return response, fmt.Errorf("error activation properties in: %v", err)
	}

	this := &dcom.ORPCThis{Version: srv.COMVersion, CID: NewGUID()}

	req := &iremotescmactivator.RemoteCreateInstanceRequest{
		ORPCThis:        this,
		ActPropertiesIn: actPropertiesIn,
	}

	rac, err := irac.RemoteCreateInstance(ctx, req)
	if err != nil {
		return response, fmt.Errorf("remote activation instance error: %v", err)
	}

	if err = activationProps.Parse(rac.ActPropertiesOut); err != nil {
		return response, fmt.Errorf("error parsing activation properties in: %v", err)
	}

	scm, propertiesOut := activationProps.SCMReplyInfoData(), activationProps.PropertiesOutInfo()
	if scm == nil || propertiesOut == nil {
		return response, fmt.Errorf("no scmReply / propertiesOut info data: %v", err)
	}

	ctx = gssapi.NewSecurityContext(ctx)

	wcc, err := dcerpc.Dial(ctx, ca.Server,
		dcerpc.WithMechanism(ssp.SPNEGO),
		dcerpc.WithMechanism(ssp.NTLM),
		dcerpc.WithCredentials(credential.NewFromPassword(ca.Username, ca.Password)),
		dcerpc.WithSeal(),
		scm.RemoteReply.OXIDBindings.EndpointsByProtocol("ncacn_ip_tcp")[1], // TODO: get the IP one
	)
	if err != nil {
		return response, fmt.Errorf("dial_wmi_endpoint: %v", err)
	}
	defer wcc.Close(ctx)

	cli, err := iremunknown2.NewRemoteUnknown2Client(ctx, wcc, dcerpc.WithSeal())
	if err != nil {
		return response, fmt.Errorf("error creating client %v", err)
	}

	rqi, err := cli.RemoteUnknown().RemoteQueryInterface(ctx, &iremunknown.RemoteQueryInterfaceRequest{
		This:            this,
		IPID:            propertiesOut.InterfaceData[0].IPID().GUID(),
		ReferencesCount: 5,
		IIDs:            []*dcom.IID{icertrequestd2.CertRequestD2IID},
	}, dcom.WithIPID(scm.RemoteReply.IPIDRemoteUnknown))
	if err != nil {
		return response, fmt.Errorf("error remote query interface 2 %v", err)
	}

	certreq, err := icertrequestd2.NewCertRequestD2Client(ctx, wcc, dcom.WithIPID(rqi.QueryInterfaceResults[0].Std.IPID))
	if err != nil {
		return response, fmt.Errorf("error creating cert request d2 client %v", err)
	}

	gpr, err := certreq.GetCAProperty(ctx, &icertrequestd2.GetCAPropertyRequest{
		This:          this,
		Authority:     ca.Name,
		PropertyID:    0x0000001D,
		PropertyType:  0x00000004,
		PropertyIndex: 0,
	})
	if err != nil {
		return response, fmt.Errorf("error getting ca property %v", err)
	}

	_, err = DecodeCertTransportBlob(gpr.PropertyValue.Buffer)
	if err != nil {
		return response, fmt.Errorf("error decoding property value %v", err)
	}

	signed, err := certreq.CertRequestD().Request(ctx, &icertrequestd.RequestRequest{
		This:       this,
		Authority:  ca.Name,
		Attributes: request.Attributes,
		Request:    &wcce.CertTransportBlob{Buffer: request.Csr, Length: uint32(len(request.Csr))},
	})
	if err != nil {
		return response, fmt.Errorf("error requesting cert %v", err)
	}
	dispositionMessage, err := DecodeCertTransportBlob(signed.DispositionMessage.Buffer)
	if err != nil {
		return response, fmt.Errorf("error decoding disposition message %v", err)
	}
	if signed.Return != 0 {
		return response, fmt.Errorf(
			"error requesting cert: %v %d %s",
			signed.Return, signed.Disposition, dispositionMessage)
	}

	certBlock := &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: signed.EncodedCert.Buffer,
	}
	certPEM := pem.EncodeToMemory(certBlock)

	response.Certificate = certPEM
	response.Return = signed.Return
	response.Disposition = Disposition(signed.Disposition)
	response.DispositionMessage = dispositionMessage

	return response, err
}

func (ca *AdcsCertificateAuthority) Templates(ctx context.Context) (templates []CertificateAuthorityTemplate, err error) {

	ctx = gssapi.NewSecurityContext(ctx)

	cc, err := dcerpc.Dial(ctx, net.JoinHostPort(ca.Server, ca.Port),
		dcerpc.WithMechanism(ssp.SPNEGO),
		dcerpc.WithMechanism(ssp.NTLM),
		dcerpc.WithCredentials(credential.NewFromPassword(ca.Username, ca.Password)))
	if err != nil {
		return templates, fmt.Errorf("dial %v", err)
	}
	defer cc.Close(ctx)

	oec, err := iobjectexporter.NewObjectExporterClient(ctx, cc, dcerpc.WithSign())
	if err != nil {
		return templates, fmt.Errorf("new_object_exporter %v", err)
	}

	srv, err := oec.ServerAlive2(ctx, &iobjectexporter.ServerAlive2Request{})
	if err != nil {
		return templates, fmt.Errorf("server_alive2 %v", err)
	}

	irac, err := iremotescmactivator.NewRemoteSCMActivatorClient(ctx, cc, dcerpc.WithSign())
	if err != nil {
		return templates, fmt.Errorf("new remote activator client %v", err)
	}

	clientContext := &dcom.Context{
		MajorVersion: 1,
		MinVersion:   1,
		ContextID:    NewGUID().GUID(),
		Flags:        uint32(dcom.ContextMarshalFlagByValue),
		Frozen:       true,
	}

	ifdClientContext, err := clientContext.InterfacePointer()
	if err != nil {
		return templates, fmt.Errorf("error creating client context interface pointer %v", err)
	}

	activationProps := &dcom.ActivationProperties{
		DestinationContext: 2,
		Properties: []dcom.ActivationProperty{
			&dcom.InstantiationInfoData{
				ClassID:          ActiveDirectoryCertificateServicesAdminClassId,
				ClassContext:     0x00100015,
				IID:              []*dcom.IID{icertadmind2.CertAdminD2IID},
				ClientCOMVersion: srv.COMVersion,
			},
			&dcom.SCMRequestInfoData{
				RemoteRequest: &dcom.CustomRemoteRequestSCMInfo{
					ClientImpLevel:             0x00000002,
					RequestedProtocolSequences: []uint16{7},
				},
			},
			&dcom.SecurityInfoData{
				ServerInfo: &dcom.COMServerInfo{
					Name: ca.Server,
				},
			},
			&dcom.SpecialPropertiesData{
				SessionID:         SessionIDAnyLoginSession,
				DefaultAuthnLevel: uint32(dcerpc.AuthLevelPktIntegrity),
				Flags:             uint32(dcom.SPDFlagsUseDefaultAuthnLevel), // undocumented flag used by certutil?
				OrigClassContext:  0x00100015,
			},
			&dcom.ActivationContextInfoData{
				IfdClientContext: ifdClientContext,
			},
			&dcom.LocationInfoData{},
		},
	}

	actPropertiesIn, err := activationProps.ActivationPropertiesIn()
	if err != nil {
		return templates, fmt.Errorf("error activation properties in: %v", err)
	}

	this := &dcom.ORPCThis{Version: srv.COMVersion, CID: NewGUID()}

	req := &iremotescmactivator.RemoteCreateInstanceRequest{
		ORPCThis:        this,
		ActPropertiesIn: actPropertiesIn,
	}

	rac, err := irac.RemoteCreateInstance(ctx, req)
	if err != nil {
		return templates, fmt.Errorf("remote activation instance error: %v", err)
	}

	if err = activationProps.Parse(rac.ActPropertiesOut); err != nil {
		return templates, fmt.Errorf("error parsing activation properties in: %v", err)
	}

	scm, propertiesOut := activationProps.SCMReplyInfoData(), activationProps.PropertiesOutInfo()
	if scm == nil || propertiesOut == nil {
		return templates, fmt.Errorf("no scmReply / propertiesOut info data: %v", err)
	}

	ctx = gssapi.NewSecurityContext(ctx)

	wcc, err := dcerpc.Dial(ctx, ca.Server,
		dcerpc.WithMechanism(ssp.SPNEGO),
		dcerpc.WithMechanism(ssp.NTLM),
		dcerpc.WithCredentials(credential.NewFromPassword(ca.Username, ca.Password)),
		dcerpc.WithSeal(),
		scm.RemoteReply.OXIDBindings.EndpointsByProtocol("ncacn_ip_tcp")[1], // TODO: get the IP one
	)
	if err != nil {
		return templates, fmt.Errorf("dial_wmi_endpoint: %v", err)
	}
	defer wcc.Close(ctx)

	cli, err := iremunknown2.NewRemoteUnknown2Client(ctx, wcc, dcerpc.WithSeal())
	if err != nil {
		return templates, fmt.Errorf("error creating client %v", err)
	}

	rqi, err := cli.RemoteUnknown().RemoteQueryInterface(ctx, &iremunknown.RemoteQueryInterfaceRequest{
		This:            this,
		IPID:            propertiesOut.InterfaceData[0].IPID().GUID(),
		ReferencesCount: 5,
		IIDs:            []*dcom.IID{icertadmind2.CertAdminD2IID},
	}, dcom.WithIPID(scm.RemoteReply.IPIDRemoteUnknown))
	if err != nil {
		return templates, fmt.Errorf("error remote query interface 2 %v", err)
	}

	certadm, err := icertadmind2.NewCertAdminD2Client(ctx, wcc, dcom.WithIPID(rqi.QueryInterfaceResults[0].Std.IPID))
	if err != nil {
		return templates, fmt.Errorf("error creating cert admin d2 client %v", err)
	}

	props, err := certadm.GetCAProperty(ctx, &icertadmind2.GetCAPropertyRequest{
		This:          this,
		Authority:     ca.Name,
		PropertyID:    CR_PROP_TEMPLATES,
		PropertyType:  0x00000004,
		PropertyIndex: 0,
	})
	if err != nil {
		return templates, fmt.Errorf("error getting ca property %v", err)
	}

	decoder := unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM).NewDecoder()
	serialized, err := decoder.Bytes(props.PropertyValue.Buffer)
	if err != nil {
		return templates, fmt.Errorf("error decoding property value %v", err)
	}

	// TODO this needs to be more defensive
	for i, item := range strings.Split(string(serialized), "\n") {
		if item == "\x00" {
			continue
		}

		if i%2 == 0 {
			templates = append(templates, CertificateAuthorityTemplate{Name: item})
		} else {
			templates[len(templates)-1].ID = item
		}
	}

	return templates, err
}
