package main

import (
	"Golang/Request"
	"Golang/Response"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"log"
	"net/http"
	"os"

	"github.com/Danny-Dasilva/CycleTLS/cycletls"
	"github.com/gorilla/mux"
)

func main() {
	port := "8000"
	if len(os.Args) > 1 {
		port = os.Args[1]
	}

	err := os.Setenv("tls13", "1")
	if err != nil {
		log.Println(err.Error())
	}

	router := mux.NewRouter()
	router.HandleFunc("/check-status", CheckStatus).Methods("GET")
	router.HandleFunc("/handle", Handle).Methods("POST")
	fmt.Println("The proxy server is running. Port: " + port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}

func CheckStatus(responseWriter http.ResponseWriter, request *http.Request) {
	responseWriter.Header().Set("Content-Type", "application/json")
	json.NewEncoder(responseWriter).Encode("good")
}

func Handle(responseWriter http.ResponseWriter, request *http.Request) {
	responseWriter.Header().Set("Content-Type", "application/json")

	var handleRequest Request.HandleRequest
	json.NewDecoder(request.Body).Decode(&handleRequest)
	client := cycletls.Init()

	var cookies []*http.Cookie
	for _, cookie := range handleRequest.Cookies {
		cookies = append(cookies, &http.Cookie{
			Name:     cookie.Name,
			Value:    cookie.Value,
			Path:     cookie.Path,
			Domain:   cookie.Domain,
			Expires:  cookie.Expires,
			MaxAge:   cookie.MaxAge,
			Secure:   cookie.Secure,
			HttpOnly: cookie.HTTPOnly,
		})
	}

	resp, err := client.Do(handleRequest.Url, cycletls.Options{
		InsecureSkipVerify: handleRequest.InsecureSkipVerify,
		Body:               handleRequest.Body,
		Proxy:              handleRequest.Proxy,
		Timeout:            handleRequest.Timeout,
		Headers:            handleRequest.Headers,
		Ja3:                handleRequest.Ja3,
		UserAgent:          handleRequest.UserAgent,
		DisableRedirect:    handleRequest.DisableRedirect,
	}, handleRequest.Method)

	var handleResponse Response.HandleResponse

	if err != nil {
		fmt.Println(err)
		handleResponse.Success = false
		handleResponse.Error = err.Error()
		json.NewEncoder(responseWriter).Encode(handleResponse)
		return
	}

	handleResponse.Success = true
	handleResponse.Payload = &Response.HandleResponsePayload{
		Text:    DecodeResponse(&resp),
		Headers: resp.Headers,
		Status:  resp.Status,
	}

	json.NewEncoder(responseWriter).Encode(handleResponse)
}

func DecodeResponse(response *cycletls.Response) string {
	switch response.Headers["Content-Encoding"] {
	case "gzip":
		reader, _ := gzip.NewReader(strings.NewReader(response.Body))
		defer reader.Close()
		readerResponse, _ := io.ReadAll(reader)
		return string(readerResponse)
	default:
		return response.Body
	}
}
