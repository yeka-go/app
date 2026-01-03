package ui

import (
	"bytes"
	"embed"
	_ "embed"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/yeka-go/app/cmd/goapp/internal/openapi"
)

/*
	Source:
	- https://cdn.jsdelivr.net/npm/redoc/bundles/redoc.standalone.js
	- https://unpkg.com/swagger-ui-dist@5.11.0/swagger-ui-bundle.js
	- https://unpkg.com/swagger-ui-dist@5.11.0/swagger-ui.css

*/

//go:embed assets/*
var fs embed.FS

type template string

const (
	SwaggerUITemplate template = "swagger-ui"
	RedoclyTemplate   template = "redocly"

	AssetRedoclyHtml string = "assets/redoc.html"
	AssetSwaggerHtml string = "assets/swagger-ui.html"
)

type Option struct {
	Template template
	Spec     []byte
	SpecName string
	BaseURL  string
}

func NewHandler(opt Option) http.Handler {
	m := http.NewServeMux()

	var title = "SwaggerUI"
	oapi, err := openapi.LoadFromBytes(opt.Spec)
	if err == nil {
		s, ok := oapi.GetPathAsString("info/title")
		if ok {
			title = s
		}
	}

	var template []byte
	if opt.Template == RedoclyTemplate {
		template, _ = fs.ReadFile(AssetRedoclyHtml)
	} else {
		template, _ = fs.ReadFile(AssetSwaggerHtml)
	}

	if opt.SpecName == "" {
		opt.SpecName = "swagger.json"
	} else {
		opt.SpecName = filepath.Base(opt.SpecName)
	}

	if !strings.HasPrefix(opt.BaseURL, "/") {
		opt.BaseURL = "/" + opt.BaseURL
	}
	if !strings.HasSuffix(opt.BaseURL, "/") {
		opt.BaseURL += "/"
	}

	m.HandleFunc(opt.BaseURL, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.NotFound(w, r)
			return
		}

		path := strings.TrimPrefix(r.URL.Path, strings.TrimSuffix(opt.BaseURL, "/"))

		switch path {
		case "/":
			var b []byte
			switch r.URL.Query().Get("template") {
			case "swagger", "swagger-ui":
				b, _ = fs.ReadFile(AssetSwaggerHtml)
			case "redoc", "redocly":
				b, _ = fs.ReadFile(AssetRedoclyHtml)
			default:
				b = template
			}
			if title != "" {
				b = bytes.ReplaceAll(b, []byte(`<title>SwaggerUI</title>`), []byte(`<title>`+title+`</title>`))
			}
			b = bytes.ReplaceAll(b, []byte(`"./swagger.json"`), []byte(`"./`+opt.SpecName+`"`))
			writeResponse(w, "text/html", b)

		case "/swagger-ui.css":
			b, _ := fs.ReadFile("assets/swagger-ui.css")
			writeResponse(w, "text/css", b)

		case "/swagger-ui-bundle.js", "/redoc.standalone.js":
			b, _ := fs.ReadFile("assets" + path)
			writeResponse(w, "text/javascript", b)

		case "/" + opt.SpecName:
			writeResponse(w, "text/plain", opt.Spec)

		default:
			http.NotFound(w, r)
		}
	})

	return m
}

func writeResponse(w http.ResponseWriter, contentType string, b []byte) {
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(http.StatusOK)
	w.Write(b)

}
