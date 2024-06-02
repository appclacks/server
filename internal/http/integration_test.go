package http_test

import (
	"bytes"
	"encoding/base64"
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

func basicAuth(username, password string) string {
	return fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", username, password))))
}

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
	assert.Equal(t, c.expectedStatus, response.StatusCode, string(body))
	if result != nil {
		fromJson(t, result, body)
	}
	if c.body != "" {
		assert.Contains(t, string(body), c.body)
	}
}

func TestIntegration(t *testing.T) {
	testUser := "testuser"
	testPassword := "testPassword"
	metricsUser := "metricsuser"
	metricsPassword := "metricsPassword"

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
			BasicAuth: apihttp.BasicAuth{
				Username: testUser,
				Password: testPassword,
			},
			Metrics: apihttp.Metrics{
				BasicAuth: apihttp.BasicAuth{
					Username: metricsUser,
					Password: metricsPassword,
				},
			},
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
		headers: map[string]string{
			"Authorization": basicAuth(testUser, testPassword),
		},
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
		headers: map[string]string{
			"Authorization": basicAuth(testUser, testPassword),
		},
	}

	listHealthcheckResult := client.ListHealthchecksOutput{}
	testHTTP(t, listHealthcheckCase, &listHealthcheckResult)
	assert.Equal(t, 1, len(listHealthcheckResult.Result))
	assert.Equal(t, dnsResult.ID, listHealthcheckResult.Result[0].ID)

	listHealthcheckCaseRegex := testCase{
		url:            "/api/v1/healthcheck?name-pattern=trololo",
		expectedStatus: 200,
		method:         "GET",
		headers: map[string]string{
			"Authorization": basicAuth(testUser, testPassword),
		},
	}

	testHTTP(t, listHealthcheckCaseRegex, &listHealthcheckResult)
	assert.Equal(t, 0, len(listHealthcheckResult.Result))

	listHealthcheckCaseRegexMatch := testCase{
		url:            "/api/v1/healthcheck?name-pattern=dns",
		expectedStatus: 200,
		method:         "GET",
		headers: map[string]string{
			"Authorization": basicAuth(testUser, testPassword),
		},
	}

	testHTTP(t, listHealthcheckCaseRegexMatch, &listHealthcheckResult)
	assert.Equal(t, 1, len(listHealthcheckResult.Result))

	// get

	getHealthcheckCase := testCase{
		url:            fmt.Sprintf("/api/v1/healthcheck/%s", dnsResult.ID),
		expectedStatus: 200,
		method:         "GET",
		headers: map[string]string{
			"Authorization": basicAuth(testUser, testPassword),
		},
	}

	getHealthcheckResult := client.Healthcheck{}
	testHTTP(t, getHealthcheckCase, &getHealthcheckResult)
	assert.Equal(t, dnsResult.ID, getHealthcheckResult.ID)
	assert.Equal(t, dnsResult.Name, getHealthcheckResult.Name)

	getByNameHealthcheckCase := testCase{
		url:            fmt.Sprintf("/api/v1/healthcheck/%s", dnsResult.Name),
		expectedStatus: 200,
		method:         "GET",
		headers: map[string]string{
			"Authorization": basicAuth(testUser, testPassword),
		},
	}

	getByNameHealthcheckResult := client.Healthcheck{}
	testHTTP(t, getByNameHealthcheckCase, &getByNameHealthcheckResult)
	assert.Equal(t, dnsResult.ID, getByNameHealthcheckResult.ID)
	assert.Equal(t, dnsResult.Name, getByNameHealthcheckResult.Name)

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
		headers: map[string]string{
			"Authorization": basicAuth(testUser, testPassword),
		},
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
		headers: map[string]string{
			"Authorization": basicAuth(testUser, testPassword),
		},
	}

	testHTTP(t, deleteHealthcheckCase, nil)

	getHealthcheckCase = testCase{
		url:            fmt.Sprintf("/api/v1/healthcheck/%s", dnsResult.ID),
		expectedStatus: 404,
		method:         "GET",
		headers: map[string]string{
			"Authorization": basicAuth(testUser, testPassword),
		},
	}
	testHTTP(t, getHealthcheckCase, nil)

	// more checks types

	// tcp

	tcpInput := client.CreateTCPHealthcheckInput{
		Timeout:     "3s",
		Name:        "tcp3",
		Description: "toto",
		Enabled:     true,
		Labels: map[string]string{
			"foo": "bar",
		},
		Interval: "100s",
		HealthcheckTCPDefinition: client.HealthcheckTCPDefinition{
			Target: "mcorbin.fr",
			Port:   443,
		},
	}

	createTCPCheckCase := testCase{
		url:            "/api/v1/healthcheck/tcp",
		expectedStatus: 200,
		payload:        tcpInput,
		method:         "POST",
		headers: map[string]string{
			"Authorization": basicAuth(testUser, testPassword),
		},
	}
	tcpResult := client.Healthcheck{}
	testHTTP(t, createTCPCheckCase, &tcpResult)
	assert.Equal(t, tcpInput.Name, tcpResult.Name)
	assert.Equal(t, true, tcpResult.Enabled)
	assert.NotEqual(t, "", tcpResult.ID)

	tcpUpdateInput := client.UpdateTCPHealthcheckInput{
		ID:          tcpResult.ID,
		Timeout:     "3s",
		Name:        "tcp4",
		Description: "toto",
		Enabled:     false,
		Labels: map[string]string{
			"foo": "bar",
		},
		Interval: "100s",
		HealthcheckTCPDefinition: client.HealthcheckTCPDefinition{
			Target: "mcorbin.fr",
			Port:   443,
		},
	}

	updateTCPCheckCase := testCase{
		url:            fmt.Sprintf("/api/v1/healthcheck/tcp/%s", tcpResult.ID),
		expectedStatus: 200,
		payload:        tcpUpdateInput,
		method:         "PUT",
		headers: map[string]string{
			"Authorization": basicAuth(testUser, testPassword),
		},
	}
	tcpUpdateResult := client.Healthcheck{}
	testHTTP(t, updateTCPCheckCase, &tcpUpdateResult)
	assert.Equal(t, tcpUpdateInput.Name, tcpUpdateResult.Name)
	assert.Equal(t, false, tcpUpdateResult.Enabled)
	assert.NotEqual(t, "", tcpUpdateResult.ID)

	// tls

	tlsInput := client.CreateTLSHealthcheckInput{
		Timeout:     "3s",
		Name:        "tls3",
		Description: "toto",
		Enabled:     true,
		Labels: map[string]string{
			"foo": "bar",
		},
		Interval: "100s",
		HealthcheckTLSDefinition: client.HealthcheckTLSDefinition{
			Target: "mcorbin.fr",
			Port:   443,
		},
	}

	createTLSCheckCase := testCase{
		url:            "/api/v1/healthcheck/tls",
		expectedStatus: 200,
		payload:        tlsInput,
		method:         "POST",
		headers: map[string]string{
			"Authorization": basicAuth(testUser, testPassword),
		},
	}
	tlsResult := client.Healthcheck{}
	testHTTP(t, createTLSCheckCase, &tlsResult)
	assert.Equal(t, tlsInput.Name, tlsResult.Name)
	assert.Equal(t, true, tlsResult.Enabled)
	assert.NotEqual(t, "", tlsResult.ID)

	tlsUpdateInput := client.UpdateTLSHealthcheckInput{
		ID:          tlsResult.ID,
		Timeout:     "3s",
		Name:        "tls4",
		Description: "toto",
		Enabled:     false,
		Labels: map[string]string{
			"foo": "bar",
		},
		Interval: "100s",
		HealthcheckTLSDefinition: client.HealthcheckTLSDefinition{
			Target: "mcorbin.fr",
			Port:   443,
		},
	}

	updateTLSCheckCase := testCase{
		url:            fmt.Sprintf("/api/v1/healthcheck/tls/%s", tlsResult.ID),
		expectedStatus: 200,
		payload:        tlsUpdateInput,
		method:         "PUT",
		headers: map[string]string{
			"Authorization": basicAuth(testUser, testPassword),
		},
	}
	tlsUpdateResult := client.Healthcheck{}
	testHTTP(t, updateTLSCheckCase, &tlsUpdateResult)
	assert.Equal(t, tlsUpdateInput.Name, tlsUpdateResult.Name)
	assert.Equal(t, false, tlsUpdateResult.Enabled)
	assert.NotEqual(t, "", tlsUpdateResult.ID)

	// http

	httpInput := client.CreateHTTPHealthcheckInput{
		Timeout:     "3s",
		Name:        "http3",
		Description: "toto",
		Enabled:     true,
		Labels: map[string]string{
			"foo": "bar",
		},
		Interval: "100s",
		HealthcheckHTTPDefinition: client.HealthcheckHTTPDefinition{
			Target:      "mcorbin.fr",
			Port:        443,
			ValidStatus: []uint{200, 201},
			Protocol:    "http",
			Method:      "GET",
		},
	}

	createHTTPCheckCase := testCase{
		url:            "/api/v1/healthcheck/http",
		expectedStatus: 200,
		payload:        httpInput,
		method:         "POST",
		headers: map[string]string{
			"Authorization": basicAuth(testUser, testPassword),
		},
	}
	httpResult := client.Healthcheck{}
	testHTTP(t, createHTTPCheckCase, &httpResult)
	assert.Equal(t, httpInput.Name, httpResult.Name)
	assert.Equal(t, true, httpResult.Enabled)
	assert.NotEqual(t, "", httpResult.ID)

	httpUpdateInput := client.UpdateHTTPHealthcheckInput{
		ID:          httpResult.ID,
		Timeout:     "3s",
		Name:        "http4",
		Description: "toto",
		Enabled:     false,
		Labels: map[string]string{
			"foo": "bar",
		},
		Interval: "100s",
		HealthcheckHTTPDefinition: client.HealthcheckHTTPDefinition{
			Target:      "mcorbin.fr",
			Port:        443,
			ValidStatus: []uint{200, 201},
			Protocol:    "http",
			Method:      "GET",
		},
	}

	updateHTTPCheckCase := testCase{
		url:            fmt.Sprintf("/api/v1/healthcheck/http/%s", httpResult.ID),
		expectedStatus: 200,
		payload:        httpUpdateInput,
		method:         "PUT",
		headers: map[string]string{
			"Authorization": basicAuth(testUser, testPassword),
		},
	}
	httpUpdateResult := client.Healthcheck{}
	testHTTP(t, updateHTTPCheckCase, &httpUpdateResult)
	assert.Equal(t, httpUpdateInput.Name, httpUpdateResult.Name)
	assert.Equal(t, false, httpUpdateResult.Enabled)
	assert.NotEqual(t, "", httpUpdateResult.ID)

	// command

	commandInput := client.CreateCommandHealthcheckInput{
		Timeout:     "3s",
		Name:        "command3",
		Description: "toto",
		Enabled:     true,
		Labels: map[string]string{
			"foo": "bar",
		},
		Interval: "100s",
		HealthcheckCommandDefinition: client.HealthcheckCommandDefinition{
			Command:   "ls",
			Arguments: []string{"-l"},
		},
	}

	createCommandCheckCase := testCase{
		url:            "/api/v1/healthcheck/command",
		expectedStatus: 200,
		payload:        commandInput,
		method:         "POST",
		headers: map[string]string{
			"Authorization": basicAuth(testUser, testPassword),
		},
	}
	commandResult := client.Healthcheck{}
	testHTTP(t, createCommandCheckCase, &commandResult)
	assert.Equal(t, commandInput.Name, commandResult.Name)
	assert.Equal(t, true, commandResult.Enabled)
	assert.NotEqual(t, "", commandResult.ID)

	commandUpdateInput := client.UpdateCommandHealthcheckInput{
		ID:          commandResult.ID,
		Timeout:     "3s",
		Name:        "command4",
		Description: "toto",
		Enabled:     false,
		Labels: map[string]string{
			"foo": "bar",
		},
		Interval: "100s",
		HealthcheckCommandDefinition: client.HealthcheckCommandDefinition{
			Command:   "ls",
			Arguments: []string{"-la"},
		},
	}

	updateCommandCheckCase := testCase{
		url:            fmt.Sprintf("/api/v1/healthcheck/command/%s", commandResult.ID),
		expectedStatus: 200,
		payload:        commandUpdateInput,
		method:         "PUT",
		headers: map[string]string{
			"Authorization": basicAuth(testUser, testPassword),
		},
	}
	commandUpdateResult := client.Healthcheck{}
	testHTTP(t, updateCommandCheckCase, &commandUpdateResult)
	assert.Equal(t, commandUpdateInput.Name, commandUpdateResult.Name)
	assert.Equal(t, false, commandUpdateResult.Enabled)
	assert.NotEqual(t, "", commandUpdateResult.ID)

	// metrics

	getHealthcheckCase = testCase{
		url:            "/metrics",
		expectedStatus: 200,
		method:         "GET",
		body:           "http_request",
		headers: map[string]string{
			"Authorization": basicAuth(metricsUser, metricsPassword),
		},
	}
	testHTTP(t, getHealthcheckCase, nil)

	cases := []testCase{
		{
			url:            "/healthz",
			expectedStatus: 200,
			method:         "GET",
		},
		{
			url:            "/metrics",
			expectedStatus: 401,
			method:         "GET",
		},
		{
			url:            "/metrics",
			expectedStatus: 401,
			method:         "GET",
			headers: map[string]string{
				"Authorization": basicAuth(testUser, testPassword),
			},
		},
		{
			url:            "/metrics",
			expectedStatus: 401,
			method:         "GET",
			headers: map[string]string{
				"Authorization": basicAuth(metricsUser, "invalid_pass"),
			},
		},
		{
			url:            "/api/v1/healthchecks",
			expectedStatus: 401,
			method:         "GET",
			headers: map[string]string{
				"Authorization": basicAuth(metricsUser, "invalid_pass"),
			},
		},
		{
			url:            "/api/v1/healthchecks",
			expectedStatus: 401,
			method:         "GET",
			headers: map[string]string{
				"Authorization": basicAuth("invalid_user", metricsPassword),
			},
		},
		{
			url:            "/api/v1/healthcheck",
			expectedStatus: 401,
			method:         "GET",
		},
		{
			url:            "/api/v1/healthcheck",
			expectedStatus: 401,
			method:         "GET",
			headers: map[string]string{
				"Authorization": basicAuth(testUser, "invalid_pass"),
			},
		},
		{
			url:            "/api/v1/healthcheck",
			expectedStatus: 401,
			method:         "GET",
			headers: map[string]string{
				"Authorization": basicAuth("invalid_user", testPassword),
			},
		},
		{
			url:            "/api/v1/healthcheck/aozjiji827YH82",
			expectedStatus: 404,
			body:           "healthcheck not found",
			method:         "GET",
			headers: map[string]string{
				"Authorization": basicAuth(testUser, testPassword),
			},
		},
	}
	for _, c := range cases {
		testHTTP(t, c, nil)
	}

}
