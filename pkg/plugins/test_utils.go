package plugins

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/spf13/afero"

	"github.com/stripe/stripe-cli/pkg/config"
)

// TestConfig Implementation out the GetConfigFolder function
type TestConfig struct {
	config.Config
}

// GetProfile returns the Mock Profile
func (c *TestConfig) GetProfile() *config.Profile {
	return &config.Profile{
		APIKey: "rk_test_11111111111111111111111111",
	}
}

// GetConfigFolder returns the absolute path for the TestConfig
func (c *TestConfig) GetConfigFolder(xdgPath string) string {
	return "/"
}

// InitConfig is not implemented
func (c *TestConfig) InitConfig() {}

// setUpFS Sets up a memMap that contains the manifest
func setUpFS() afero.Fs {
	// test plugin manifest
	// Note that only some of entries have actual checksums that match with what the test server returns.
	manifestContent, _ := os.ReadFile("./test_artifacts/plugins.toml")
	fs := afero.NewMemMapFs()
	// fs.Mkdir("test_config_folder", os.ModePerm)
	afero.WriteFile(fs, "/plugins.toml", manifestContent, os.ModePerm)
	return fs
}

// TestServers is a struct containing test servers that will be useful for unit testing plugin logic
type TestServers struct {
	ArtifactoryServer *httptest.Server
	StripeServer      *httptest.Server
}

// CloseAll calls Close() on each of the httptest servers.
func (ts *TestServers) CloseAll() {
	ts.ArtifactoryServer.Close()
	ts.StripeServer.Close()
}

// setUpServers sets up a local stripe server and artifactory server for unit tests
func setUpServers(t *testing.T, manifestContent []byte) TestServers {
	artifactoryServer := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		switch url := req.URL.String(); {
		case url == "/plugins.toml":
			res.Write(manifestContent)
		case strings.Contains(url, "/appA/2.0.1"):
			res.Write([]byte("hello, I am appA_2.0.1"))
		case strings.Contains(url, "/appA/1.0.1"):
			res.Write([]byte("hello, I am appA_1.0.1"))
		case strings.Contains(url, "/appA/0.0.1"):
			res.Write([]byte("hello, I am appA_0.0.1"))
		case strings.Contains(url, "/appA/0.0.0"):
			// Binary exists that is not in the manifest
			res.Write([]byte("hello, I am appA_0.0.0"))
		case strings.Contains(url, "/appB/1.2.1"):
			// Mismatching checksums
			res.Write([]byte("hello, I am appB_1.2.1"))
		default:
			t.Errorf("Received an unexpected request URL: %s", req.URL.String())
		}
	}))

	// The checksums in the test toml files are the same for each OS variation of the release for unit testing purposes
	stripeServer := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		switch url := req.URL.String(); url {
		case "/v1/stripecli/get-plugin-url":
			res.Write([]byte(fmt.Sprintf(`{"base_url": "%s"}`, artifactoryServer.URL)))
		default:
			t.Errorf("Received an unexpected request URL: %s", req.URL.String())
		}
	}))

	return TestServers{
		ArtifactoryServer: artifactoryServer,
		StripeServer:      stripeServer,
	}
}
