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

func CreatePrivateKeyCom() interface{} {
	var err error
	// windows.

	// imagePath, _ := windows.UTF16PtrFromString(`C:\Users\User\Pictures\image.jpg`)
	// fmt.Println("[+] Changing background now...")
	// _, _, err := procSystemParamInfo.Call(20, 0, uintptr(unsafe.Pointer(imagePath)), 0x001A)

	// CX509PrivateKey

	// certEnrollDll.

	// fmt.Print

	connection := &ole.Connection{nil}

	// oleutil.ConnectObject()

	err = connection.Initialize()
	if err != nil {
		return fmt.Sprintf("result: %v", err)
	}
	defer connection.Uninitialize()

	// CX509PrivateKey

	// err = connection.Create("QBXMLRP2.RequestProcessor.1")
	// err = connection.Create("CERTENROLLLib.IX509PrivateKey")
	// \X509Enrollment.CX509PrivateKey

	err = connection.Create("X509Enrollment.CX509PrivateKey")
	if err != nil {
		if err.(*ole.OleError).Code() == ole.CO_E_CLASSSTRING {
			return fmt.Sprintf("result: %v", err)
		}
		return fmt.Sprintf("result: %v", err)
	}
	defer connection.Release()

	dispatch, err := connection.Dispatch()
	if err != nil {
		return fmt.Sprintf("result: %v", err)
	}
	defer dispatch.Release()

	// dispatch.S

	// obj := connection.Object
	// obj.*
	// var pkDispatch *ole.VARIANT

	// _, err = dispatch.Call("OpenConnection2", "", "Test Application 1", 1)
	// pkDispatch, err = dispatch.Call("Create")
	// if err != nil {
	// 	return fmt.Sprintf("error result: %v", err)
	// }

	// CX509CertificateRequestPkcs10 pkcs10 = new CX509CertificateRequestPkcs10();

	// pkcs10.InitializeFromPrivateKey(X509CertificateEnrollmentContext.ContextMachine, privateKey, "");

	// err = connection.Create("X509Enrollment.CX509CertificateRequestPkcs10")
	// if err != nil {
	// 	if err.(*ole.OleError).Code() == ole.CO_E_CLASSSTRING {
	// 		return fmt.Sprintf("result: %v", err)
	// 	}
	// 	return fmt.Sprintf("result: %v", err)
	// }

	pkcs10Id, err := oleutil.ClassIDFrom("X509Enrollment.CX509CertificateRequestPkcs10")
	if err != nil {
		return fmt.Sprintf("error result: %v", err)
	}

	pkId, err := oleutil.ClassIDFrom("X509Enrollment.CX509PrivateKey")
	if err != nil {
		return fmt.Sprintf("error result: %v", err)
	}

	unknownPkcs10, err := ole.CreateInstance(pkcs10Id, nil)
	if err != nil {
		return fmt.Sprintf("error result: %v", err)
	}
	pkcs10, err := unknownPkcs10.QueryInterface(ole.IID_IDispatch)
	if err != nil {
		return fmt.Sprintf("error result: %v", err)
	}

	unknownPk, err := ole.CreateInstance(pkId, nil)
	if err != nil {
		return fmt.Sprintf("error result: %v", err)
	}
	pkDispatch, err := unknownPk.QueryInterface(ole.IID_IDispatch)
	if err != nil {
		return fmt.Sprintf("error result: %v", err)
	}

	_ = oleutil.MustCallMethod(pkDispatch, "Create")

	// pk.ToIDispatch().VTable()

	// return fmt.Sprintf("pkResult %s", result.ToString())

	// result := oleutil.MustCallMethod(pkcs10, "InitializeFromPrivateKey", ContextMachine, uintptr(unsafe.Pointer(pk.RawVTable)), "")
	result := oleutil.MustCallMethod(pkcs10, "InitializeFromPrivateKey",
		ContextMachine,
		pkDispatch.VTable(),
		"")

	// uintptr(unsafe.Pointer(pk))
	// uintptr(unsafe.Pointer(pk.ToIDispatch().RawVTable)),
	// networkEnum := oleutil.MustCallMethod(pkcs10, "GetNetworkConnections").ToIDispatch()

	// err = conn

	// result.ToString()

	// err = fmt.Errorf("something happened")
	// return fmt.Sprintf("success result: %s %s", pkcs10Id.String(), pkId.String())
	return fmt.Sprintf("result %s", result.ToString())
}

func CreatePrivateKey() interface{} {
	return CreatePrivateKeyCom()
}
