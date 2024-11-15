package views

import (
	"embed"
	"io"
	"text/template"

	"github.com/labstack/echo/v4"
)

//go:embed *.html
//go:embed **/*.html
var views embed.FS

var (
// dashboard = parse("issued.html")
// newCertificate = parse("certificates/new.html")
// profileShow = parse("profile/show.html")
// profileEdit = parse("profile/edit.html")
// layout = "layout.html"
)

type DashboardParams struct {
	Title   string
	Message string
}

// func IssuedCertificates(w io.Writer, p IssuedCertificatesParams, view string) {
// 	return c.Render(http.StatusOK, "issued.html", nil)
// }

type IssuedCertificatesParams struct {
	Title   string
	Message string
}

// func Dashboard(w io.Writer, p DashboardParams, partial string) error {
// 	if partial == "" {
// 		partial = "layout.html"
// 	}
// 	return dashboard.ExecuteTemplate(w, partial, p)
// }

// func NewCertificate(w io.Writer, p NewCeritificateParams) error {
// 	return newCertificate.ExecuteTemplate(w, layout, p)
// }

type NewCeritificateParams struct {
	Title string
}

// type ProfileShowParams struct {
// 	Title   string
// 	Message string
// }

// func ProfileShow(w io.Writer, p ProfileShowParams, partial string) error {
// 	if partial == "" {
// 		partial = "layout.html"
// 	}
// 	return profileShow.ExecuteTemplate(w, partial, p)
// }

// type ProfileEditParams struct {
// 	Title   string
// 	Message string
// }

// func ProfileEdit(w io.Writer, p ProfileEditParams, partial string) error {
// 	if partial == "" {
// 		partial = "layout.html"
// 	}
// 	return profileEdit.ExecuteTemplate(w, partial, p)
// }

// func parse(file string) *template.Template {
// 	return template.Must(
// 		template.New("layout.html").ParseFS(files, "sidebar.html", "header.html", "layout.html", file))
// }

// var funcs template.FuncMap = template.FuncMap{
// 	"uppercase": func(v string) string {
// 		return strings.ToUpper(v)
// 	},
// }

type TemplateRegistry struct {
	// templates *template.Template
	templates map[string]*template.Template
}

func (t *TemplateRegistry) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	// return html.Dashboard(w, html.DashboardParams{}, "layout.html")
	return t.templates[name].ExecuteTemplate(w, "layout.html", data)
}

func (t *TemplateRegistry) Register(name string) {
	// t.templates[name] = template.Must()

	parsed, err := template.New(name).ParseFS(views, "sidebar.html", "header.html", "layout.html", name)

	if err != nil {
		panic(err)
	}

	t.templates[name] = parsed

	// template.New?
	// t.templates[name] = template.New("layout.html").ParseFS(files, "sidebar.html", "header.html", "layout.html", file))
}

func NewTemplates(views []string) *TemplateRegistry {
	// templates := make(map[string]*template.Template)

	templates := &TemplateRegistry{
		templates: make(map[string]*template.Template),
	}

	for _, view := range views {
		templates.Register(view)
	}

	return templates
	// return &Template{
	// 	map[string]*template.Template{
	// 		"issued.html": parse("issued.html"),
	// 	},
	// }
}
