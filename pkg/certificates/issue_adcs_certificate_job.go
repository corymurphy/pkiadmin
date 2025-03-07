package certificates

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/corymurphy/pkiadmin/pkg/adcs"
	"github.com/corymurphy/pkiadmin/pkg/repo"
	"github.com/corymurphy/pkiadmin/pkg/scheduling"
)

type IssueADCSCertificateArguments struct {
	ID int64
}
type IssueADCSCertificate struct{}

func (i IssueADCSCertificateArguments) Kind() string { return "IssueADCSCertificate" }

func (i *IssueADCSCertificate) Run(log *log.Logger, args IssueADCSCertificateArguments) (next *scheduling.Job, err error) {
	log.Println("creating adcs certificate")
	ctx := context.Background()

	db, err := sql.Open("sqlite3", "pkiadmin.db")
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}
	defer db.Close()
	queries := repo.New(db)

	content, err := queries.GetCertificateRequestWithContent(ctx, args.ID)

	log.Println("csr id", args.ID)

	if err != nil {
		return nil, fmt.Errorf("error getting certificate request: %w", err)
	}

	cas, err := queries.ListCertificateAuthorities(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting certificate authorities: %w", err)
	}

	if len(cas) == 0 {
		return nil, fmt.Errorf("no certificate authorities found")
	}

	caData := cas[0]
	ca := adcs.NewCA(
		adcs.WithName(caData.Name),
		adcs.WithServer(caData.Server),
		adcs.WithUsername(caData.Username),
		adcs.WithPassword(caData.Password),
		adcs.WithPort("135"), // TODO: get from repo
	)

	response, err := ca.Request(ctx, adcs.CertificateSigningRequest{
		Csr:        content.Csr,
		Attributes: "CertificateTemplate:ServerAuthentication-CngRsa",
	})
	if err != nil {
		return nil, fmt.Errorf("error requesting certificate: %w", err)
	}

	log.Printf("response.Certificate:\n %+v", string(response.Certificate))
	log.Println("response.Disposition", response.Disposition)
	log.Println("response.DispositionMessage", response.DispositionMessage)

	if response.Disposition != adcs.Issued {
		return nil, fmt.Errorf("error requesting certificate: %d %s",
			response.Disposition, response.DispositionMessage)
	}

	err = queries.UpdateCertificateContentPublicKey(ctx, repo.UpdateCertificateContentPublicKeyParams{
		ID:        content.ID,
		PublicKey: response.Certificate,
	})

	if err != nil {
		return nil, fmt.Errorf("error updating certificate content: %w", err)
	}

	return nil, nil
}
