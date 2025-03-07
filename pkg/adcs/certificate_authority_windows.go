//go:build windows

package adcs

import (
	"context"
	"fmt"

	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
)

var (
	CSR_BASE64                       = 1
	CR_OUT_BASE64              int32 = 0x1
	CR_DISP_DENIED             int32 = 2
	CR_DISP_ISSUED             int32 = 3
	CR_DISP_UNDER_SUBMISSION   int32 = 5
	CR_DISP_ERROR              int32 = 6
	CR_DISP_INCOMPLETE         int32 = 999 // i don't know the value of this constant
	CR_DISP_ISSUED_OUT_OF_BAND int32 = 999 // i don't know the value of this constant
)

type CertificateAuthority interface {
	Submit() (result string, err error)
}

type AdcsCertificateAuthority struct {
	Name     string
	Server   string
	Username string
	Password string
	Port     string
}

func NewAdcsCertificateAuthority(name, server, username, password string) *AdcsCertificateAuthority {
	panic("not implemented")
}

func (ca *AdcsCertificateAuthority) Ping() (err error) {
	panic("not implemented")
}

func (ca *AdcsCertificateAuthority) Request(ctx context.Context, request CertificateSigningRequest) (response CertificateSigningResponse, err error) {
	panic("not implemented")
}

func (ca *AdcsCertificateAuthority) ConnectionString() string {
	return fmt.Sprintf("%s\\%s", ca.Server, ca.Name)
}

func (ca *AdcsCertificateAuthority) TemplateString(template string) string {
	return fmt.Sprintf("CertificateTemplate:%s", template)
}

func (ca *AdcsCertificateAuthority) Submit(csr, template string) (response *ole.VARIANT, err error) {

	ureq, err := oleutil.CreateObject("CertificateAuthority.Request")
	if err != nil {
		return nil, fmt.Errorf("error creating CertificateAuthorityOle object: %v", err)
	}
	defer ureq.Release()

	req, err := ureq.QueryInterface(ole.IID_IDispatch)
	if err != nil {
		return nil, fmt.Errorf("error querying CertificateAuthorityOle query interface: %v", err)
	}
	defer req.Release()

	result, err := oleutil.CallMethod(
		req,
		"Submit",
		CSR_BASE64,
		csr,
		ca.TemplateString(template),
		ca.ConnectionString())

	if err != nil {
		return nil, fmt.Errorf("error calling CertificateAuthorityOle Submit method: %v", err)
	}

	if result.Value() == nil {
		return nil, fmt.Errorf("error calling CertificateAuthorityOle Submit method: result is nil")
	}

	if result.Value().(int32) == CR_DISP_UNDER_SUBMISSION {
		result, err = oleutil.CallMethod(req, "GetRequestId")
		if err != nil {
			return nil, fmt.Errorf("error calling CertificateAuthorityOle GetRequestId method: %v", err)
		}

		return nil, fmt.Errorf("error calling CertificateAuthorityOle Submit method: result is under submission with id %d", result.Value().(int32))
	}

	if result.Value().(int32) != CR_DISP_ISSUED {
		return nil, fmt.Errorf("error calling CertificateAuthorityOle Submit method: result is not issued")
	}

	// fmt.Println(result.Val)

	// oleutil.GetActiveObject())

	result, err = oleutil.CallMethod(req, "GetCertificate", CR_OUT_BASE64)
	if err != nil {
		return nil, fmt.Errorf("error calling CertificateAuthorityOle GetRequestId method: %v", err)
	}

	return result, nil

}
