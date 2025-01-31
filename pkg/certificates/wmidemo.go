package certificates

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"

	"github.com/oiweiwei/go-msrpc/dcerpc"
	"github.com/rs/zerolog"

	"github.com/oiweiwei/go-msrpc/msrpc/dcom"
	"github.com/oiweiwei/go-msrpc/msrpc/dcom/iactivation/v0"
	"github.com/oiweiwei/go-msrpc/msrpc/dcom/iobjectexporter/v0"
	"github.com/oiweiwei/go-msrpc/msrpc/dcom/wmio/query"

	"github.com/oiweiwei/go-msrpc/msrpc/dcom/wmio"

	"github.com/oiweiwei/go-msrpc/msrpc/dcom/wmi"
	"github.com/oiweiwei/go-msrpc/msrpc/dcom/wmi/iwbemlevel1login/v0"
	"github.com/oiweiwei/go-msrpc/msrpc/dcom/wmi/iwbemservices/v0"

	"github.com/oiweiwei/go-msrpc/ssp/gssapi"

	"github.com/oiweiwei/go-msrpc/msrpc/erref/hresult"
	_ "github.com/oiweiwei/go-msrpc/msrpc/erref/win32"
	_ "github.com/oiweiwei/go-msrpc/msrpc/erref/wmi"
)

var (
	class    string
	method   string
	arg      string
	resource string
	debug    bool
	args     wmio.Values
)

func WmiDemo2() (result interface{}) {

	class = "Win32_Process"
	method = "Create"
	arg = `{"CommandLine":"calc.exe","CurrentDirectory":"C:\\"}`
	resource = "root\\cimv2"
	debug = false

	log := zerolog.New(os.Stderr)

	if !debug {
		log = zerolog.New(io.Discard)
	}

	log.Info().Str("class", class).Str("method", method).Str("args", fmt.Sprintf("%+v", args)).Msg("execute")

	ctx := gssapi.NewSecurityContext(context.Background())

	// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-dcom/2be2642e-67a1-4690-883b-642b505ddb1d

	// ObjectExporter uses well-known endpoint 135.
	cc, err := dcerpc.Dial(ctx, net.JoinHostPort(os.Getenv("SERVER"), "135"))
	if err != nil {
		fmt.Fprintln(os.Stderr, "dial_well_known_endpoint", err)
		return
	}

	defer cc.Close(ctx)

	// new object exporter client.
	cli, err := iobjectexporter.NewObjectExporterClient(ctx, cc, dcerpc.WithSign(), dcerpc.WithTargetName(os.Getenv("TARGET")))
	if err != nil {
		fmt.Fprintln(os.Stderr, "new_object_exporter", err)
		return
	}

	// server-alive to determine the bindings.
	srv, err := cli.ServerAlive2(ctx, &iobjectexporter.ServerAlive2Request{})
	if err != nil {
		fmt.Fprintln(os.Stderr, "server_alive2", err)
		return
	}

	if debug {
		fmt.Println("-----------------------")
		fmt.Println("OBJECT RESOLVER STRING BINDINGS")
		fmt.Println("-----------------------")

		fmt.Println(j(srv.ObjectResolverBindings.GetStringBindings()))

		fmt.Println("-----------------------")
		fmt.Println("OBJECT RESOLVER SECURITY BINDINGS")
		fmt.Println("-----------------------")

		fmt.Println(j(srv.ObjectResolverBindings.GetSecurityBindings()))
	}

	// dial activation client.
	// cc, err = dcerpc.Dial(ctx, net.JoinHostPort(os.Getenv("SERVER"), "135"))
	// if err != nil {
	//	fmt.Fprintln(os.Stderr, "dial_well_known_endpoint_2", err)
	//	return
	// }

	// new activation-client.
	iact, err := iactivation.NewActivationClient(ctx, cc, dcerpc.WithSign(), dcerpc.WithTargetName(os.Getenv("TARGET")))
	if err != nil {
		fmt.Fprintln(os.Stderr, "new_activation_client", err)
		return
	}

	// activate the WMI interface.
	act, err := iact.RemoteActivation(ctx, &iactivation.RemoteActivationRequest{
		ORPCThis: &dcom.ORPCThis{Version: srv.COMVersion},
		ClassID:  wmi.Level1LoginClassID.GUID(),
		IIDs:     []*dcom.IID{iwbemlevel1login.Level1LoginIID},
		// for TCP/IP it must be []uint16{7} / for named pipes: []uint16{15}.
		RequestedProtocolSequences: []uint16{7, 15},
	})

	if err != nil {
		fmt.Fprintln(os.Stderr, "remote_activation", err)
		return
	}

	if act.HResult != 0 {
		fmt.Fprintln(os.Stderr, hresult.FromCode(uint32(act.HResult)))
		return
	}

	if debug {
		fmt.Println("-----------------------")
		fmt.Println("REMOTE ACTIVATION")
		fmt.Println("-----------------------")

		fmt.Println(j(act))

		for i, ifd := range act.InterfaceData {
			fmt.Println("-----------------------")
			fmt.Println(i, "REMOTE ACTIVATION INTERFACE")
			fmt.Println("-----------------------")
			fmt.Println(j(ifd.GetObjectReference()))
		}
	}

	std := act.InterfaceData[0].GetStandardObjectReference().Std

	// dial WMI using OXID bindings. (use ncacn_tcp_ip).
	wcc, err := dcerpc.Dial(ctx, os.Getenv("SERVER"), act.OXIDBindings.EndpointsByProtocol("ncacn_ip_tcp")...)
	if err != nil {
		fmt.Fprintln(os.Stderr, "dial_wmi_endpoint", err)
		return
	}

	defer wcc.Close(ctx)

	// establish context that will be shared between NewLevel1LoginClient and
	// NewServicesClient.
	ctx = gssapi.NewSecurityContext(ctx)

	l1login, err := iwbemlevel1login.NewLevel1LoginClient(ctx, wcc,
		dcom.WithIPID(std.IPID),
		dcerpc.WithSign(),
		dcerpc.WithTargetName(os.Getenv("TARGET")), dcerpc.WithLogger(log))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	pos, err := l1login.EstablishPosition(ctx, &iwbemlevel1login.EstablishPositionRequest{
		This: &dcom.ORPCThis{Version: srv.COMVersion},
	})

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	if debug {
		fmt.Println(j(pos))
	}

	// login to WMI.
	login, err := l1login.NTLMLogin(ctx, &iwbemlevel1login.NTLMLoginRequest{
		This:            &dcom.ORPCThis{Version: srv.COMVersion},
		NetworkResource: resource,
	})

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	if debug {
		fmt.Println(j(login))
	}

	ns := login.Namespace

	// start services client.
	svcs, err := iwbemservices.NewServicesClient(ctx, wcc, dcom.WithIPID(ns.InterfacePointer().IPID()))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	builder := query.NewBuilder(ctx, svcs, srv.COMVersion)

	if debug {
		obj, _ := builder.Spawn(class).Method(method).Values(args, wmio.JSONValueToType).Object()
		fmt.Println(j(obj))
	}

	// use simple query builder to execute the Create static method of the Win32_Process.
	out, err := builder.Spawn(class).Method(method).Values(args, wmio.JSONValueToType).Exec().Object()
	if err != nil {
		fmt.Fprintln(os.Stderr, "svcs_exec_method", err)
		return
	}

	fmt.Println(j(out.Values()))
	return nil
}
