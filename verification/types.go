package verification

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"runtime"
	"strings"

	"github.com/lhbelfanti/ditto/v2/http"
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

	// Request is one request of a journey. Route is "METHOD /path/{param}/v1" for a route of the app, or
	// "METHOD http://host/path" for another service; Params fills its {param} placeholders.
	Request struct {
		Route  string
		Params map[string]string
		Query  url.Values
		Token  string
		Header map[string]string
		Body   any
		Status int
	}

	// Step is a request a journey sent: its route, the status the journey expects and the status it got.
	// A setup step prepares the journey and is not part of its feature card.
	Step struct {
		Route    string   `json:"route"`
		Want     int      `json:"want"`
		Got      int      `json:"got"`
		Setup    bool     `json:"setup,omitempty"`
		Exchange Exchange `json:"exchange"`
	}

	// Exchange is the evidence of a step, with its tokens, codes and cookies redacted.
	Exchange struct {
		URL            string            `json:"url"`
		RequestHeader  map[string]string `json:"request_header,omitempty"`
		RequestBody    any               `json:"request_body,omitempty"`
		ResponseHeader map[string]string `json:"response_header,omitempty"`
		ResponseBody   any               `json:"response_body,omitempty"`
	}

	// Result is what a journey did: its steps, and why it stopped when it failed.
	Result struct {
		Journey string `json:"journey"`
		Steps   []Step `json:"steps"`
		Failure string `json:"failure,omitempty"`
	}

	// Report is the result of every journey of a drive, in the order they ran.
	Report struct {
		Results []Result `json:"results"`
	}

	// Run is the journey being proved: Prove calls its methods to send requests and check their answers.
	Run struct {
		ctx     context.Context
		client  http.Client
		baseURL string
		result  Result
	}
)

// Do sends request as a step of the journey, and stops the journey when the answer is not request.Status.
func (r *Run) Do(request Request) http.Response {
	return r.send(request, false)
}

// Send sends request to prepare the journey, such as a sign-in. It stops the journey only when request.Status
// is set and the answer is another one.
func (r *Run) Send(request Request) http.Response {
	return r.send(request, true)
}

// Decode reads the JSON body of response into target, and stops the journey when it can't.
func (r *Run) Decode(response http.Response, target any) {
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

// send sends request and records it as a step, setup or not.
func (r *Run) send(request Request, setup bool) http.Response {
	method, target, err := resolve(request, r.baseURL)
	if err != nil {
		r.fail(fmt.Errorf("%w: %w", ErrInvalidRequest, err))
	}

	outgoing := r.client.NewRequest(method, target)
	for key, value := range request.Header {
		outgoing.WithHeader(key, value)
	}

	if request.Token != "" {
		outgoing.WithHeader("Authorization", "Bearer "+request.Token)
	}

	if request.Body != nil {
		outgoing.WithBody(request.Body)
	}

	response, err := outgoing.Call(r.ctx)
	r.result.Steps = append(r.result.Steps, Step{
		Route:    request.Route,
		Want:     request.Status,
		Got:      response.StatusCode,
		Setup:    setup,
		Exchange: redact(target, request, response),
	})
	if err != nil {
		r.fail(fmt.Errorf("%w: %w", ErrCantSendRequest, err))
	}

	if (!setup || request.Status != 0) && response.StatusCode != request.Status {
		r.fail(fmt.Errorf("%w: %s answered %d, want %d", ErrUnexpectedStatus, request.Route, response.StatusCode, request.Status))
	}

	return response
}

// fail records err as the failure of the journey and stops it.
func (r *Run) fail(err error) {
	r.result.Failure = err.Error()
	runtime.Goexit()
}

// Failed reports whether a journey of the report failed.
func (r Report) Failed() bool {
	for _, result := range r.Results {
		if result.Failure != "" {
			return true
		}
	}

	return false
}

// Summary returns one line per journey, PASS with its step count or FAIL with its failure, and a last line with the totals.
func (r Report) Summary() string {
	var summary strings.Builder
	failed := 0
	for _, result := range r.Results {
		if result.Failure != "" {
			failed++
			fmt.Fprintf(&summary, "FAIL %s: %s\n", result.Journey, result.Failure)
			continue
		}

		fmt.Fprintf(&summary, "PASS %s (%d steps)\n", result.Journey, len(result.Steps))
	}

	fmt.Fprintf(&summary, "%d journeys, %d failed\n", len(r.Results), failed)

	return summary.String()
}
