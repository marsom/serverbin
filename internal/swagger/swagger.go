package swagger

import (
	"bytes"
	"embed"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"net/url"
	"path"

	"github.com/marsom/serverbin/internal/core"
	"github.com/marsom/serverbin/internal/httphandler"
	"gopkg.in/yaml.v3"
)

//go:embed dist
var swaggerAssets embed.FS

//go:embed api.yaml
var apiAssets string

//go:embed management-api.yaml
var managmentApiAssets string

func rootHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "" || r.URL.Path == "/" {
			if len(r.URL.Query()) == 0 {
				http.Redirect(w, r, "/?configUrl=%2Fswagger-config.yaml", http.StatusFound)
				return
			}
		}

		// serve swagger ui
		w.Header().Set("Access-Control-Allow-Origin", "*")

		next.ServeHTTP(w, r)
	})
}

func MustUiHandler() http.Handler {
	handlerAssets, err := fs.Sub(swaggerAssets, "dist")
	if err != nil {
		log.Panicf("could not create swagger handler: %s", err)
	}

	return rootHandler(http.FileServer(http.FS(handlerAssets)))
}

type apiTemplate struct {
	httphandler.Config
	Paths             []string
	BaseUrl           *url.URL
	ManagementBaseUrl *url.URL
}

type namedUrl struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type config struct {
	URL  string     `yaml:"url,omitempty"`
	URLs []namedUrl `yaml:"urls,omitempty"`
}

func ConfigHandler(relativePath string, configs ...httphandler.Config) http.HandlerFunc {
	swaggerConfig := config{
		URL:  "",
		URLs: nil,
	}

	for _, c := range configs {
		swaggerConfig.URLs = append(swaggerConfig.URLs, namedUrl{
			Name: "ServerBin API - " + c.Path,
			URL:  path.Join(c.Path, relativePath),
		})
	}

	swaggerConfig.URLs = append(swaggerConfig.URLs, namedUrl{
		Name: "Management API",
		URL:  "/management-api.yaml",
	})

	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		jsonWriter := yaml.NewEncoder(w)
		_ = jsonWriter.Encode(&swaggerConfig)
	}
}

func ManagementDefinitionHandler(configs ...httphandler.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data := newApiTemplate(configs, r)

		tmpl := template.Must(template.New("management-api.yaml").Parse(managmentApiAssets))

		var b bytes.Buffer
		if err := tmpl.Execute(&b, data); err != nil {
			w.WriteHeader(http.StatusInternalServerError)

			log.Printf("templated rendering failed: %s", err)
			return
		}

		yamlData := make(map[string]interface{})
		if err := yaml.Unmarshal(b.Bytes(), &yamlData); err != nil {
			w.WriteHeader(http.StatusInternalServerError)

			log.Printf("yaml redering failed: %s", err)
			return
		}

		w.WriteHeader(200)
		jsonWriter := yaml.NewEncoder(w)
		_ = jsonWriter.Encode(yamlData)
	}
}

func DefinitionHandler(configs ...httphandler.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data := newApiTemplate(configs, r)

		tmpl := template.Must(template.New("api.yaml").Parse(apiAssets))

		var b bytes.Buffer
		if err := tmpl.Execute(&b, data); err != nil {
			w.WriteHeader(http.StatusInternalServerError)

			log.Printf("templated rendering failed: %s", err)
			return
		}

		yamlData := make(map[string]interface{})
		if err := yaml.Unmarshal(b.Bytes(), &yamlData); err != nil {
			w.WriteHeader(http.StatusInternalServerError)

			log.Printf("yaml redering failed: %s", err)
			return
		}

		w.WriteHeader(200)
		jsonWriter := yaml.NewEncoder(w)
		_ = jsonWriter.Encode(yamlData)
	}
}

func newApiTemplate(configs []httphandler.Config, r *http.Request) apiTemplate {
	data := apiTemplate{
		Config: configs[0],
		Paths:  []string{},
	}

	for _, config := range configs {
		data.Paths = append(data.Paths, config.Path)
	}

	baseUrl, _ := core.FindBaseUrl(r, data.Server.BaseUrl, data.Server.TrustedAddresses)
	managementBaseUrl, _ := core.FindBaseUrl(r, data.Server.ManagementBaseUrl, data.Server.TrustedAddresses)

	data.BaseUrl = baseUrl
	data.ManagementBaseUrl = managementBaseUrl

	return data
}
