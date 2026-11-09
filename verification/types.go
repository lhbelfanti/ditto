package verification

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"runtime"
	"strings"
	"sync"

	_http "github.com/lhbelfanti/ditto/v2/http"
)

type (
	// Journey is one use case the verification proves end to end. Setup prepares what its steps need, such as
	// sessions, and Steps lists its requests in the order they are sent. Steps is also called without Setup, to
	// write the table of its feature card, so it only reads what Setup left and never sends anything itself.
	Journey interface {
		Setup(verifier *Verifier)
		Steps() []Step
	}

	// Entry registers a journey under the ID its feature card gives it.
	Entry struct {
		ID      string
		Journey Journey
	}

	// Drive proves every journey in order and reports the steps each one took.
	Drive func(ctx context.Context, entries []Entry) Report

	// Step is one request of a journey and what it must answer. Path is the route pattern, with each {name} filled
	// from PathValues when the step is sent; a full URL, of another service, is sent as it is. With changes the
	// request with what only an earlier step knows, such as its cookies. A step without Expect checks only Status.
	Step struct {
		Description string
		Method      string
		Path        string
		PathValues  Values
		Query       map[string]string
		Bearer      string
		Body        any
		With        With
		Status      int
		Expect      Expect
	}

	// Values returns the path values of a step when it is sent, so that they can come from an earlier step.
	Values func() map[string]string

	// With changes the request of a step before it is sent.
	With func(request *_http.Request)

	// Expect checks the response of a step, beyond its status.
	Expect func(response _http.Response) error

	// Checker checks the data of a response, decoded as T, and keeps what a later step needs.
	Checker[T any] func(data T) error

	// Record is a request a journey sent: the step it proves, the status the journey expects, the status it got and
	// the exchange the connection carried. A setup record prepares the journey and is not part of its feature card.
	Record struct {
		Description string   `json:"description,omitempty"`
		Want        int      `json:"want"`
		Got         int      `json:"got"`
		Setup       bool     `json:"setup,omitempty"`
		Exchange    Exchange `json:"exchange"`
	}

	// Exchange is the evidence of a step as it went through the connection, with its tokens, codes and cookies
	// redacted.
	Exchange struct {
		Method         string            `json:"method"`
		URL            string            `json:"url"`
		RequestHeader  map[string]string `json:"request_header,omitempty"`
		RequestBody    any               `json:"request_body,omitempty"`
		ResponseHeader map[string]string `json:"response_header,omitempty"`
		ResponseBody   any               `json:"response_body,omitempty"`
	}

	// Result is what a journey did: the requests it sent, and why it stopped when it failed.
	Result struct {
		Journey string
		Records []Record
		Failure error
	}

	// Report is the result of every journey of a drive, in the order they ran.
	Report struct {
		Results []Result
	}

	// Verifier sends the requests of the journey being proved. Setup, and the helpers it calls, send theirs with
	// Send; the steps are sent by the drive.
	Verifier struct {
		ctx      context.Context
		client   *_http.CustomClient
		recorder *recorder
		baseURL  string
		result   Result
	}

	// reportJSON is the report as report.json holds it.
	reportJSON struct {
		Results []resultJSON `json:"results"`
	}

	// resultJSON is the result of a journey in report.json, with its failure as a message.
	resultJSON struct {
		Journey string   `json:"journey"`
		Records []Record `json:"records"`
		Failure string   `json:"failure,omitempty"`
	}

	// recorder is the transport of a Verifier: it passes each request to next and keeps the exchange of the last
	// one.
	recorder struct {
		mu       sync.Mutex
		next     http.RoundTripper
		exchange Exchange
	}
)

// Request returns the method and the route of the step, with its query, as its feature card shows them.
func (s Step) Request() string {
	request := s.Method + " " + s.Path
	if len(s.Query) == 0 {
		return request
	}

	query := make(url.Values, len(s.Query))
	for key, value := range s.Query {
		query.Set(key, value)
	}

	return request + "?" + query.Encode()
}

// path returns Path with each {name} replaced by its path value.
func (s Step) path() string {
	if s.PathValues == nil {
		return s.Path
	}

	path := s.Path
	for name, value := range s.PathValues() {
		path = strings.ReplaceAll(path, "{"+name+"}", value)
	}

	return path
}

// URL returns the URL of path on the app: a path that starts with "/" goes to the base URL of the instance, and a
// full URL, of another service, is returned as it is.
func (v *Verifier) URL(path string) string {
	if strings.HasPrefix(path, "/") {
		return strings.TrimSuffix(v.baseURL, "/") + path
	}

	return path
}

// NewRequest creates a request sent through the connection of the journey, so that its record keeps the exchange.
func (v *Verifier) NewRequest(method, url string) *_http.Request {
	return v.client.NewRequest(method, url)
}

// Send sends request to prepare the journey, such as a sign-in, and stops the journey when the answer is not
// status. Its record is marked as a setup record.
func (v *Verifier) Send(request *_http.Request, status int) _http.Response {
	return v.send(request, status, true, "")
}

// Decode reads the JSON body of response into target, and stops the journey when it can't.
func (v *Verifier) Decode(response _http.Response, target any) {
	err := json.Unmarshal([]byte(response.Body), target)
	if err != nil {
		v.fail(fmt.Errorf("%w: %w", ErrCantDecodeBody, err))
	}
}

// prove sends step, and stops the journey when its answer is not its status or its Expect fails.
func (v *Verifier) prove(step Step) {
	request := v.NewRequest(step.Method, v.URL(step.path()))
	for key, value := range step.Query {
		request.WithQueryParam(key, value)
	}

	if step.Bearer != "" {
		request.WithBearer(step.Bearer)
	}

	if step.Body != nil {
		request.WithBody(step.Body)
	}

	if step.With != nil {
		step.With(request)
	}

	response := v.send(request, step.Status, false, step.Description)
	if step.Expect == nil {
		return
	}

	err := step.Expect(response)
	if err != nil {
		v.fail(fmt.Errorf("%w: %s: %w", ErrUnexpectedResponse, step.Description, err))
	}
}

// send sends request and records it, setup or not, with the description of its step and the exchange the recorder
// kept.
func (v *Verifier) send(request *_http.Request, status int, setup bool, description string) _http.Response {
	response, err := request.Call(v.ctx)
	record := Record{Description: description, Want: status, Got: response.StatusCode, Setup: setup, Exchange: v.recorder.take()}
	v.result.Records = append(v.result.Records, record)
	if err != nil {
		v.fail(fmt.Errorf("%w: %w", ErrCantSendRequest, err))
	}

	if response.StatusCode != status {
		exchange := record.Exchange
		v.fail(fmt.Errorf("%w: %s %s answered %d, want %d", ErrUnexpectedStatus, exchange.Method, exchange.URL, response.StatusCode, status))
	}

	return response
}

// fail records err as the failure of the journey and stops it.
func (v *Verifier) fail(err error) {
	v.result.Failure = err
	runtime.Goexit()
}

// RoundTrip sends req through next, and keeps its exchange, redacted, before it returns the response.
func (r *recorder) RoundTrip(req *http.Request) (*http.Response, error) {
	requestBody := requestBodyOf(req)

	resp, err := r.next.RoundTrip(req)
	if err != nil {
		r.keep(redact(req, requestBody, nil, nil))
		return nil, err
	}

	responseBody := responseBodyOf(resp)
	r.keep(redact(req, requestBody, resp, responseBody))

	return resp, nil
}

// keep stores exchange as the last one.
func (r *recorder) keep(exchange Exchange) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.exchange = exchange
}

// take returns the last exchange and forgets it, so that a request sent outside the recorder gets none.
func (r *recorder) take() Exchange {
	r.mu.Lock()
	defer r.mu.Unlock()

	exchange := r.exchange
	r.exchange = Exchange{}

	return exchange
}

// Failed reports whether a journey of the report failed.
func (r Report) Failed() bool {
	for _, result := range r.Results {
		if result.Failure != nil {
			return true
		}
	}

	return false
}

// Summary returns one line per journey, PASS with its request count or FAIL with its failure, and a last line with the
// totals.
func (r Report) Summary() string {
	var summary strings.Builder
	failed := 0
	for _, result := range r.Results {
		if result.Failure != nil {
			failed++
			fmt.Fprintf(&summary, "FAIL %s: %s\n", result.Journey, result.Failure)
			continue
		}

		fmt.Fprintf(&summary, "PASS %s (%d requests)\n", result.Journey, len(result.Records))
	}

	fmt.Fprintf(&summary, "%d journeys, %d failed\n", len(r.Results), failed)

	return summary.String()
}

// MarshalJSON writes the report as report.json holds it, with each failure as its message.
func (r Report) MarshalJSON() ([]byte, error) {
	results := make([]resultJSON, 0, len(r.Results))
	for _, result := range r.Results {
		written := resultJSON{Journey: result.Journey, Records: result.Records}
		if result.Failure != nil {
			written.Failure = result.Failure.Error()
		}

		results = append(results, written)
	}

	return json.Marshal(reportJSON{Results: results})
}
