package certificates

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

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
	ctx := context.Background() // TODO: content should be provided by worker

	db, err := sql.Open("sqlite3", "pkiadmin.db")
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}
	defer db.Close()
	queries := repo.New(db)

	content, err := queries.GetCertificateContentByNameEncodingRequestID(
		ctx, repo.GetCertificateContentByNameEncodingRequestIDParams{
			Name:                 "csr",
			Encoding:             "der",
			CertificateRequestID: args.ID,
		})

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
		Csr:        content.Content,
		Attributes: "CertificateTemplate:ServerAuthentication-CngRsa",
	})
	if err != nil {
		queries.UpdateCertificateRequestTimelineByRequest(ctx,
			repo.UpdateCertificateRequestTimelineByRequestParams{
				CertificateRequestID: args.ID,
				Event:                int64(Submitted),
				Status:               int64(Failed),
				UpdatedAt:            time.Now(),
			})
		return nil, fmt.Errorf("error requesting certificate: %w", err)
	}

	log.Printf("response.Certificate:\n %+v", string(response.Certificate))
	log.Println("response.Disposition", response.Disposition)
	log.Println("response.DispositionMessage", response.DispositionMessage)

	queries.UpdateCertificateRequestTimelineByRequest(ctx,
		repo.UpdateCertificateRequestTimelineByRequestParams{
			CertificateRequestID: args.ID,
			Event:                int64(Submitted),
			Status:               int64(Completed),
			UpdatedAt:            time.Now(),
		})

	if response.Disposition != adcs.Issued {
		queries.UpdateCertificateRequestTimelineByRequest(ctx,
			repo.UpdateCertificateRequestTimelineByRequestParams{
				CertificateRequestID: args.ID,
				Event:                int64(Issued),
				Status:               int64(Failed),
				UpdatedAt:            time.Now(),
			})
		return nil, fmt.Errorf("error requesting certificate: %d %s",
			response.Disposition, response.DispositionMessage)
	}

	if _, err = queries.CreateCertificateContent(ctx, repo.CreateCertificateContentParams{
		CertificateRequestID: args.ID,
		Content:              response.Certificate,
		UpdatedAt:            time.Now(),
		CreatedAt:            time.Now(),
		Name:                 "certificate",
		Encoding:             "pem",
	}); err != nil {
		return nil, fmt.Errorf("failed to create certificate content csr: %w", err)
	}

	if err != nil {
		queries.UpdateCertificateRequestTimelineByRequest(ctx,
			repo.UpdateCertificateRequestTimelineByRequestParams{
				CertificateRequestID: args.ID,
				Event:                int64(Issued),
				Status:               int64(Failed),
				UpdatedAt:            time.Now(),
			})
		return nil, fmt.Errorf("error updating certificate content: %w", err)
	}

	queries.UpdateCertificateRequestTimelineByRequest(ctx,
		repo.UpdateCertificateRequestTimelineByRequestParams{
			CertificateRequestID: args.ID,
			Event:                int64(Issued),
			Status:               int64(Completed),
			UpdatedAt:            time.Now(),
		})

	return nil, nil
}
