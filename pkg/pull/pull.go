package pull

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type AuthToken struct {
	Authorization string `json:"Authorization"`
	Accept string `json:"Accept"`
}

type Token struct {
	Token string `json:"token"`
	AccessToken string `json:"access_token"`
}

const (
	WWW_AUTHENTICATE = "Www-Authenticate"
	ACCEPT = "application/vnd.docker.distribution.manifest.v2+json"
)

func Pull(reg, repo, tag string) error {
	url := fmt.Sprintf("https://%s/v2/", reg)
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	fmt.Println(resp)
	authUrl := "https://auth.docker.io/token"
	regService := "registry.docker.io"
	if resp.StatusCode == http.StatusUnauthorized {
		authHeader := strings.Split(resp.Header.Get(WWW_AUTHENTICATE), "\"")
		authUrl = authHeader[1]
		if len(authHeader) > 3 {
			regService = authHeader[3]
		} else {
			regService = ""
		}
	}
	getAuthHead(authUrl, regService, repo)
	return nil
}

func getAuthHead(authUrl, regService, repo string) (*AuthToken, error) {
	url := fmt.Sprintf("%s?service=%s&scope=repository:%s:pull", authUrl, regService, repo)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	token := &Token{}
	response, err := ReadJSONResponse(resp, token)
	if err != nil {
		return nil, err
	}
	fmt.Println(response)
	return &AuthToken{
		Authorization: fmt.Sprintf("Bearer %s", ),
	}, nil
}

func ReadJSONResponse(response *http.Response, responseStruct interface{}) (*http.Response, error) {
	defer response.Body.Close()
	err := json.NewDecoder(response.Body).Decode(responseStruct)
	if err != nil && err.Error() == "EOF" {
		return response, nil
	}
	return response, nil
}
