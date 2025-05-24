package utils

import (
	"encoding/json"
	"io"
	"net/http"
)

type jsonData struct {
	Url        string      `json:"url"`
	StatusCode string      `json:"status_code"`
	Body       interface{} `json:"body"`
}

func NewJsonData(url, code string, body interface{}) jsonData {
	return jsonData{
		Url:        url,
		StatusCode: code,
		Body:       body,
	}
}

func ProcessResponse(resp *http.Response, showBody bool, limit int) (string, error) {
	defer resp.Body.Close()

	var body interface{}
	if showBody {
		rawBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", nil
		}
		var parsed interface{}
		if err := json.Unmarshal(rawBody, &parsed); err == nil {
			slc, ok := parsed.([]interface{})
			if ok && limit > 0 && limit < len(slc) {
				body = slc[:limit]
			} else {
				body = parsed
			}
		} else {
			body = string(rawBody)
		}
	}

	jsonResult := NewJsonData(resp.Request.URL.String(), resp.Status, body)
	prettyJson, err := json.MarshalIndent(jsonResult, "", " ")
	if err != nil {
		return "", err
	}

	return string(prettyJson) + "\n", nil
}
