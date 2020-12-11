package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"regexp"
	"strings"

	"github.com/parnurzeal/gorequest"
)

// Client facilitates making HTTP requests to the Karma API
type Client struct {
	http *gorequest.SuperAgent
	host string
}

// HTTPError is an error returned by a failed API call
type HTTPError struct {
	StatusCode int
	RequestURL *url.URL
	Message    string
}

func (err HTTPError) Error() string {
	if msgs := strings.SplitN(err.Message, "\n", 2); len(msgs) > 1 {
		return fmt.Sprintf("HTTP %d: %s (%s)\n%s", err.StatusCode, msgs[0], err.RequestURL, msgs[1])
	} else if err.Message != "" {
		return fmt.Sprintf("HTTP %d: %s (%s)", err.StatusCode, err.Message, err.RequestURL)
	}
	return fmt.Sprintf("HTTP %d (%s)", err.StatusCode, err.RequestURL)
}

// NewClient returns an authenticated client
func NewClient(host string, token string, teamID string, version string) Client {
	authorization := fmt.Sprintf("token %s", token)

	return Client{
		http: gorequest.New().
			Set("Authorization", authorization).
			Set("X-Team-Id", teamID).
			Set("X-CLI-Version", version),
		host: host,
	}
}

// Post makes a POST request to server
func (c Client) Post(endpoint string, payload interface{}, data interface{}) error {
	resp, _, errs := c.http.Clone().
		Post(c.host + endpoint).
		Send(payload).
		EndStruct(&data)

	if len(errs) != 0 {
		return errs[0]
	}

	success := resp.StatusCode >= 200 && resp.StatusCode < 300

	if !success {
		return HandleHTTPError(resp)
	}

	return nil
}

// Get makes a GET request to API server and assign response to data struct
func (c Client) Get(endpoint string, data interface{}) error {
	resp, _, errs := c.http.Clone().
		Get(c.host + endpoint).
		EndStruct(&data)

	if len(errs) != 0 {
		return errs[0]
	}

	success := resp.StatusCode >= 200 && resp.StatusCode < 300

	if !success {
		return HandleHTTPError(resp)
	}

	return nil
}

// HandleHTTPError catches HTTP errors and prints them out
func HandleHTTPError(resp gorequest.Response) error {
	httpError := HTTPError{
		StatusCode: resp.StatusCode,
		RequestURL: resp.Request.URL,
	}

	if !jsonTypeRE.MatchString(resp.Header.Get("Content-Type")) {
		httpError.Message = resp.Status
		return httpError
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		httpError.Message = err.Error()
		return httpError
	}

	var parsedBody struct {
		Message string `json:"message"`
		Errors  []json.RawMessage
	}

	if err := json.Unmarshal(body, &parsedBody); err != nil {
		return httpError
	}

	type errorObject struct {
		Message  string
		Resource string
		Field    string
		Code     string
	}

	messages := []string{parsedBody.Message}
	for _, raw := range parsedBody.Errors {
		switch raw[0] {
		case '"':
			var errString string
			_ = json.Unmarshal(raw, &errString)
			messages = append(messages, errString)
		case '{':
			var errInfo errorObject
			_ = json.Unmarshal(raw, &errInfo)
			msg := errInfo.Message
			if errInfo.Code != "custom" {
				msg = fmt.Sprintf("%s.%s %s", errInfo.Resource, errInfo.Field, errorCodeToMessage(errInfo.Code))
			}
			if msg != "" {
				messages = append(messages, msg)
			}
		}
	}
	httpError.Message = strings.Join(messages, "\n")

	return httpError
}

func errorCodeToMessage(code string) string {
	// https://docs.github.com/en/rest/overview/resources-in-the-rest-api#client-errors
	switch code {
	case "missing", "missing_field":
		return "is missing"
	case "invalid", "unprocessable":
		return "is invalid"
	case "already_exists":
		return "already exists"
	default:
		return code
	}
}

var jsonTypeRE = regexp.MustCompile(`[/+]json($|;)`)
