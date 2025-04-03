package main

import (
	"context"
	"crypto/sha1"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"text/template"
	"time"

	"github.com/corymurphy/pkiadmin/pkg/adcs"
	"github.com/corymurphy/pkiadmin/pkg/certificates"
	"github.com/corymurphy/pkiadmin/pkg/repo"
	"github.com/corymurphy/pkiadmin/pkg/scheduling"
	"github.com/corymurphy/pkiadmin/pkg/shared"
	"github.com/corymurphy/pkiadmin/views"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

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

// func init() {
// 	godotenv.Load(".env")
// }

func main() {

	// os.Remove("pkiadmin.db")

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

	err = m.Up() // TODO: this returns an error if there are no changes, look into this
	if err != nil {
		fmt.Println(err)
	}

	e := echo.New()
	// e.Use(middleware.Logger())

	// TODO figure out a better way to do this
	e.Renderer = views.NewTemplates([]string{
		"home.html",
		"certificates/new.html",
		"requests/list.html",
		"requests/view.html",
		"settings/api.html",
		"settings/ca.html",
		"scheduler/scheduled.html",
		"scheduler/completed.html",
		"scheduler/failed.html",
		"scheduler/queued.html",
		"scheduler/inprogress.html",
		"scheduler/workers.html",
	},
		[]string{
			"views/settings/*partial.html",
			"views/*partial.html",
		},
		template.FuncMap{
			"formatTime": func(t time.Time) string {
				return t.Format("2006-01-02 15:04:05")
			},
			"splitNewLine": func(log string) []string {
				return strings.Split(log, "\n")
			},
			"jsonPrint": func(data interface{}) string {
				result, _ := json.MarshalIndent(data, "", "    ")
				return string(result)
			},
			"toLower": func(s string) string {
				return strings.ToLower(s)
			},
			"CertificateTimelineStatusString": func(status int64) string {
				return certificates.RequestTimelineStatus(status).String()
			},
			"CertificateTimelineEventString": func(event int64) string {
				return certificates.RequestTimelineEvent(event).String()
			},
			// TODO use generics
			"ByteLength": func(data []byte) int {
				return len(data)
			},
		})

	e.Static("/css", "css")
	e.Static("/images", "images")
	e.Static("/js", "js")
	e.Static("/favicon.ico", "favicon.ico")

	queries := repo.New(db)
	queue := scheduling.NewQueue(queries)

	scheduler := scheduling.New(
		scheduling.WithProcessor(&certificates.CreateRsaCsrJob{}),
		scheduling.WithProcessor(&certificates.IssueADCSCertificate{}),
		scheduling.WithWorkerCount(2),
		scheduling.WithQueue(queue),
	)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	e.GET("/", func(c echo.Context) error {
		return c.Render(http.StatusOK, "home.html", nil)
	})

	e.POST("/certificates/request", func(c echo.Context) error {
		ctx := context.Background()

		cn := c.FormValue("cn")
		template := c.FormValue("template")
		org := c.FormValue("org")
		san := c.FormValue("san")

		keyLength, err := strconv.ParseInt(c.FormValue("keyLength"), 10, 64)
		if err != nil {
			return c.String(400, "invalid keylength")
		}

		hashAlgorithmId, err := strconv.ParseInt(c.FormValue("hashAlgorithmId"), 10, 64)
		if err != nil {
			return c.String(400, "invalid hash algorithm")
		}

		cipherAlgorithmId, err := strconv.ParseInt(c.FormValue("cipherAlgorithmId"), 10, 64)
		if err != nil {
			return c.String(400, "invalid cipher algorithm")
		}

		ca, err := strconv.ParseInt(c.FormValue("ca"), 10, 64)
		if err != nil {
			return c.String(400, "invalid ca")
		}

		now := time.Now()
		request, err := queries.CreateCertificate(ctx, repo.CreateCertificateParams{
			DisplayName:             cn,
			Organization:            sql.NullString{String: org, Valid: true},
			SubjectAlternativeNames: sql.NullString{String: san, Valid: true},
			KeyLength:               keyLength,
			HashAlgorithmID:         hashAlgorithmId,
			CipherAlgorithmID:       cipherAlgorithmId,
			RequestedOn:             now,
		})
		if err != nil {
			return c.String(500, fmt.Errorf("error creating certificate request: %w", err).Error())
		}

		_, err = queries.CreateCertificateRequestAuthority(ctx, repo.CreateCertificateRequestAuthorityParams{
			CertificateID:          request,
			CertificateAuthorityID: ca,
			TemplateName:           template,
		})
		if err != nil {
			return c.String(500, fmt.Errorf("error creating certificate request authority: %w", err).Error())
		}

		timeline := []struct{ Status, Event int64 }{
			{int64(certificates.Completed), int64(certificates.Requested)},
			{int64(certificates.Pending), int64(certificates.Approved)},
			{int64(certificates.Pending), int64(certificates.Generated)},
			{int64(certificates.Pending), int64(certificates.Submitted)},
			{int64(certificates.Pending), int64(certificates.Issued)},
		}

		for _, t := range timeline {
			_, err = queries.CreateCertificateTimeline(ctx, repo.CreateCertificateTimelineParams{
				CertificateID: request,
				Event:         t.Event,
				Status:        t.Status,
				CreatedAt:     now,
				UpdatedAt:     now,
			})
			if err != nil {
				return c.String(500, fmt.Errorf("error creating certificate request timeline: %w", err).Error())
			}
		}

		return c.Redirect(302, "/requests/list.html")
	})

	e.GET("certificates/new.html", func(c echo.Context) error {
		ctx := context.Background()
		data := make(map[string]interface{})

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

		cas, err := queries.ListCertificateAuthorities(ctx)
		if err != nil {
			return c.Render(500, "error", err)
		}

		data["CertificateAuthorities"] = cas
		data["CipherAlgorithms"] = cipherAlgorithms
		data["HashAlgorithms"] = hasAlgorithms
		data["KeyLengths"] = keyLengths

		return c.Render(http.StatusOK, "certificates/new.html", data)
	})

	e.GET("requests/list.html", func(c echo.Context) error {
		ctx := context.Background()
		// requests, err := queries.ListCertificate(ctx)
		// requests, err := queries.CertificatesAndHashAlgorithm(ctx)

		page, err := strconv.ParseInt(c.QueryParam("page"), 10, 64)
		if err != nil {
			page = 0
		}
		limit, err := strconv.ParseInt(c.QueryParam("limit"), 10, 64)
		if err != nil {
			limit = 10
		}
		fmt.Println(page, limit)

		count, err := queries.GetCertificatesCount(ctx)
		if err != nil {
			return c.Render(500, "error", err)
		}

		requests, err := queries.CertificatesAndHashAlgorithmPaginated(ctx, repo.CertificatesAndHashAlgorithmPaginatedParams{
			Limit:  limit,
			Offset: page * limit,
		})

		if err != nil {
			return c.Render(500, "error", err)
		}
		data := make(map[string]interface{})
		data["Certificates"] = requests
		data["CertificateCount"] = count
		data["Pages"] = make([]int, shared.RoundUp(float64(count)/float64(limit)))
		data["Page"] = page
		data["Start"] = page * limit

		if ((limit * page) + limit) > count {
			data["End"] = count
		} else {
			data["End"] = (limit * page) + limit
		}

		return c.Render(http.StatusOK, "requests/list.html", data)
	})

	e.DELETE("requests/:id", func(c echo.Context) (err error) {
		ctx := context.Background()
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			return c.String(400, "invalid id")
		}
		time.Sleep(1 * time.Second)
		err = queries.DeleteCertificate(ctx, id)
		if err != nil {
			return c.String(500, "something happened")
		}

		queries.DeleteCertificateTimelines(ctx, id)
		return c.NoContent(200)
	})

	e.POST("requests/approve/:id", func(c echo.Context) error {
		ctx := c.Request().Context()
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)

		if err != nil {
			return c.String(400, "invalid id")
		}

		req, err := queries.GetCertificate(ctx, id)
		if err != nil {
			return c.String(500, "Unable to approve request")
		}
		if req.Status == repo.Approved {
			return c.String(400, "request already approved")
		}

		err = queries.UpdateCertificateStatus(ctx,
			repo.UpdateCertificateStatusParams{
				Status: repo.Approved,
				ID:     id,
			})

		if err != nil {
			return c.String(500, "Unable to approve request")
		}

		queries.UpdateCertificateTimelineByRequest(ctx,
			repo.UpdateCertificateTimelineByRequestParams{
				CertificateID: id,
				Event:         int64(certificates.Approved),
				Status:        int64(certificates.Completed),
				UpdatedAt:     time.Now(),
			})

		performAt := time.Now().Add((5 * time.Second))
		queue.EnqueueJob(ctx, performAt, &scheduling.Job{
			Id:         uuid.New(),
			Retry:      false,
			RetryCount: 0,
			CreatedAt:  time.Now(),
			Arguments: certificates.CreateRsaCsrArguments{
				Organization: []string{req.Organization.String},
				CommonName:   req.DisplayName,
				DNSNames:     []string{req.DisplayName},
				KeyLength:    int(req.KeyLength),
				ID:           id,
			},
		})

		return c.String(200, "Approved")
	})

	e.GET("requests/view/:id", func(c echo.Context) error {
		ctx := context.Background()

		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			return c.String(400, "invalid id")
		}

		request, err := queries.GetCertificateDetailed(ctx, id)

		if err != nil {
			fmt.Println("error getting request", err)
			return c.Render(500, "error", err)
		}

		timeline, _ := queries.ListCertificateTimeline(ctx, request.ID)
		contents, err := queries.ListCertificateContent(ctx, id)
		if err != nil {
			fmt.Println("error getting contents", err)
		}

		// cert := []byte{}
		// privateKey := []byte{}
		// for _, content := range contents {
		// 	if content.Name == "certificate" {
		// 		cert = content.Content
		// 	}
		// 	if content.Name == "privatekey" {
		// 		privateKey = content.Content
		// 	}
		// }

		// if len(cert) > 0 && len(privateKey) > 0 {

		// }

		data := make(map[string]interface{})
		data["ID"] = request.ID
		data["DisplayName"] = request.DisplayName
		data["KeyLength"] = request.KeyLength
		data["HashAlgorithm"] = request.HashAlgorithm.String
		data["CipherAlgorithm"] = request.CipherAlgorithm.String

		data["Contents"] = contents
		data["Timeline"] = timeline

		csrContent, err := queries.GetCertificateContentByNameEncodingRequestID(ctx,
			repo.GetCertificateContentByNameEncodingRequestIDParams{
				Name:          "csr",
				Encoding:      "der",
				CertificateID: id,
			})
		if err != nil {
			return c.Render(http.StatusOK, "requests/view.html", data)
		}

		csr, err := x509.ParseCertificateRequest(csrContent.Content)
		if err != nil {
			fmt.Println("error parsing csr", err)
			return c.Render(http.StatusOK, "requests/view.html", data)
		}

		sigHash := sha1.Sum(csr.Signature)
		data["Thumbprint"] = hex.EncodeToString(sigHash[:])

		return c.Render(http.StatusOK, "requests/view.html", data)
	})

	e.GET("certificates/:id/download", func(c echo.Context) error {
		ctx := context.Background()

		name := c.QueryParam("name")
		encoding := c.QueryParam("encoding")
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			return c.String(400, "invalid id")
		}

		contents, err := queries.ListCertificateContent(ctx, id)
		if err != nil {
			return c.Render(500, "error", err)
		}
		for _, content := range contents {
			if content.Name == name && content.Encoding == encoding {
				c.Response().Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.%s", name, encoding))
				c.Response().Header().Set("Content-Length", fmt.Sprintf("%d", len(content.Content)))
				c.Response().Header().Set("Content-Type", "application/x-x509-ca-cert")
				c.Response().Write(content.Content)
				c.Response().WriteHeader(200)
				return nil
			}
		}
		return c.String(404, "certificate content not found")
	})

	// e.GET("requests/:id/download/csr.pem", func(c echo.Context) error {
	// 	ctx := context.Background()

	// 	idStr := c.Param("id")

	// 	if err != nil {
	// 		return c.String(400, "invalid id")
	// 	}

	// 	id, err := strconv.ParseInt(idStr, 10, 64)
	// 	if err != nil {
	// 		return c.String(400, "invalid id")
	// 	}

	// 	request, err := queries.GetCertificateWithContent(ctx, id)

	// 	if err != nil {
	// 		return c.Render(500, "error", err)
	// 	}

	// 	csrBlock := &pem.Block{
	// 		Type:  "CERTIFICATE REQUEST",
	// 		Bytes: request.Csr,
	// 	}
	// 	csrPem := pem.EncodeToMemory(csrBlock)

	// 	return c.String(200, string(csrPem))
	// })

	// e.GET("requests/:id/download/privatekey.pem", func(c echo.Context) error {
	// 	ctx := context.Background()

	// 	idStr := c.Param("id")

	// 	if err != nil {
	// 		return c.String(400, "invalid id")
	// 	}

	// 	id, err := strconv.ParseInt(idStr, 10, 64)
	// 	if err != nil {
	// 		return c.String(400, "invalid id")
	// 	}

	// 	request, err := queries.GetCertificateWithContent(ctx, id)

	// 	if err != nil {
	// 		return c.Render(500, "error", err)
	// 	}

	// 	return c.String(200, string(request.PrivateKey))
	// })

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

	e.GET("/ca/status/:id", func(c echo.Context) error {
		idStr := c.Param("id")

		data := make(map[string]interface{})
		data["ID"] = idStr
		data["Color"] = "green"
		data["URL"] = "/ca/status/{{ .ID }}"
		data["Message"] = "Connected"

		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			return c.String(400, "invalid id")
		}
		caData, err := queries.GetCertificateAuthority(c.Request().Context(), id)
		if err != nil {
			return c.Render(500, "error", err)
		}

		ca := adcs.NewCA(
			adcs.WithName(caData.Name),
			adcs.WithServer(caData.Server),
			adcs.WithUsername(caData.Username),
			adcs.WithPassword(caData.Password),
			adcs.WithPort("135"), // TODO: get from repo
		)
		err = ca.Ping()

		if err != nil {
			data["Color"] = "red"
			data["Message"] = "Disconnected"
		}
		return c.Render(200, "badge-trigger", data)
	})

	e.DELETE("ca/:id", func(c echo.Context) error {
		ctx := context.Background()
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			return c.String(400, "invalid id")
		}
		time.Sleep(1 * time.Second)
		err = queries.DeleteCertificateAuthority(ctx, id)
		if err != nil {
			return c.String(500, "something happened")
		}
		return c.NoContent(200)
	})

	e.GET("ca/templates.html", func(c echo.Context) (err error) {
		ctx := c.Request().Context()

		id, err := strconv.ParseInt(c.FormValue("ca"), 10, 64)
		if err != nil {
			return c.String(400, "invalid id")
		}

		caData, err := queries.GetCertificateAuthority(ctx, id)

		if err != nil {
			return c.Render(500, "error", err)
		}

		data := make(map[string]interface{})

		ca := adcs.NewCA(
			adcs.WithName(caData.Name),
			adcs.WithServer(caData.Server),
			adcs.WithUsername(caData.Username),
			adcs.WithPassword(caData.Password),
			adcs.WithPort("135"), // TODO: get from repo
		)

		templates, err := ca.Templates(ctx)
		if err != nil {
			return c.Render(500, "error", err)
		}

		data["Templates"] = templates

		return c.Render(200, "ca-templates", data)
	})

	e.GET("settings/ca.html", func(c echo.Context) error {
		ctx := c.Request().Context()
		cas, err := queries.ListCertificateAuthorities(ctx)

		if err != nil {
			return c.Render(500, "error", err)
		}

		data := make(map[string]interface{})

		var caList []interface{}

		for _, caData := range cas {
			ca := adcs.NewCA(
				adcs.WithName(caData.Name),
				adcs.WithServer(caData.Server),
				adcs.WithUsername(caData.Username),
				adcs.WithPassword(caData.Password),
				adcs.WithPort("135"), // TODO: get from repo
			)

			templates, _ := ca.Templates(ctx)

			caList = append(caList, struct {
				ID        int64
				Name      string
				Server    string
				Username  string
				Templates []adcs.CertificateAuthorityTemplate
			}{
				ID:        caData.ID,
				Name:      caData.Name,
				Server:    caData.Server,
				Username:  caData.Username,
				Templates: templates,
			})
		}

		data["CertificateAuthorities"] = caList

		return c.Render(http.StatusOK, "settings/ca.html", data)
	})

	e.POST("/settings/ca", func(c echo.Context) error {
		name := c.FormValue("name")
		server := c.FormValue("server")
		username := c.FormValue("username")
		password := c.FormValue("password")

		cid, err := queries.CreateCredential(c.Request().Context(), repo.CreateCredentialParams{
			Username: username,
			Password: password,
		})

		if err != nil {
			return c.Render(500, "error", err)
		}

		caid, err := queries.CreateCertificateAuthority(
			c.Request().Context(),
			repo.CreateCertificateAuthorityParams{
				Name:         name,
				Server:       server,
				CredentialID: cid,
			},
		)

		if err != nil {
			return c.Render(500, "error", err)
		}

		ca, err := queries.GetCertificateAuthority(c.Request().Context(), caid)

		if err != nil {
			return c.Render(500, "error", err)
		}

		data := make(map[string]interface{})
		data["ID"] = ca.ID
		data["Name"] = ca.Name
		data["Server"] = ca.Server
		data["Username"] = ca.Username
		data["Password"] = "********"

		return c.Render(200, "ca-add", data)
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
		ctx := c.Request().Context()

		data := struct {
			err  error
			data string
		}{}
		cas, _ := queries.ListCertificateAuthorities(ctx)
		caData := cas[0]
		ca := adcs.NewCA(
			adcs.WithName(caData.Name),
			adcs.WithServer(caData.Server),
			adcs.WithUsername(caData.Username),
			adcs.WithPassword(caData.Password),
			adcs.WithPort("135"), // TODO: get from repo
		)

		// err := ca.Ping()

		// if err != nil {
		// 	data.data = "error pinging ca"
		// 	data.err = err
		// 	return c.Render(http.StatusOK, "error", data)
		// }

		_, err := ca.Templates(ctx)
		// data.data = response
		data.err = err

		return c.Render(http.StatusOK, "error", data)
	})

	e.GET("scheduler/workers.html", func(c echo.Context) error {
		// cas, err := queries.ListCertificateAuthorities(c.Request().Context())
		// if err != nil {
		// 	return c.Render(500, "error", err)
		// }

		data := make(map[string]interface{})
		data["Workers"] = scheduler.Workers()

		return c.Render(http.StatusOK, "scheduler/workers.html", data)
	})

	e.GET("scheduler/scheduled.html", func(c echo.Context) error {
		jobs, err := queries.ListScheduledSet(c.Request().Context())
		if err != nil {
			return c.Render(500, "error", err)
		}

		data := make(map[string]interface{})
		data["ScheduledJobs"] = jobs

		return c.Render(http.StatusOK, "scheduler/scheduled.html", data)
	})

	e.DELETE("scheduler/scheduled/:id", func(c echo.Context) error {
		ctx := c.Request().Context()

		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			return c.String(400, "invalid id")
		}

		err = queries.DeleteScheduledSet(ctx, id)
		if err != nil {
			return c.String(500, "something happened")
		}
		return c.NoContent(200)
	})

	e.DELETE("scheduler/queued/:id", func(c echo.Context) error {
		ctx := c.Request().Context()

		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			return c.String(400, "invalid id")
		}

		err = queries.DeleteSchedulerQueueJob(ctx, id)
		if err != nil {
			return c.String(500, "something happened")
		}
		return c.NoContent(200)
	})

	e.GET("scheduler/queued.html", func(c echo.Context) error {
		ctx := c.Request().Context()

		jobs, err := queries.ListSchedulerQueueJobs(ctx)
		if err != nil {
			return c.Render(500, "error", err)
		}

		data := make(map[string]interface{})
		data["Queued"] = jobs

		return c.Render(http.StatusOK, "scheduler/queued.html", data)
	})

	e.GET("scheduler/inprogress.html", func(c echo.Context) error {
		jobs, err := queries.ListInProgressSet(c.Request().Context())
		if err != nil {
			return c.Render(500, "error", err)
		}

		data := make(map[string]interface{})
		data["InProgress"] = jobs

		return c.Render(http.StatusOK, "scheduler/inprogress.html", data)
	})

	e.GET("scheduler/scheduled/count", func(c echo.Context) error {
		ctx := c.Request().Context()

		count, err := queries.CountScheduledSet(ctx)

		if err != nil {
			return c.String(500, "something happened")
		}

		data := make(map[string]interface{})
		data["Value"] = count
		data["UpdatePath"] = "/scheduler/scheduled/count"

		return c.String(200, fmt.Sprintf("%d", count))
	})

	e.GET("scheduler/queued/count", func(c echo.Context) error {
		ctx := c.Request().Context()

		count, err := queries.CountSchedulerQueueJob(ctx)

		if err != nil {
			return c.String(500, "something happened")
		}

		return c.String(200, fmt.Sprintf("%d", count))
	})

	e.GET("scheduler/inprogress/count", func(c echo.Context) error {
		ctx := c.Request().Context()

		count, err := queries.CountInProgressSet(ctx)

		if err != nil {
			return c.String(500, "something happened")
		}

		return c.String(200, fmt.Sprintf("%d", count))
	})

	e.POST("scheduler/failed/retry/:id", func(c echo.Context) error {
		ctx := c.Request().Context()

		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			return c.String(400, "invalid id")
		}

		failed, err := queries.GetFailedJob(ctx, id)
		if err != nil {
			fmt.Printf("error retrying job: %v", err)
			return c.String(500, "something happened")
		}

		metadata := scheduling.Metadata{
			Id:         failed.ID,
			Retry:      failed.Retry,
			RetryCount: failed.RetryCount,
			CreatedAt:  failed.CreatedAt,
		}
		job, err := scheduler.GetProcessor(failed.Processor, failed.Arguments, metadata)

		if err != nil {
			fmt.Printf("error retrying job: %v", err)
			return c.String(500, "something happened")
		}

		performAt := time.Now().Add((5 * time.Second))
		err = queue.EnqueueJob(c.Request().Context(), performAt, &scheduling.Job{
			Id:         uuid.New(),
			Retry:      false,
			RetryCount: 0,
			CreatedAt:  time.Now(),
			Arguments:  job.Args(),
		})

		if err != nil {
			fmt.Printf("error retrying job: %v", err)
			return c.String(500, "something happened")
		}

		err = queries.DeleteFailedJob(ctx, id)
		if err != nil {
			fmt.Printf("error deleting failed job: %v", err)
		}

		return c.NoContent(200)

	})

	e.DELETE("scheduler/inprogress/:id", func(c echo.Context) error {
		ctx := c.Request().Context()

		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			return c.String(400, "invalid id")
		}

		err = queries.DeleteInProgressSet(ctx, id)
		if err != nil {
			return c.String(500, "something happened")
		}
		return c.NoContent(200)
	})

	e.DELETE("scheduler/completed/:id", func(c echo.Context) error {
		ctx := c.Request().Context()

		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			return c.String(400, "invalid id")
		}

		err = queries.DeleteCompletedJob(ctx, id)
		if err != nil {
			return c.String(500, "something happened")
		}
		return c.NoContent(200)
	})

	e.DELETE("scheduler/failed/:id", func(c echo.Context) error {
		ctx := c.Request().Context()

		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			return c.String(400, "invalid id")
		}

		err = queries.DeleteFailedJob(ctx, id)
		if err != nil {
			return c.String(500, "something happened")
		}
		return c.NoContent(200)
	})

	e.GET("scheduler/failed.html", func(c echo.Context) error {
		jobs, err := queries.ListFailedJobs(c.Request().Context())
		if err != nil {
			return c.Render(500, "error", err)
		}

		data := make(map[string]any)
		data["Failed"] = jobs

		return c.Render(http.StatusOK, "scheduler/failed.html", data)
	})

	e.GET("scheduler/completed.html", func(c echo.Context) error {
		jobs, err := queries.ListCompletedJobs(c.Request().Context())
		if err != nil {
			return c.Render(500, "error", err)
		}

		data := make(map[string]any)
		data["Completed"] = jobs

		return c.Render(http.StatusOK, "scheduler/completed.html", data)
	})

	// e.GET("run", func(c echo.Context) error {

	// 	jobs, _ := queries.ListSchedulerQueueJobs(c.Request().Context())

	// 	job := jobs[0]

	// 	// _, args, err := scheduler.GetProcessor(job.Processor, job.Arguments)

	// 	proc, err := scheduler.GetProcessor(job.Processor, job.Arguments, scheduling.Metadata{
	// 		Id:         job.ID,
	// 		Retry:      job.Retry,
	// 		RetryCount: job.RetryCount,
	// 		CreatedAt:  job.CreatedAt,
	// 	})

	// 	if err != nil {
	// 		return c.Render(http.StatusOK, "error", err)
	// 	}
	// 	proc.Run()

	// 	return c.Render(http.StatusOK, "error", nil)
	// })

	// e.GET("queue", func(c echo.Context) error {

	// 	queries.CreateSchedulerQueueJob(c.Request().Context(), repo.CreateSchedulerQueueJobParams{
	// 		ID:         uuid.New(),
	// 		Retry:      false,
	// 		RetryCount: 0,
	// 		CreatedAt:  time.Now(),
	// 		EnqueuedAt: time.Now(),
	// 		Arguments:  []byte(`{"test": "test","Greeting": "hello world"}`),
	// 		Processor:  scheduling.HelloWorldArguments{}.Kind(),
	// 	})

	// 	return c.Render(http.StatusOK, "error", "hello world")
	// })

	// e.GET("schedule2", func(c echo.Context) error {

	// 	performAt := time.Now().Add((8 * time.Second))

	// 	err := queue.EnqueueJob(c.Request().Context(), performAt, &scheduling.Job{
	// 		Id:         uuid.New(),
	// 		Retry:      true,
	// 		RetryCount: 0,
	// 		CreatedAt:  time.Now(),
	// 		Arguments:  scheduling.ErrorArguments{},
	// 	})

	// 	// queries.CreateInProgressSet(c.Request().Context(), repo.CreateInProgressSetParams{
	// 	// 	ID:         uuid.New(),
	// 	// 	Retry:      false,
	// 	// 	RetryCount: 0,
	// 	// 	CreatedAt:  time.Now(),
	// 	// 	EnqueuedAt: time.Now(),
	// 	// 	Processor:  "test",
	// 	// })
	// 	return c.Render(http.StatusOK, "error", err)
	// })

	e.GET("debug", func(c echo.Context) error {

		jobs, _ := queries.ListFailedJobs(c.Request().Context())

		for _, job := range jobs {
			fmt.Println(job.Log)
		}

		return c.Render(http.StatusOK, "error", nil)
	})

	// e.GET("schedule", func(c echo.Context) error {

	// 	performAt := time.Now().Add((5 * time.Second))

	// 	err := queue.EnqueueJob(c.Request().Context(), performAt, &scheduling.Job{
	// 		Id:         uuid.New(),
	// 		Retry:      false,
	// 		RetryCount: 0,
	// 		CreatedAt:  time.Now(),
	// 		Arguments: scheduling.HelloWorldArguments{
	// 			Greeting: "Hello, World!",
	// 		},
	// 	})

	// 	// queries.CreateInProgressSet(c.Request().Context(), repo.CreateInProgressSetParams{
	// 	// 	ID:         uuid.New(),
	// 	// 	Retry:      false,
	// 	// 	RetryCount: 0,
	// 	// 	CreatedAt:  time.Now(),
	// 	// 	EnqueuedAt: time.Now(),
	// 	// 	Processor:  "test",
	// 	// })
	// 	return c.Render(http.StatusOK, "error", err)
	// })

	go func() {
		if err := scheduler.Run(); err != nil {
			panic(fmt.Errorf("error running scheduler - %w", err))
		}
		fmt.Printf("scheduler stopped")
	}()
	go func() {
		if err := e.Start(":8956"); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatal("failed to start server - ", err)
		}
	}()

	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		// e.Logger.Fatal(err)
		e.Logger.Error(err)
	}

	if err := scheduler.Close(); err != nil {
		panic(fmt.Errorf("error closing scheduler - %w", err))
	}
}
