package adcs

import (
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

type CertificateAuthorityOle struct {
	Hostname string
	Name     string
}

func NewCertificateAuthorityOle(name, hostname string) *CertificateAuthorityOle {
	return &CertificateAuthorityOle{
		Hostname: hostname,
		Name:     name,
	}
}

func (ca *CertificateAuthorityOle) ConnectionString() string {
	return fmt.Sprintf("%s\\%s", ca.Hostname, ca.Name)
}

func (ca *CertificateAuthorityOle) TemplateString(template string) string {
	return fmt.Sprintf("CertificateTemplate:%s", template)
}

func (ca *CertificateAuthorityOle) Submit(csr, template string) (response *ole.VARIANT, err error) {

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
