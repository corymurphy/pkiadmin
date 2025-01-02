package certificates

import (
	"fmt"

	"golang.org/x/sys/windows"

	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
)

var (
	user32DLL           = windows.NewLazyDLL("user32.dll")
	certEnrollDll       = windows.NewLazyDLL("CertEnroll.dll")
	procSystemParamInfo = user32DLL.NewProc("SystemParametersInfoW")
	ContextMachine      = 2
)

func CreatePrivateKeySandbox() interface{} {
	var err error
	var response string
	var result *ole.VARIANT

	err = ole.CoInitialize(0)
	if err != nil {
		return fmt.Sprintf("error result 1: %v", err)
	}
	defer ole.CoUninitialize()

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

	if err != nil {
		return fmt.Sprintf("error result 10: %v", err)
	}

	result, err = oleutil.CallMethod(xe, "InitializeFromRequest", pkcs)
	response += fmt.Sprintf("<p>xe init result %s</p>", result.ToString())

	if err != nil {
		return fmt.Sprintf("error result 11: %v", err)
	}

	csr, err := oleutil.CallMethod(xe, "CreateRequest", 1)
	response += fmt.Sprintf("<p>xe create request result %s</p>", csr.ToString())

	if err != nil {
		return fmt.Sprintf("error result 12: %v", err)
	}

	ureq, err := oleutil.CreateObject("CertificateAuthority.Request")
	defer ureq.Release()

	if err != nil {
		return fmt.Sprintf("error result 13: %v", err)
	}

	req, err := ureq.QueryInterface(ole.IID_IDispatch)

	if err != nil {
		return fmt.Sprintf("error result 14: %v", err)
	}

	result, err = oleutil.CallMethod(req, "Submit", 1, csr, "CertificateTemplate:ServerAuthentication-CngRsa", "certmgr-adds.lab2.internal\\Lab Root Authority")
	response += fmt.Sprintf("<p>req submit result %s</p>", result.ToString())

	if err != nil {
		return fmt.Sprintf("error result 15: %v", err)
	}

	return response

}

func CreatePrivateKey() interface{} {
	return CreatePrivateKeySandbox()
}
