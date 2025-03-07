//go:build !windows

package certificates

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"net"
	"os"

	"github.com/oiweiwei/go-msrpc/dcerpc"
	"github.com/oiweiwei/go-msrpc/msrpc/dcom"
	"github.com/oiweiwei/go-msrpc/msrpc/dcom/iobjectexporter/v0"
	"github.com/oiweiwei/go-msrpc/msrpc/dcom/iremunknown/v0"
	"github.com/oiweiwei/go-msrpc/msrpc/dcom/iremunknown2/v0"

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

	_ "github.com/corymurphy/pkiadmin/pkg/dcom"
)

func init() {

	godotenv.Load(".env")

	// cred := credential.NewFromPassword(os.Getenv("USERNAME"), os.Getenv("PASSWORD"))

	// fmt.Println("---------", os.Getenv("SERVER"), cred.DomainName(), cred.UserName(), "---------")
	gssapi.AddCredential(credential.NewFromPassword(os.Getenv("USERNAME"), os.Getenv("PASSWORD")))
	gssapi.AddMechanism(ssp.SPNEGO)
	gssapi.AddMechanism(ssp.NTLM)
	gssapi.AddMechanism(ssp.KRB5)

	errors.AddMapper(hresult.Mapper{})
}

var (
	_ = wcce.GoPackage
	_ = wccec.GoPackage
	_ = icertrequestd2.GoPackage
	_ = iremotescmactivator.GoPackage
	_ = iremunknown2.GoPackage
)

var (
	// d99e6e74-fc88-11d0-b498-00a0c90312f3
	ActiveDirectoryCertificateServicesClassId = &dcom.ClassID{Data1: 0xd99e6e74, Data2: 0xfc88, Data3: 0x11d0, Data4: []byte{0xb4, 0x98, 0x00, 0xa0, 0xc9, 0x03, 0x12, 0xf3}}

	SessionIDAnyLoginSession = uint32(0xffffffff)
)

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

	irac, err := iremotescmactivator.NewRemoteSCMActivatorClient(ctx, cc, dcerpc.WithSign())
	if err != nil {
		fmt.Fprintln(os.Stderr, err, "new remote activator client")
		return
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
		fmt.Println("error creating client context interface pointer", err)
		return
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
					Name: os.Getenv("SERVER"),
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
		fmt.Println("error activation properties in", err)
		return
	}

	this := &dcom.ORPCThis{Version: srv.COMVersion, CID: NewGUID()}

	req := &iremotescmactivator.RemoteCreateInstanceRequest{
		ORPCThis:        this,
		ActPropertiesIn: actPropertiesIn,
	}

	rac, err := irac.RemoteCreateInstance(ctx, req)
	if err != nil {
		fmt.Println("error remote activation instance", err)
		return
	}

	if err := activationProps.Parse(rac.ActPropertiesOut); err != nil {
		fmt.Println("error parsing activation properties in", err)
		return
	}

	scm, propertiesOut := activationProps.SCMReplyInfoData(), activationProps.PropertiesOutInfo()
	if scm == nil || propertiesOut == nil {
		fmt.Println("no scmReply / propertiesOut info data")
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
		This:            this,
		IPID:            propertiesOut.InterfaceData[0].IPID().GUID(),
		ReferencesCount: 5,
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

	pingres, err := certreq.Ping2(ctx, &icertrequestd2.Ping2Request{This: this, Authority: os.Getenv("CANAME")})
	if err != nil || pingres.Return != 0 {
		fmt.Println("error pinging cert request d2", err)
		return
	}

	gpr, err := certreq.GetCAProperty(ctx, &icertrequestd2.GetCAPropertyRequest{
		This:          this,
		Authority:     os.Getenv("CANAME"),
		PropertyID:    0x0000001D,
		PropertyType:  0x00000004,
		PropertyIndex: 0,
	})
	if err != nil {
		fmt.Println("error getting ca property", err)
		return
	}

	_, err = DecodeCertTransportBlob(gpr.PropertyValue.Buffer)
	if err != nil {
		fmt.Println("error decoding property value", err)
		return
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		fmt.Println("error generating private key", err)
		return err
	}

	id := make([]byte, 16)
	if _, err := rand.Read(id); err != nil {
		fmt.Println("error generating random hex:", err)
		return err
	}
	id = []byte(fmt.Sprintf("%x", id))

	privateKeyPem := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey)})
	if err := os.WriteFile(fmt.Sprintf(".data/%s.key", id), privateKeyPem, 0600); err != nil {
		fmt.Println("error writing private key file:", err)
		return err
	}

	template := &x509.CertificateRequest{
		Subject: pkix.Name{
			CommonName:   "web.lab2.internal",
			Organization: []string{"CAM"},
		},
		DNSNames: []string{"ad.lab2.internal", "dev.lab2.internal"},
	}

	csr, err := x509.CreateCertificateRequest(rand.Reader, template, privateKey)
	if err != nil {
		fmt.Println("error creating certificate request:", err)
		return err
	}

	csrBlock := &pem.Block{
		Type:  "CERTIFICATE REQUEST",
		Bytes: csr,
	}

	csrPem := pem.EncodeToMemory(csrBlock)

	signed, err := certreq.CertRequestD().Request(ctx, &icertrequestd.RequestRequest{
		This:       this,
		Authority:  os.Getenv("CANAME"),
		Attributes: "CertificateTemplate:ServerAuthentication-CngRsa",
		Request:    &wcce.CertTransportBlob{Buffer: csr, Length: uint32(len(csrPem))},
	})
	if err != nil {
		fmt.Println("error requesting cert", err)
		return
	}

	certBlock := &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: signed.EncodedCert.Buffer,
	}
	certPEM := pem.EncodeToMemory(certBlock)

	if err := os.WriteFile(fmt.Sprintf(".data/%s.crt", id), certPEM, 0600); err != nil {
		fmt.Println("error writing signed certifcate file:", err)
		return err
	}

	return "success"
}
