// Copyright 2015 Vadim Kravcenko
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

package pull

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// Request Methods

type Requester struct {
	Client *http.Client
	connControl chan struct{}
}

type APIRequest struct {
	Method   string
	Url      string
	Payload  io.Reader
	Headers  http.Header
	Suffix   string
}

func NewRequester() *Requester {
	return &Requester{
		Client: &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
		connControl: make(chan struct{}),
	}
}

func (ar *APIRequest) SetHeader(key string, value string) *APIRequest {
	ar.Headers.Set(key, value)
	return ar
}

func NewAPIRequest(method, url string, payload io.Reader) *APIRequest {
	var headers = http.Header{}
	var suffix string
	ar := &APIRequest{method, url, payload, headers, suffix}
	return ar
}

func (r *Requester) PostJSON(url string, payload io.Reader, responseStruct interface{}, querystring map[string]string) (*http.Response, error) {
	ar := NewAPIRequest("POST", url, payload)
	ar.SetHeader("Content-Type", "application/x-www-form-urlencoded")
	ar.Suffix = "api/json"
	return r.Do(ar, responseStruct, querystring)
}

func (r *Requester) Post(url string, payload io.Reader, responseStruct interface{}, querystring map[string]string) (*http.Response, error) {
	ar := NewAPIRequest("POST", url, payload)
	ar.SetHeader("Content-Type", "application/x-www-form-urlencoded")
	ar.Suffix = ""
	return r.Do(ar, responseStruct, querystring)
}
func (r *Requester) PostForm(url string, payload io.Reader, responseStruct interface{}, formString map[string]string) (*http.Response, error) {
	ar := NewAPIRequest("POST", url, payload)
	ar.SetHeader("Content-Type", "application/x-www-form-urlencoded")
	ar.Suffix = ""
	return r.DoPostForm(ar, responseStruct, formString)
}

func (r *Requester) PostFiles(url string, payload io.Reader, responseStruct interface{}, querystring map[string]string, files []string) (*http.Response, error) {
	ar := NewAPIRequest("POST", url, payload)
	return r.Do(ar, responseStruct, querystring, files)
}

func (r *Requester) PostXML(url string, xml string, responseStruct interface{}, querystring map[string]string) (*http.Response, error) {
	payload := bytes.NewBuffer([]byte(xml))
	ar := NewAPIRequest("POST", url, payload)
	ar.SetHeader("Content-Type", "application/xml;charset=utf-8")
	ar.Suffix = ""
	return r.Do(ar, responseStruct, querystring)
}

func (r *Requester) GetJSON(url string, responseStruct interface{}, query map[string]string) (*http.Response, error) {
	ar := NewAPIRequest("GET", url, nil)
	ar.SetHeader("Content-Type", "application/json")
	ar.Suffix = "api/json"
	return r.Do(ar, responseStruct, query)
}

func (r *Requester) GetXML(url string, responseStruct interface{}, query map[string]string) (*http.Response, error) {
	ar := NewAPIRequest("GET", url, nil)
	ar.SetHeader("Content-Type", "application/xml")
	ar.Suffix = ""
	return r.Do(ar, responseStruct, query)
}

func (r *Requester) Get(url string, responseStruct interface{}, header, querystring map[string]string) (*http.Response, error) {
	ar := NewAPIRequest("GET", url, nil)
	for k, v := range header {
		ar.SetHeader(k, v)
	}
	ar.Suffix = ""
	return r.Do(ar, responseStruct, querystring)
}

func (r *Requester) GetHtml(url string, responseStruct interface{}, querystring map[string]string) (*http.Response, error) {
	ar := NewAPIRequest("GET", url, nil)
	ar.Suffix = ""
	return r.DoGet(ar, responseStruct, querystring)
}

func (r *Requester) SetClient(client *http.Client) *Requester {
	r.Client = client
	return r
}

func (r *Requester) DoGet(ar *APIRequest, responseStruct interface{}, options ...interface{}) (*http.Response, error) {
	fileUpload := false
	var files []string
	URL, err := url.Parse(ar.Url + ar.Suffix)

	if err != nil {
		return nil, err
	}

	for _, o := range options {
		switch v := o.(type) {
		case map[string]string:

			querystring := make(url.Values)
			for key, val := range v {
				querystring.Set(key, val)
			}

			URL.RawQuery = querystring.Encode()
		case []string:
			fileUpload = true
			files = v
		}
	}
	var req *http.Request
	if fileUpload {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		for _, file := range files {
			fileData, err := os.Open(file)
			if err != nil {
				return nil, err
			}

			part, err := writer.CreateFormFile("file", filepath.Base(file))
			if err != nil {
				return nil, err
			}
			if _, err = io.Copy(part, fileData); err != nil {
				return nil, err
			}
			defer fileData.Close()
		}
		var params map[string]string
		json.NewDecoder(ar.Payload).Decode(&params)
		for key, val := range params {
			if err = writer.WriteField(key, val); err != nil {
				return nil, err
			}
		}
		if err = writer.Close(); err != nil {
			return nil, err
		}
		req, err = http.NewRequest(ar.Method, URL.String(), body)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", writer.FormDataContentType())
	} else {
		req, err = http.NewRequest(ar.Method, URL.String(), ar.Payload)
		if err != nil {
			return nil, err
		}
	}

	req.Close = true
	req.Header.Add("Accept", "*/*")
	for k := range ar.Headers {
		req.Header.Add(k, ar.Headers.Get(k))
	}
	r.connControl <- struct{}{}
	if response, err := r.Client.Do(req); err != nil {
		<-r.connControl
		return nil, err
	} else {
		<-r.connControl
		errorText := response.Header.Get("X-Error")
		if errorText != "" {
			return nil, errors.New(errorText)
		}
		err := CheckResponse(response)
		if err != nil {
			return nil, err
		}
		switch responseStruct.(type) {
		case *string:
			return r.ReadRawResponse(response, responseStruct)
		default:
			return r.ReadJSONResponse(response, responseStruct)
		}

	}

}

func (r *Requester) Do(ar *APIRequest, responseStruct interface{}, options ...interface{}) (*http.Response, error) {

	fileUpload := false
	var files []string
	URL, err := url.Parse(ar.Url + ar.Suffix)

	if err != nil {
		return nil, err
	}

	for _, o := range options {
		switch v := o.(type) {
		case map[string]string:

			querystring := make(url.Values)
			for key, val := range v {
				querystring.Set(key, val)
			}

			URL.RawQuery = querystring.Encode()
		case []string:
			fileUpload = true
			files = v
		}
	}
	var req *http.Request
	if fileUpload {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		for _, file := range files {
			fileData, err := os.Open(file)
			if err != nil {
				return nil, err
			}

			part, err := writer.CreateFormFile("file", filepath.Base(file))
			if err != nil {
				return nil, err
			}
			if _, err = io.Copy(part, fileData); err != nil {
				return nil, err
			}
			defer fileData.Close()
		}
		var params map[string]string
		json.NewDecoder(ar.Payload).Decode(&params)
		for key, val := range params {
			if err = writer.WriteField(key, val); err != nil {
				return nil, err
			}
		}
		if err = writer.Close(); err != nil {
			return nil, err
		}
		req, err = http.NewRequest(ar.Method, URL.String(), body)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", writer.FormDataContentType())
	} else {
		req, err = http.NewRequest(ar.Method, URL.String(), ar.Payload)
		if err != nil {
			return nil, err
		}
	}

	req.Close = true
	req.Header.Add("Accept", "*/*")
	for k := range ar.Headers {
		req.Header.Add(k, ar.Headers.Get(k))
	}
	//r.connControl <- struct{}{}
	if response, err := r.Client.Do(req); err != nil {
		//<-r.connControl
		return nil, err
	} else {
		//<-r.connControl
		errorText := response.Header.Get("X-Error")
		if errorText != "" {
			return nil, errors.New(errorText)
		}
		err := CheckResponse(response)
		if err != nil {
			return nil, err
		}
		switch responseStruct.(type) {
		case *string:
			return r.ReadRawResponse(response, responseStruct)
		case *map[string]interface{}:
			return r.ReadJSONResponse(response, responseStruct)
		default:
			return r.ReadJSONResponse(response, responseStruct)
		}

	}

}

func (r *Requester) DoPostForm(ar *APIRequest, responseStruct interface{}, form map[string]string) (*http.Response, error) {

	URL, err := url.Parse(ar.Url + ar.Suffix)

	if err != nil {
		return nil, err
	}
	formValue := make(url.Values)
	for k, v := range form {
		formValue.Set(k, v)
	}
	req, err := http.NewRequest("POST", URL.String(), strings.NewReader(formValue.Encode()))

	req.Close = true
	req.Header.Add("Accept", "*/*")
	for k := range ar.Headers {
		req.Header.Add(k, ar.Headers.Get(k))
	}
	r.connControl <- struct{}{}
	if response, err := r.Client.Do(req); err != nil {
		<-r.connControl
		return nil, err
	} else {
		<-r.connControl
		errorText := response.Header.Get("X-Error")
		if errorText != "" {
			return nil, errors.New(errorText)
		}
		err := CheckResponse(response)
		if err != nil {
			return nil, err
		}
		switch responseStruct.(type) {
		case *string:
			return r.ReadRawResponse(response, responseStruct)
		default:
			return r.ReadJSONResponse(response, responseStruct)
		}

	}
}

func (r *Requester) ReadRawResponse(response *http.Response, responseStruct interface{}) (*http.Response, error) {
	defer response.Body.Close()

	content, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	if str, ok := responseStruct.(*string); ok {
		*str = string(content)
	} else {
		return nil, fmt.Errorf("Could not cast responseStruct to *string")
	}

	return response, nil
}

func (r *Requester) ReadJSONResponse(response *http.Response, responseStruct interface{}) (*http.Response, error) {
	defer response.Body.Close()
	err := json.NewDecoder(response.Body).Decode(responseStruct)
	if err != nil && err.Error() == "EOF" {
		return response, nil
	}
	return response, nil
}

func CheckResponse(r *http.Response) error {

	switch r.StatusCode {
	case http.StatusOK, http.StatusCreated, http.StatusAccepted, http.StatusNoContent, http.StatusFound, http.StatusNotModified:
		return nil
	}
	defer r.Body.Close()
	return fmt.Errorf("error response")
}
