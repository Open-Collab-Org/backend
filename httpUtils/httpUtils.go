package httpUtils

import (
	"context"
	"encoding/json"
	"github.com/apex/log"
	"net/http"
	"strconv"
)

// Read the request body as JSON and unmarshal it into `dto`.
// The request body is unmarshalled with json.Unmarshal.
func ReadJson(request *http.Request, dto interface{}) error {
	bodyBytes, err := ReadBody(request)
	if err != nil {
		return err
	}

	err = json.Unmarshal(bodyBytes, dto)
	if err != nil {
		return err
	}

	return nil
}

// Marshal a go object into JSON and send it as the response body. `data` is
// the data to be sent, it is marshaled  with json.Marshal and sent
// with http.ResponseWriter.Write.
func WriteJson(writer http.ResponseWriter, ctx context.Context, data interface{}) error {
	logger := log.FromContext(ctx)

	bytes, err := json.Marshal(data)
	if err != nil {
		logger.WithError(err).Error("Failed to serialize JSON.")

		return err
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.Header().Set("Content-Length", strconv.Itoa(len(bytes)))
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
			return nil, err
		}

		bytes = append(bytes, chunk[:n]...)

		if n < 1 {
			break
		}
	}

	return bytes, nil
}
