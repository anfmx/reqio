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
		var rawBody []byte
		if limit > 0 {
			rawBody = make([]byte, limit)
			n, _ := resp.Body.Read(rawBody)
			rawBody = rawBody[:n]
		} else {
			rawBody, _ = io.ReadAll(resp.Body)
		}
		var parsed interface{}
		if err := json.Unmarshal(rawBody, &parsed); err == nil {
			body = parsed
		} else {
			body = string(rawBody)
		}
		json.NewDecoder(resp.Body).Decode(&body)
	}

	jsonResult := NewJsonData(resp.Request.URL.String(), resp.Status, body)
	prettyJson, err := json.MarshalIndent(jsonResult, "", " ")
	if err != nil {
		return "", err
	}

	return string(prettyJson) + "\n", nil
}
