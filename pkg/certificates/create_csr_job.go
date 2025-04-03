package certificates

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"database/sql"
	"encoding/pem"
	"fmt"
	"log"
	"time"

	"github.com/corymurphy/pkiadmin/pkg/repo"
	"github.com/corymurphy/pkiadmin/pkg/scheduling"
	"github.com/google/uuid"
)

type CreateRsaCsrArguments struct {
	KeyLength    int
	Organization []string
	CommonName   string
	DNSNames     []string
	ID           int64
}
type CreateRsaCsrJob struct{}

func (i CreateRsaCsrArguments) Kind() string { return "CreateRsaCsr" }

func (i *CreateRsaCsrJob) Run(log *log.Logger, args CreateRsaCsrArguments) (next *scheduling.Job, err error) {
	log.Println("creating rsa csr with key length", args.KeyLength)
	ctx := context.Background()

	db, err := sql.Open("sqlite3", "pkiadmin.db")
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}
	defer db.Close()
	queries := repo.New(db)

	privateKey, err := rsa.GenerateKey(rand.Reader, args.KeyLength)
	if err != nil {
		return nil, fmt.Errorf("error generating private key: %w", err)
	}

	privateKeyPem := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey)})

	template := &x509.CertificateRequest{
		Subject: pkix.Name{
			CommonName:   args.CommonName,
			Organization: args.Organization,
		},
		DNSNames: args.DNSNames,
	}

	csr, err := x509.CreateCertificateRequest(rand.Reader, template, privateKey)
	if err != nil {
		fmt.Println("error creating certificate request:", err)
		return nil, err
	}

	now := time.Now()

	if _, err = queries.CreateCertificateContent(ctx, repo.CreateCertificateContentParams{
		CertificateID: args.ID,
		Content:       csr,
		UpdatedAt:     now,
		CreatedAt:     now,
		Name:          "csr",
		Encoding:      "der",
	}); err != nil {
		return nil, fmt.Errorf("failed to create certificate content csr: %w", err)
	}

	if _, err = queries.CreateCertificateContent(ctx, repo.CreateCertificateContentParams{
		CertificateID: args.ID,
		Content:       privateKeyPem,
		UpdatedAt:     now,
		CreatedAt:     now,
		Name:          "privatekey",
		Encoding:      "pem",
	}); err != nil {
		return nil, fmt.Errorf("failed to create certificate content csr: %w", err)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to update certificate request content id: %w", err)
	}

	queries.UpdateCertificateTimelineByRequest(ctx,
		repo.UpdateCertificateTimelineByRequestParams{
			CertificateID: args.ID,
			Event:         int64(Generated),
			Status:        int64(Completed),
			UpdatedAt:     time.Now(),
		})

	return &scheduling.Job{
		Id:         uuid.New(),
		Retry:      true,
		RetryCount: 0,
		CreatedAt:  time.Now(),
		Arguments:  IssueADCSCertificateArguments{ID: args.ID},
	}, nil
}
