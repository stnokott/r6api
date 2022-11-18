package api

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/pkg/errors"
)

type ubiErrResp struct {
	Error     string      `json:"error"`
	ErrorCode json.Number `json:"errorCode"`
	Message   string      `json:"message"`
}

func request(r *http.Request, dst any) (err error) {
	r.Header.Add("User-Agent", "inofficial private non-commercial stats API (nfkottenhahn@web.de)")
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
		return err
	}
	errs := []string{errData.Message, errData.Error, errData.ErrorCode.String()}
	for _, errContent := range errs {
		if errContent != "" {
			return errors.New(errContent)
		}
	}
	return nil
}
