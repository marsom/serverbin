package swagger

import (
	"bytes"
	"html/template"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/marsom/serverbin/internal/httphandler"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestDefinitionHandler(t *testing.T) {
	baseUrl, err := url.Parse("http://localhost:8080")
	require.Nil(t, err)

	managementBaseUrl, err := url.Parse("http://localhost:8080")
	require.Nil(t, err)

	data := apiTemplate{
		Config: httphandler.Config{
			Path: "",
			Server: httphandler.Server{
				MaxRequestBody:    1024,
				BaseUrl:           baseUrl,
				ManagementBaseUrl: managementBaseUrl,
				TrustedAddresses:  nil,
			},
			Cookie: nil,
			Delay: &httphandler.Delay{
				MaxDuration: 10 * time.Second,
			},
			Slow:     nil,
			Redirect: nil,
		},
		Paths:             []string{"/"},
		BaseUrl:           baseUrl,
		ManagementBaseUrl: managementBaseUrl,
	}

	tmpl, err := template.New("api.yaml").Parse(apiAssets)
	require.Nil(t, err)
	require.NotNil(t, tmpl)

	var b bytes.Buffer
	err = tmpl.Execute(&b, data)
	require.Nil(t, err)

	yamlData := make(map[string]interface{})
	require.Nil(t, yaml.Unmarshal(b.Bytes(), &yamlData))

	jsonWriter := yaml.NewEncoder(os.Stdout)

	_ = jsonWriter.Encode(yamlData)

}
