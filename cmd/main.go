package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/corymurphy/pkiadmin/pkg/certificates"
	"github.com/corymurphy/pkiadmin/pkg/repo"
	"github.com/corymurphy/pkiadmin/views"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	// "github.com/golang-migrate/migrate/database/v4/sqlite3"
	// "github.com/golang-migrate/migrate/source/file"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"

	// _ "github.com/golang-migrate/migrate/v4/database"

	// _ "github.com/golang-migrate/migrate/v4/source/github"
	"github.com/golang-migrate/migrate/v4/source/file"

	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

type SQLiteRepository struct {
	db *sql.DB
}

func NewSQLiteRepository(db *sql.DB) *SQLiteRepository {
	return &SQLiteRepository{
		db: db,
	}
}

func (r *SQLiteRepository) Migrate() error {
	query := `
    CREATE TABLE IF NOT EXISTS websites(
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT NOT NULL UNIQUE,
        url TEXT NOT NULL,
        rank INTEGER NOT NULL
    );
    `

	_, err := r.db.Exec(query)
	return err
}

func main() {

	os.Remove("pkiadmin.db")

	db, err := sql.Open("sqlite3", "pkiadmin.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	instance, err := sqlite3.WithInstance(db, &sqlite3.Config{})
	if err != nil {
		panic(err)
	}

	fSrc, err := (&file.File{}).Open("./db/migrations")
	if err != nil {
		panic(err)
	}

	m, err := migrate.NewWithInstance("file", fSrc, "sqlite3", instance)
	if err != nil {
		panic(err)
	}

	if err = m.Up(); err != nil {
		// fmt.Printf("attempting to rollback \n %v", err)
		// panic(err)
		// panic(m.Down())
		fmt.Println(err)
	}

	e := echo.New()
	e.Use(middleware.Logger())

	e.Renderer = views.NewTemplates([]string{
		"issued.html",
		"certificates/new.html",
		"requests/list.html",
		"requests/view.html",
		"settings/api.html",
	})

	e.Static("/css", "css")
	e.Static("/images", "images")

	queries := repo.New(db)

	e.GET("/", func(c echo.Context) error {
		return c.Render(http.StatusOK, "issued.html", nil)
	})

	e.POST("/certificates/request", func(c echo.Context) error {
		ctx := context.Background()
		// fmt.Println("here")

		displayName := c.FormValue("displayName")

		keyLength, err := strconv.ParseInt(c.FormValue("keyLength"), 10, 64)
		if err != nil {
			return c.String(400, "invalid keylength")
		}

		hashAlgorithmId, err := strconv.ParseInt(c.FormValue("hashAlgorithmId"), 10, 64)
		if err != nil {
			return c.String(400, "invalid keylength")
		}

		cipherAlgorithmId, err := strconv.ParseInt(c.FormValue("cipherAlgorithmId"), 10, 64)
		if err != nil {
			return c.String(400, "invalid keylength")
		}

		cryptoApiId, err := strconv.ParseInt(c.FormValue("cryptoApiId"), 10, 64)
		if err != nil {
			return c.String(400, "invalid keylength")
		}

		signingRequestApiId, err := strconv.ParseInt(c.FormValue("signingRequestApiId"), 10, 64)
		if err != nil {
			return c.String(400, "invalid keylength")
		}

		_, err = queries.CreateCertificateRequest(ctx, repo.CreateCertificateRequestParams{
			DisplayName:                   sql.NullString{String: displayName, Valid: true},
			KeyLength:                     sql.NullInt64{Int64: keyLength, Valid: true},
			HashAlgorithmID:               sql.NullInt64{Int64: hashAlgorithmId, Valid: true},
			CipherAlgorithmID:             sql.NullInt64{Int64: cipherAlgorithmId, Valid: true},
			CertificateCryptographicApiID: sql.NullInt64{Int64: cryptoApiId, Valid: true},
			SigningRequestApiID:           sql.NullInt64{Int64: signingRequestApiId, Valid: true},
		})

		if err != nil {
			return c.String(500, "something happened")
		}

		return c.Redirect(302, "/requests/list.html")
	})

	e.GET("certificates/new.html", func(c echo.Context) error {
		ctx := context.Background()
		data := make(map[string]interface{})

		cryptoApis, err := queries.ListCertCryptoApi(ctx)

		if err != nil {
			return c.Render(500, "error", err)
		}

		signingApis, err := queries.ListSigningRequestApi(ctx)

		if err != nil {
			return c.Render(500, "error", err)
		}

		cipherAlgorithms, err := queries.ListCipherAlgorithm(ctx)

		if err != nil {
			return c.Render(500, "error", err)
		}

		hasAlgorithms, err := queries.ListHashAlgorithm(ctx)

		if err != nil {
			return c.Render(500, "error", err)
		}

		keyLengths := []string{}

		for _, alg := range cipherAlgorithms {
			keyLengths = append(keyLengths, strconv.FormatInt(alg.Keysize.Int64, 10))
		}

		data["CryptoApis"] = cryptoApis
		data["SigningApis"] = signingApis
		data["CipherAlgorithms"] = cipherAlgorithms
		data["HashAlgorithms"] = hasAlgorithms
		data["KeyLengths"] = keyLengths

		return c.Render(http.StatusOK, "certificates/new.html", data)
	})

	e.GET("requests/list.html", func(c echo.Context) error {
		ctx := context.Background()
		// requests, err := queries.ListCertificateRequest(ctx)
		requests, err := queries.CertificateRequestsAndHashAlgorithm(ctx)

		// requests[0].CertificateCryptographicApi
		// requests[0].SigningRequestApi

		if err != nil {
			return c.Render(500, "error", err)
		}
		data := make(map[string]interface{})
		data["CertificateRequests"] = requests

		return c.Render(http.StatusOK, "requests/list.html", data)
	})

	e.GET("requests/view/:id", func(c echo.Context) error {
		ctx := context.Background()

		idStr := c.Param("id")

		if err != nil {
			return c.String(400, "invalid id")
		}

		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			return c.String(400, "invalid id")
		}

		csr, err := queries.GetCertificateRequestDetailed(ctx, id)

		if err != nil {
			return c.Render(500, "error", err)
		}

		return c.Render(http.StatusOK, "requests/view.html", csr)
	})

	e.GET("settings/api.html", func(c echo.Context) error {
		ctx := context.Background()
		cryptoApis, err := queries.ListCertCryptoApi(ctx)

		if err != nil {
			return c.Render(500, "error", err)
		}

		signingApis, err := queries.ListSigningRequestApi(ctx)

		if err != nil {
			return c.Render(500, "error", err)
		}

		cipherAlgorithms, err := queries.ListCipherAlgorithm(ctx)

		if err != nil {
			return c.Render(500, "error", err)
		}

		hasAlgorithms, err := queries.ListHashAlgorithm(ctx)

		if err != nil {
			return c.Render(500, "error", err)
		}

		data := make(map[string]interface{})
		data["CryptoApis"] = cryptoApis
		data["SigningApis"] = signingApis
		data["CipherAlgorithms"] = cipherAlgorithms
		data["HashAlgorithms"] = hasAlgorithms

		return c.Render(http.StatusOK, "settings/api.html", data)
	})

	e.DELETE("settings/cryptoapi/:id", func(c echo.Context) error {

		ctx := context.Background()

		idStr := c.Param("id")

		if err != nil {
			return c.String(400, "invalid id")
		}

		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			return c.String(400, "invalid id")
		}

		// api, err := queries.GetCertCryptoApi(ctx, id)

		queries.DeleteCertCryptoApi(ctx, id)

		return c.NoContent(200)
	})

	e.GET("sandbox", func(c echo.Context) error {
		data := certificates.CreatePrivateKey()
		return c.Render(http.StatusOK, "error", data)
	})

	e.Logger.Fatal(e.Start(":8956"))
}
