package verification

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"sync"

	_http "github.com/lhbelfanti/ditto/v2/http"
)

type (
	// Journey is one use case the verification proves end to end, named by the ID its feature card gives it.
	Journey struct {
		ID    string
		Prove Prove
	}

	// Prove sends the requests of a journey through run; the first failed step stops it.
	Prove func(run *Run)

	// Drive proves every journey in order and reports the steps each one took.
	Drive func(ctx context.Context, journeys []Journey) Report

	// Step is a request a journey sent: the status the journey expects, the status it got and the exchange the
	// connection carried. A setup step prepares the journey and is not part of its feature card.
	Step struct {
		Want     int      `json:"want"`
		Got      int      `json:"got"`
		Setup    bool     `json:"setup,omitempty"`
		Exchange Exchange `json:"exchange"`
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

	// Result is what a journey did: its steps, and why it stopped when it failed.
	Result struct {
		Journey string
		Steps   []Step
		Failure error
	}

	// Report is the result of every journey of a drive, in the order they ran.
	Report struct {
		Results []Result
	}

	// Run is the journey being proved: Prove builds its requests with NewRequest and sends them with Do and Setup.
	Run struct {
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
		Journey string `json:"journey"`
		Steps   []Step `json:"steps"`
		Failure string `json:"failure,omitempty"`
	}

	// recorder is the transport of a Run: it passes each request to next and keeps the exchange of the last one.
	recorder struct {
		mu       sync.Mutex
		next     http.RoundTripper
		exchange Exchange
	}
)

// URL returns the URL of path on the app: a path that starts with "/" goes to the base URL of the instance, and a
// full URL, of another service, is returned as it is.
func (r *Run) URL(path string) string {
	if strings.HasPrefix(path, "/") {
		return strings.TrimSuffix(r.baseURL, "/") + path
	}

	return path
}

// NewRequest creates a request sent through the connection of the journey, so that its step keeps the exchange.
func (r *Run) NewRequest(method, url string) *_http.Request {
	return r.client.NewRequest(method, url)
}

// Do sends request as a step of the journey, and stops the journey when the answer is not status.
func (r *Run) Do(request *_http.Request, status int) _http.Response {
	return r.send(request, status, false)
}

// Setup sends request to prepare the journey, such as a sign-in, and stops the journey when the answer is not
// status. Its step is marked as a setup step.
func (r *Run) Setup(request *_http.Request, status int) _http.Response {
	return r.send(request, status, true)
}

// Decode reads the JSON body of response into target, and stops the journey when it can't.
func (r *Run) Decode(response _http.Response, target any) {
	err := json.Unmarshal([]byte(response.Body), target)
	if err != nil {
		r.fail(fmt.Errorf("%w: %w", ErrCantDecodeBody, err))
	}
}

// Require stops the journey with reason when condition is false.
func (r *Run) Require(condition bool, reason string) {
	if !condition {
		r.fail(fmt.Errorf("%w: %s", ErrRequirementNotMet, reason))
	}
}

// send sends request and records it as a step, setup or not, with the exchange the recorder kept.
func (r *Run) send(request *_http.Request, status int, setup bool) _http.Response {
	response, err := request.Call(r.ctx)
	step := Step{Want: status, Got: response.StatusCode, Setup: setup, Exchange: r.recorder.take()}
	r.result.Steps = append(r.result.Steps, step)
	if err != nil {
		r.fail(fmt.Errorf("%w: %w", ErrCantSendRequest, err))
	}

	if response.StatusCode != status {
		exchange := step.Exchange
		r.fail(fmt.Errorf("%w: %s %s answered %d, want %d", ErrUnexpectedStatus, exchange.Method, exchange.URL, response.StatusCode, status))
	}

	return response
}

// fail records err as the failure of the journey and stops it.
func (r *Run) fail(err error) {
	r.result.Failure = err
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

// Summary returns one line per journey, PASS with its step count or FAIL with its failure, and a last line with the
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

		fmt.Fprintf(&summary, "PASS %s (%d steps)\n", result.Journey, len(result.Steps))
	}

	fmt.Fprintf(&summary, "%d journeys, %d failed\n", len(r.Results), failed)

	return summary.String()
}

// MarshalJSON writes the report as report.json holds it, with each failure as its message.
func (r Report) MarshalJSON() ([]byte, error) {
	results := make([]resultJSON, 0, len(r.Results))
	for _, result := range r.Results {
		written := resultJSON{Journey: result.Journey, Steps: result.Steps}
		if result.Failure != nil {
			written.Failure = result.Failure.Error()
		}

		results = append(results, written)
	}

	return json.Marshal(reportJSON{Results: results})
}
