//go:build windows

package certificates

import (
	"fmt"

	"github.com/corymurphy/pkiadmin/pkg/adcs"
	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
	"golang.org/x/sys/windows"
)

var (
	user32DLL                                         = windows.NewLazyDLL("user32.dll")
	certEnrollDll                                     = windows.NewLazyDLL("CertEnroll.dll")
	procSystemParamInfo                               = user32DLL.NewProc("SystemParametersInfoW")
	ContextMachine                                    = 2
	INSTALL_RESPONSE_RESTRICITON_ALLOW_UNTRUSTED_ROOT = 0x4
	XCN_CRYPT_STRING_BASE64                           = 0x1
	X509_INSTALL_PASSWORD                             = ""
)

// typedef enum EncodingType {
// 	XCN_CRYPT_STRING_BASE64HEADER = 0,
// 	XCN_CRYPT_STRING_BASE64 = 0x1,
// 	XCN_CRYPT_STRING_BINARY = 0x2,
// 	XCN_CRYPT_STRING_BASE64REQUESTHEADER = 0x3,
// 	XCN_CRYPT_STRING_HEX = 0x4,
// 	XCN_CRYPT_STRING_HEXASCII = 0x5,
// 	XCN_CRYPT_STRING_BASE64_ANY = 0x6,
// 	XCN_CRYPT_STRING_ANY = 0x7,
// 	XCN_CRYPT_STRING_HEX_ANY = 0x8,
// 	XCN_CRYPT_STRING_BASE64X509CRLHEADER = 0x9,
// 	XCN_CRYPT_STRING_HEXADDR = 0xa,
// 	XCN_CRYPT_STRING_HEXASCIIADDR = 0xb,
// 	XCN_CRYPT_STRING_HEXRAW = 0xc,
// 	XCN_CRYPT_STRING_BASE64URI = 0xd,
// 	XCN_CRYPT_STRING_ENCODEMASK = 0xff,
// 	XCN_CRYPT_STRING_CHAIN = 0x100,
// 	XCN_CRYPT_STRING_TEXT = 0x200,
// 	XCN_CRYPT_STRING_PERCENTESCAPE = 0x8000000,
// 	XCN_CRYPT_STRING_HASHDATA = 0x10000000,
// 	XCN_CRYPT_STRING_STRICT = 0x20000000,
// 	XCN_CRYPT_STRING_NOCRLF = 0x40000000,
// 	XCN_CRYPT_STRING_NOCR = 0x80000000
//   } ;

// typedef enum InstallResponseRestrictionFlags {
// 	AllowNone = 0,
// 	AllowNoOutstandingRequest = 0x1,
// 	AllowUntrustedCertificate = 0x2,
// 	AllowUntrustedRoot = 0x4
// } ;

func IssueCertificate() interface{} {
	var err error
	var response string
	var result *ole.VARIANT

	err = ole.CoInitialize(0)
	if err != nil {
		return fmt.Sprintf("error CoInitialize result 1: %v", err)
	}
	defer ole.CoUninitialize()

	oleutil.CreateObject("sql")

	upkcs, err := oleutil.CreateObject("X509Enrollment.CX509CertificateRequestPkcs10")
	defer upkcs.Release()

	if err != nil {
		return fmt.Sprintf("error result 2: %v", err)
	}

	pkcs, err := upkcs.QueryInterface(ole.IID_IDispatch)

	if err != nil {
		return fmt.Sprintf("error result 3: %v", err)
	}

	upk, err := oleutil.CreateObject("X509Enrollment.CX509PrivateKey")
	defer upk.Release()

	if err != nil {
		return fmt.Sprintf("error result 4: %v", err)
	}

	pk, err := upk.QueryInterface(ole.IID_IDispatch)

	if err != nil {
		return fmt.Sprintf("error result 5: %v", err)
	}

	result, err = oleutil.PutProperty(pk, "MachineContext", true)
	response += fmt.Sprintf("<p>pk machine context result %s</p>", result.ToString())

	if err != nil {
		return fmt.Sprintf("error result 6: %v", err)
	}

	result, err = oleutil.PutProperty(pk, "Length", 4096)
	response += fmt.Sprintf("<p>pk length result %s</p>", result.ToString())

	if err != nil {
		return fmt.Sprintf("error result 6: %v", err)
	}

	result, err = oleutil.CallMethod(pk, "Create")
	response += fmt.Sprintf("<p>pk create result %s</p>", result.ToString())

	if err != nil {
		return fmt.Sprintf("error result 7: %v", err)
	}

	udn, err := oleutil.CreateObject("X509Enrollment.CX500DistinguishedName")

	if err != nil {
		return fmt.Sprintf("error result 8: %v", err)
	}

	dn, err := udn.QueryInterface(ole.IID_IDispatch)

	if err != nil {
		return fmt.Sprintf("error result 9: %v", err)
	}

	result, err = oleutil.CallMethod(dn, "Encode", "CN=certmgr-dev.lab2.internal", 0)
	response += fmt.Sprintf("<p>dn encode result %s</p>", result.ToString())

	if err != nil {
		return fmt.Sprintf("error result 10: %v", err)
	}

	result, err = oleutil.CallMethod(pkcs, "InitializeFromPrivateKey", ContextMachine, pk, "")
	response += fmt.Sprintf("<p>pkcs init result %s</p>", result.ToString())

	if err != nil {
		return fmt.Sprintf("error result 8: %v", err)
	}

	result, err = oleutil.PutProperty(pkcs, "Subject", dn)
	response += fmt.Sprintf("<p>pkcs subject result %s</p>", result.ToString())

	if err != nil {
		return fmt.Sprintf("error result 8: %v", err)
	}

	uxe, err := oleutil.CreateObject("X509Enrollment.CX509Enrollment")
	defer uxe.Release()

	if err != nil {
		return fmt.Sprintf("error result 9: %v", err)
	}

	xe, err := uxe.QueryInterface(ole.IID_IDispatch)
	defer xe.Release()

	if err != nil {
		return fmt.Sprintf("error result 10: %v", err)
	}

	result, err = oleutil.CallMethod(xe, "InitializeFromRequest", pkcs)
	response += fmt.Sprintf("<p>xe init result %s</p>", result.ToString())

	if err != nil {
		return fmt.Sprintf("error result 11: %v", err)
	}

	csr, err := oleutil.CallMethod(xe, "CreateRequest", 1)
	// response += fmt.Sprintf("<p>xe create request result %s</p>", csr.ToString())

	if err != nil {
		return fmt.Sprintf("error result 12: %v", err)
	}

	ca := adcs.NewCertificateAuthorityOle("Lab Root Authority", "certmgr-adds.lab2.internal")
	issued, err := ca.Submit(csr.ToString(), "ServerAuthentication-CngRsa")
	// response += fmt.Sprintf("<p>ca submit result %s</p>", result.ToString())

	if err != nil {
		return fmt.Sprintf("error result 13: %v", err)
	}

	// "X509Enrollment.CX509Enrollment"

	// uxe.Release()
	// xe.Release()

	uxei, err := oleutil.CreateObject("X509Enrollment.CX509Enrollment")
	defer uxei.Release()

	if err != nil {
		return fmt.Sprintf("error result 9: %v", err)
	}

	xei, err := uxei.QueryInterface(ole.IID_IDispatch)
	defer xei.Release()

	if err != nil {
		return fmt.Sprintf("error result 10: %v", err)
	}

	result, err = oleutil.CallMethod(xei, "Initialize", ContextMachine)
	response += fmt.Sprintf("<p>xei init result %s</p>", result.ToString())

	if err != nil {
		return fmt.Sprintf("error result 14: %v", err)
	}

	result, err = oleutil.CallMethod(xei,
		"InstallResponse",
		INSTALL_RESPONSE_RESTRICITON_ALLOW_UNTRUSTED_ROOT,
		issued.ToString(),
		XCN_CRYPT_STRING_BASE64,
		X509_INSTALL_PASSWORD)
	response += fmt.Sprintf("<p>xei install response result %s</p>", result.Value())

	if err != nil {
		return fmt.Sprintf("error result 15: %v", err)
	}

	return response

}
