package main

import (
	"net/http"
	"os"

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
		panic(err)
		// panic(m.Down())
	}

	e := echo.New()
	e.Use(middleware.Logger())

	e.Renderer = views.NewTemplates([]string{
		"issued.html",
		"certificates/new.html",
		"requests/list.html",
		"settings/api.html",
	})

	e.Static("/css", "css")

	e.GET("/", func(c echo.Context) error {
		return c.Render(http.StatusOK, "issued.html", nil)
	})

	e.GET("certificates/new.html", func(c echo.Context) error {
		return c.Render(http.StatusOK, "certificates/new.html", nil)
	})

	e.GET("requests/list.html", func(c echo.Context) error {
		return c.Render(http.StatusOK, "requests/list.html", nil)
	})

	e.GET("settings/api.html", func(c echo.Context) error {
		return c.Render(http.StatusOK, "settings/api.html", nil)
	})

	e.Logger.Fatal(e.Start(":8956"))
}
