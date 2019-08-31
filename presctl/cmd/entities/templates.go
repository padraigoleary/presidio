package entities

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/viper"
)

const templateURLFormat = "http://%s:%s/api/v1/templates/%s/%s/%s"
const actionURLFormat = "http://%s:%s/api/v1/projects/%s/%s"

const (
	create = 0
	update = 1
	delete = 2
	get    = 3
)

type restOp int

// if error is not null print it and exit the program
func check(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// getBodyString returns the response's body
func getBodyString(response *http.Response) string {
	body, err := ioutil.ReadAll(response.Body)
	check(err)

	len := response.ContentLength
	bodyStr := string(body[:len])

	return bodyStr
}

// prettifyJSON makes a valid json string 'pretty' (human readable / indented)
func prettifyJSON(body string) string {
	jsonStr := &bytes.Buffer{}
	err := json.Indent(jsonStr, []byte(body), "", "  ")
	check(err)
	return jsonStr.String()
}

// prettyPrintJSON prints the given json in an indented manner
func prettyPrintJSON(jsonBody string) {
	jsonStr := prettifyJSON(jsonBody)
	fmt.Printf("Result: %s\n", jsonStr)
}

// saveToFile saves the given json in a indented manner to a file
func saveToFile(jsonBody string, outputFilePath string) {
	jsonStr := prettifyJSON(jsonBody)
	file, err := os.Create(outputFilePath)
	check(err)
	_, err = file.WriteString(jsonStr)
	defer file.Close()
	check(err)
	fmt.Printf("Template saved to: %s\n", outputFilePath)
}

// templateRestCommand sends a POST REST command to the Presidio instance in order to manage templates
func templateRestCommand(httpClient httpClient, op restOp, projectName string, actionName string, templateName string, fileContentStr string, outputFilePath string) {
	var ip = viper.GetString("presidio_ip")
	var port = viper.GetString("presidio_port")

	url := fmt.Sprintf(templateURLFormat,
		ip,
		port,
		projectName,
		actionName,
		templateName)

	var req *http.Request
	var err error
	switch op {
	case create:
		req, err = http.NewRequest("POST", url, strings.NewReader(fileContentStr))
	case update:
		req, err = http.NewRequest("PUT", url, strings.NewReader(fileContentStr))
	case delete:
		req, err = http.NewRequest("DELETE", url, nil)
	case get:
		req, err = http.NewRequest("GET", url, nil)
	}
	check(err)

	response, err := httpClient.Do(req)
	check(err)

	if response.StatusCode >= 300 {
		errMsg := fmt.Sprintf("Operation failed. Returned status code: %d",
			response.StatusCode)
		fmt.Println(errMsg)
		os.Exit(1)
	}

	if op != get {
		fmt.Println("Success")
		return
	}

	unquotedStr, err := strconv.Unquote(getBodyString(response))
	check(err)
	if outputFilePath != "" {
		saveToFile(unquotedStr, outputFilePath)
	} else {
		prettyPrintJSON(unquotedStr)
	}

	fmt.Println("Success")
}

// CreateTemplate creates a new template
func CreateTemplate(httpClient httpClient, projectName string, actionName string, templateName string, fileContentStr string) {
	templateRestCommand(httpClient, create, projectName, actionName, templateName, fileContentStr, "")
}

// UpdateTemplate updates an existing template
func UpdateTemplate(httpClient httpClient, projectName string, actionName string, templateName string, fileContentStr string) {
	templateRestCommand(httpClient, update, projectName, actionName, templateName, fileContentStr, "")
}

// DeleteTemplate deletes an existing template
func DeleteTemplate(httpClient httpClient, projectName string, actionName string, templateName string) {
	templateRestCommand(httpClient, delete, projectName, actionName, templateName, "", "")
}

// GetTemplate retrieved an existing template, can be logged or saved to a file
func GetTemplate(httpClient httpClient, projectName string, actionName string, templateName string, outputFilePath string) {
	templateRestCommand(httpClient, get, projectName, actionName, templateName, "", outputFilePath)
}
