package request

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/pkg/errors"
	"github.com/stnokott/r6api/constants"
)

type ubiErrResp struct {
	Error     string      `json:"error"`
	ErrorCode json.Number `json:"errorCode"`
	Message   string      `json:"message"`
}

// Plain executes r with the default HTTP client and returns the plain body.
// Remember to close it after reading.
func Plain(r *http.Request) (io.ReadCloser, error) {
	r.Header.Add("User-Agent", constants.USER_AGENT)
	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

// JSON executes r with the default HTTP client and performs API-related processing such as deserialization and error-checking.
// If no errors occur, it attempts to unmarshal the response body into dst.
func JSON(r *http.Request, dst any) (err error) {
	r.Header.Add("User-Agent", constants.USER_AGENT)
	r.Header.Add("Accept", "application/json")
	var resp *http.Response
	resp, err = http.DefaultClient.Do(r)
	if err != nil {
		return
	}

	var data []byte
	data, err = io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	if err = checkForErrors(data); err != nil {
		if resp.StatusCode != 200 {
			err = errors.Wrapf(err, "unexpected status code %d", resp.StatusCode)
		}
		return
	}
	err = json.Unmarshal(data, dst)
	return
}

func checkForErrors(data []byte) error {
	var errData ubiErrResp
	if err := json.Unmarshal(data, &errData); err != nil {
		return errors.New("no further information available")
	}
	errs := []string{errData.Message, errData.Error, errData.ErrorCode.String()}
	for _, errContent := range errs {
		if errContent != "" {
			return errors.New(errContent)
		}
	}
	return nil
}
