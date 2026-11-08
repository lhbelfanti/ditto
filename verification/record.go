package verification

import (
	"bytes"
	"io"
	"net/http"

	_http "github.com/lhbelfanti/ditto/v2/http"
)

// newRecordingClient returns a copy of client whose requests go through a recorder, and the recorder. A client
// without a transport sends through http.DefaultTransport, as net/http does.
func newRecordingClient(client *_http.CustomClient) (*_http.CustomClient, *recorder) {
	next := client.HTTPClient.Transport
	if next == nil {
		next = http.DefaultTransport
	}

	recording := &recorder{next: next}
	httpClient := *client.HTTPClient
	httpClient.Transport = recording

	return &_http.CustomClient{HTTPClient: &httpClient}, recording
}

// requestBodyOf returns the body req carries, read from a copy so that the request still sends it.
func requestBodyOf(req *http.Request) []byte {
	if req.GetBody == nil {
		return nil
	}

	body, err := req.GetBody()
	if err != nil {
		return nil
	}

	data, _ := io.ReadAll(body)

	return data
}

// responseBodyOf reads the body of resp and puts it back, so that the caller still reads it whole.
func responseBodyOf(resp *http.Response) []byte {
	if resp.Body == nil {
		return nil
	}

	data, _ := io.ReadAll(resp.Body)
	resp.Body = io.NopCloser(bytes.NewReader(data))

	return data
}
