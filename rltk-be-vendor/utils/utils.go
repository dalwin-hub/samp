package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type HttpResp struct {
	Status bool        `json:"code"`
	Data   interface{} `json:"data"`
	Error  interface{} `json:"error"`
}

// JSON
func JSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.WriteHeader(statusCode)
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		fmt.Fprintf(w, "%s", err.Error())
	}
}

func SuccessResp(w http.ResponseWriter, statusCode int, data interface{}) {
	w.WriteHeader(statusCode)

	c := struct {
		Status bool        `json:"status"`
		Data   interface{} `json:"data"`
	}{Status: true, Data: data}
	b, err := json.MarshalIndent(&c, "", "\t")

	err = json.NewEncoder(w).Encode(json.RawMessage(string(b)))
	if err != nil {
		fmt.Fprintf(w, "%s", err.Error())
	}
}

func FailureResp(w http.ResponseWriter, statusCode int, failureMsg interface{}) {
	w.WriteHeader(statusCode)

	c := struct {
		Status bool        `json:"status"`
		Data   interface{} `json:"data"`
		Error  interface{} `json:"error"`
	}{Status: false, Data: nil, Error: failureMsg}
	b, err := json.MarshalIndent(&c, "", "\t")
	if err != nil {
		fmt.Fprintf(w, "%s", err.Error())
	}

	err = json.NewEncoder(w).Encode(json.RawMessage(string(b)))
	if err != nil {
		fmt.Fprintf(w, "%s", err.Error())
	}
}

func ERROR(w http.ResponseWriter, statusCode int, err error) {
	if err != nil {
		JSON(w, statusCode, struct {
			Error string `json:"error"`
		}{
			Error: err.Error(),
		})
		return
	}
	JSON(w, http.StatusBadRequest, nil)
}
