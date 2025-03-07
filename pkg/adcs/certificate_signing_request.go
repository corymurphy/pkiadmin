package adcs

type Disposition uint32

const (
	Issued          Disposition = 0x00000003
	UnderSubmission Disposition = 0x00000005
)

// type NewCertificateSigningRequest func() (csr CertificateSigningRequest) {
// 	return CertificateSigningRequest{

// 	}
// }

func NewCertificateSigningRequest() *CertificateSigningRequest {

	// csrBlock := &pem.Block{
	// 	Type:  "CERTIFICATE REQUEST",
	// 	Bytes: csr,
	// }

	// csrPem := pem.EncodeToMemory(csrBlock)

	return &CertificateSigningRequest{
		Attributes: "CertificateTemplate:ServerAuthentication-CngRsa",
	}
}

type CertificateSigningRequest struct {
	// TemplateName string
	// Attributes   map[string]string
	Csr        []byte
	Attributes string
}

type CertificateSigningResponse struct {
	Certificate        []byte
	Return             int32
	Disposition        Disposition
	DispositionMessage string
}
