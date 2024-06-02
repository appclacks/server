package http_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"log/slog"

	"github.com/appclacks/go-client"
	"github.com/appclacks/server/config"
	"github.com/appclacks/server/internal/database"
	apihttp "github.com/appclacks/server/internal/http"
	"github.com/appclacks/server/internal/http/handlers"
	"github.com/appclacks/server/pkg/healthcheck"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

func toJson(t *testing.T, s any) []byte {
	t.Helper()
	result, err := json.Marshal(s)
	assert.NoError(t, err, "fail to marshal to json")
	return result
}

func fromJson(t *testing.T, s any, data []byte) {
	t.Helper()
	err := json.Unmarshal(data, s)
	assert.NoError(t, err, "fail to unmarshal to json data %s", string(data))
}

func readBody(t *testing.T, body io.ReadCloser) []byte {
	b, err := io.ReadAll(body)
	defer body.Close()
	assert.NoError(t, err)
	return b
}

// func basicAuth(username, password string) string {
// 	return fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", username, password))))
// }

type testCase struct {
	url            string
	expectedStatus int
	method         string
	payload        any
	headers        map[string]string
	form           map[string]string
	body           string
}

var baseURL = "http://127.0.0.1:10000"
var httpClient = &http.Client{
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	},
}

func testHTTP(t *testing.T, c testCase, result any) {
	t.Helper()
	var reqBody io.Reader
	if c.payload != nil {
		reqBody = bytes.NewBuffer(toJson(t, c.payload))
	}
	if c.form != nil {
		form := url.Values{}
		for k, v := range c.form {
			form.Add(k, v)
		}
		reqBody = strings.NewReader(form.Encode())
	}
	request, err := http.NewRequest(
		c.method,
		fmt.Sprintf("%s%s", baseURL, c.url),
		reqBody)
	assert.NoError(t, err)
	if c.payload != nil {
		request.Header.Set("Content-Type", "application/json; charset=UTF-8")
	}
	if c.form != nil {
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	for k, v := range c.headers {
		request.Header.Set(k, v)
	}
	response, err := httpClient.Do(request)
	assert.NoError(t, err)
	body := readBody(t, response.Body)
	assert.Equal(t, response.StatusCode, c.expectedStatus, string(body))
	if result != nil {
		fromJson(t, result, body)
	}
	if c.body != "" {
		assert.Contains(t, string(body), c.body)
	}
}

func TestIntegration(t *testing.T) {
	reg := prometheus.NewRegistry()
	config := config.Configuration{
		Database: database.Configuration{
			Migrations: "../../dev/migrations",
			Username:   "appclacks",
			Password:   "appclacks",
			Database:   "appclacks",
			Host:       "127.0.0.1",
			Port:       5432,
			SSLMode:    "disable",
		},
		HTTP: apihttp.Configuration{
			Host: "127.0.0.1",
			Port: 10000,
		},
		Healthchecks: config.Healthchecks{
			Probers: 1,
		},
	}
	logger := slog.Default()
	store, err := database.New(logger, config.Database, config.Healthchecks.Probers)
	healthcheckService := healthcheck.New(logger, store)
	assert.NoError(t, err)
	handlersBuilder := handlers.NewBuilder(healthcheckService)
	server, err := apihttp.NewServer(logger, config.HTTP, reg, handlersBuilder)
	assert.NoError(t, err)
	_, err = store.Exec("truncate healthcheck cascade;")
	assert.NoError(t, err)

	err = server.Start()
	assert.NoError(t, err)
	time.Sleep(1 * time.Second)

	// healthchecks

	dnsInput := client.CreateDNSHealthcheckInput{
		Timeout:     "3s",
		Name:        "dns1",
		Description: "toto",
		Enabled:     true,
		Labels: map[string]string{
			"foo": "bar",
		},
		Interval: "100s",
		HealthcheckDNSDefinition: client.HealthcheckDNSDefinition{
			Domain: "mcorbin.fr",
		},
	}
	createDNSCase := testCase{
		url:            "/api/v1/healthcheck/dns",
		expectedStatus: 200,
		payload:        dnsInput,
		method:         "POST",
	}

	dnsResult := client.Healthcheck{}

	testHTTP(t, createDNSCase, &dnsResult)
	assert.Equal(t, dnsInput.Name, dnsResult.Name)
	assert.Equal(t, true, dnsResult.Enabled)
	assert.NotEqual(t, "", dnsResult.ID)

	// list

	listHealthcheckCase := testCase{
		url:            "/api/v1/healthcheck",
		expectedStatus: 200,
		method:         "GET",
	}

	listHealthcheckResult := client.ListHealthchecksOutput{}
	testHTTP(t, listHealthcheckCase, &listHealthcheckResult)
	assert.Equal(t, 1, len(listHealthcheckResult.Result))
	assert.Equal(t, dnsResult.ID, listHealthcheckResult.Result[0].ID)

	listHealthcheckCaseRegex := testCase{
		url:            "/api/v1/healthcheck?name-pattern=trololo",
		expectedStatus: 200,
		method:         "GET",
	}

	testHTTP(t, listHealthcheckCaseRegex, &listHealthcheckResult)
	assert.Equal(t, 0, len(listHealthcheckResult.Result))

	listHealthcheckCaseRegexMatch := testCase{
		url:            "/api/v1/healthcheck?name-pattern=dns",
		expectedStatus: 200,
		method:         "GET",
	}

	testHTTP(t, listHealthcheckCaseRegexMatch, &listHealthcheckResult)
	assert.Equal(t, 1, len(listHealthcheckResult.Result))

	// get

	getHealthcheckCase := testCase{
		url:            fmt.Sprintf("/api/v1/healthcheck/%s", dnsResult.ID),
		expectedStatus: 200,
		method:         "GET",
	}

	getHealthcheckResult := client.Healthcheck{}
	testHTTP(t, getHealthcheckCase, &getHealthcheckResult)
	assert.Equal(t, dnsResult.ID, getHealthcheckResult.ID)
	assert.Equal(t, dnsResult.Name, getHealthcheckResult.Name)

	// update

	dnsUpdateInput := client.UpdateDNSHealthcheckInput{
		Timeout:     "10s",
		Name:        "dns2",
		Description: "toto",
		Enabled:     true,
		Labels: map[string]string{
			"foo": "bar",
		},
		Interval: "100s",
		HealthcheckDNSDefinition: client.HealthcheckDNSDefinition{
			Domain: "mcorbin.fr",
		},
	}
	updateDNSCase := testCase{
		url:            fmt.Sprintf("/api/v1/healthcheck/dns/%s", dnsResult.ID),
		expectedStatus: 200,
		payload:        dnsUpdateInput,
		method:         "PUT",
	}

	dnsResult = client.Healthcheck{}

	testHTTP(t, updateDNSCase, &dnsResult)
	assert.Equal(t, dnsResult.Name, dnsUpdateInput.Name)
	assert.Equal(t, dnsResult.Timeout, dnsUpdateInput.Timeout)
	assert.Equal(t, true, dnsResult.Enabled)
	assert.NotEqual(t, "", dnsResult.ID)

	// delete

	deleteHealthcheckCase := testCase{
		url:            fmt.Sprintf("/api/v1/healthcheck/%s", dnsResult.ID),
		expectedStatus: 200,
		method:         "DELETE",
	}

	testHTTP(t, deleteHealthcheckCase, nil)

	getHealthcheckCase = testCase{
		url:            fmt.Sprintf("/api/v1/healthcheck/%s", dnsResult.ID),
		expectedStatus: 404,
		method:         "GET",
	}
	testHTTP(t, getHealthcheckCase, nil)

}
