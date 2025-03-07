package adcs

import "github.com/corymurphy/pkiadmin/pkg/repo"

type AdcsCertificateAuthorityOption func(*AdcsCertificateAuthority)

func WithPort(port string) AdcsCertificateAuthorityOption {
	return func(ca *AdcsCertificateAuthority) {
		ca.Port = port
	}
}

func WithName(name string) AdcsCertificateAuthorityOption {
	return func(ca *AdcsCertificateAuthority) {
		ca.Name = name
	}
}

func WithServer(server string) AdcsCertificateAuthorityOption {
	return func(ca *AdcsCertificateAuthority) {
		ca.Server = server
	}
}

func WithUsername(username string) AdcsCertificateAuthorityOption {
	return func(ca *AdcsCertificateAuthority) {
		ca.Username = username
	}
}

func WithPassword(password string) AdcsCertificateAuthorityOption {
	return func(ca *AdcsCertificateAuthority) {
		ca.Password = password
	}
}

func FromListRepository(repo repo.ListCertificateAuthoritiesRow) AdcsCertificateAuthorityOption {
	return func(ca *AdcsCertificateAuthority) {
		ca.Name = repo.Name
		ca.Server = repo.Server
		ca.Username = repo.Username
		ca.Password = repo.Password
	}
}

func NewCA(opts ...AdcsCertificateAuthorityOption) *AdcsCertificateAuthority {
	ca := &AdcsCertificateAuthority{}

	for _, opt := range opts {
		opt(ca)
	}

	return ca
}
