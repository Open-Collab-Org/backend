package utils

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/apex/log"
	"github.com/go-playground/validator/v10"
	"io"
	"net/http"
	"strconv"
)

// Read the request body as JSON and unmarshal it into `dto`.
// The request body is unmarshalled with json.Unmarshal.
func ReadJson(ctx context.Context, request *http.Request, dto interface{}) error {
	logger := log.FromContext(ctx)

	bodyBytes, err := ReadBody(request)
	if err != nil {
		logger.WithError(err).Error("Failed to read request body")
		return err
	}

	err = json.Unmarshal(bodyBytes, dto)
	if err != nil {
		logger.WithError(err).Info("Failed to unmarshal json")
		return err
	}

	validate := validator.New()
	err = validate.Struct(dto)
	if err != nil {
		return err
	}

	return nil
}

// Marshal a go object into JSON and send it as the response body. `data` is
// the data to be sent, it is marshaled  with json.Marshal and sent
// with http.ResponseWriter.Write.
func WriteJson(writer http.ResponseWriter, ctx context.Context, status int, data interface{}) error {
	logger := log.FromContext(ctx)

	bytes, err := json.Marshal(data)
	if err != nil {
		logger.WithError(err).Error("Failed to serialize JSON.")

		return err
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.Header().Set("Content-Length", strconv.Itoa(len(bytes)))
	writer.WriteHeader(status)
	_, err = writer.Write(bytes)
	if err != nil {
		logger.WithError(err).Error("Failed to write JSON response.")

		return err
	}

	return nil
}

// Get an int value from query parameter `param`.
// Returns the value of the parameter and whether it was set. If the parameter was
// not set, `def` is returned as the value.
//
// Note: if the parameter value is not an integer, it is treated as if the parameter
// was not set.
func IntFromQuery(request *http.Request, param string, def int) (int, bool) {
	values := request.URL.Query()[param]
	if len(values) < 1 {
		return def, false
	}

	val, err := strconv.Atoi(values[0])
	if err != nil {
		return def, false
	} else {
		return val, true
	}
}

// Read the request's body into a slice of bytes.
func ReadBody(r *http.Request) ([]byte, error) {

	bytes := make([]byte, 0)

	for {
		chunk := make([]byte, 2048)
		n, err := r.Body.Read(chunk)
		if err != nil {
			if !errors.Is(err, io.EOF) {
				return nil, err
			}
		}

		bytes = append(bytes, chunk[:n]...)

		if n < 1 {
			break
		}
	}

	return bytes, nil
}
