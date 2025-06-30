package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type malformedRequest struct {
	status int
	msg    string
}

func (mr *malformedRequest) Error() string {
	return mr.msg
}

// Parse a JSON object in request body into `dst`
// If encountered an internal error during parsing, this function will return a malformedRequest struct (which implments the error interface), which has the http response status and error messsage (string).
// Note that this function is meant to be helper middleware, so it only reads from the request, and doesn't write back a response.
// The considered parsing internal errors include:
// 1. checking for an incorrect Content-Type header would allow us to 'fail fast' if there is an unexpected content-type provided, and we can send the client a helpful error message without spending unnecessary resources on parsing the request body.
// 2. prevent our server resources being wasted if a malcious client sends a very large request body,
// 3. Disallow fields that are not present in the dst interface
// 4. The request may contain multiple json objects, and the json.Decoder.Decode(&dst) will only parse the first object, so we need to return a http.StatusBadRequest if there are multiple JSON objects. We can only check this after we've parsed the body once.
func DecodeJSONBody(w http.ResponseWriter, r *http.Request, dst any) error {
	// check "Content-Type: application/json;"
	contentType := r.Header.Get("Content-Type")
	if contentType != "" {
		mediaType := strings.ToLower(strings.TrimSpace(strings.Split(contentType, ";")[0]))
		if mediaType != "application/json" {
			msg := "Content-Type header is not application/json"
			http.Error(w, msg, http.StatusUnsupportedMediaType)
		}
	}

	// check request size
	r.Body = http.MaxBytesReader(w, r.Body, 262144)

	// decode and wrap the internal errors to return meaningful error messages
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	err := dec.Decode(&dst)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var maxBytesError *http.MaxBytesError

		switch {
		case errors.As(err, &syntaxError):
			msg := fmt.Sprintf("Request body contains badly-formed JSON at position %d", syntaxError.Offset)
			return &malformedRequest{http.StatusBadRequest, msg}

		case errors.As(err, &unmarshalTypeError):
			msg := fmt.Sprintf("Request body contains an invalid value for the %q field (at position %d)", unmarshalTypeError.Field, unmarshalTypeError.Offset)
			return &malformedRequest{http.StatusBadRequest, msg}

		case errors.As(err, &maxBytesError):
			msg := fmt.Sprintf("Request body must not be larger than %d bytes", maxBytesError.Limit)
			return &malformedRequest{http.StatusRequestEntityTooLarge, msg}

		case errors.Is(err, io.ErrUnexpectedEOF):
			msg := "Request body contains badly-formed JSON"
			return &malformedRequest{http.StatusBadRequest, msg}

		case errors.Is(err, io.EOF):
			msg := "Request body must not be empty"
			return &malformedRequest{http.StatusBadRequest, msg}
		default:
			return err
		}
	}

	// successful decoding of the first JSON object
	// now check if there are more than one JSON object by trying to parse again
	// if error is NOT io.EOF, then there are more.
	// we don't care what's in the next object, so pass an empty struct
	err = dec.Decode(&struct{}{})
	if !errors.Is(err, io.EOF) {
		msg := "Request body must contain 1 single JSON object"
		return &malformedRequest{http.StatusBadRequest, msg}
	}
	return nil
}
