package pull

import (
	"encoding/json"
	"fmt"
	"log"
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

type manifestsBody struct {
	schemaVersion float64
	mediaType string
	config map[string]interface{}
	layers []map[string]interface{}
}

const (
	WWW_AUTHENTICATE = "Www-Authenticate"
	ACCEPT = "application/vnd.docker.distribution.manifest.v2+json"
	ACCEPT_LIST = "application/vnd.docker.distribution.manifest.list.v2+json"
	MANIFSTS_URL = "https://%s/v2/%s/manifests/%s"
)

func Pull(reg, repo, tag string) error {
	url := fmt.Sprintf("https://%s/v2/", reg)
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
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
	authToken, err := getAuthHead(authUrl, regService, repo)
	if err != nil {
		return err
	}
	err = getManifests(authToken, repo, reg, tag)
	if err != nil {
		return err
	}
	return nil
}

func getManifests(authToken, repo, reg, tag string) error {
	r := NewRequester()
	url := fmt.Sprintf(MANIFSTS_URL, reg, repo, tag)

	//resMap := &map[string]interface{}{}
	manifests := &manifestsBody{}
	resp, err := r.Get(url, manifests, map[string]string{
		"Authorization": authToken,
		"Accept": ACCEPT,
	},nil)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		log.Printf("[-] Cannot fetch manifest for %s [HTTP %v]\n", repo, resp.StatusCode)
		resp, err = r.Get(url, manifests, map[string]string{
			"Authorization": authToken,
			"Accept": ACCEPT_LIST,
		},nil)
		if resp.StatusCode == http.StatusOK {
			log.Printf("[+] Manifests found for this tag (use the @digest format to pull the corresponding image):")
			//m := *resMap
			//manifests := make(map[string]interface{})(m["manifests"])
			//for k, v := range manifests {
			//
			//}
		}

	}

	//for _, layer := range manifests.layers {
	//	ublob := layer["digest"]
	//}

	return nil
}

func getAuthHead(authUrl, regService, repo string) (string, error) {
	url := fmt.Sprintf("%s?service=%s&scope=repository:%s:pull", authUrl, regService, repo)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	token := &Token{}
	response, err := ReadJSONResponse(resp, token)
	if err != nil {
		return "", err
	}
	fmt.Println(response)
	return fmt.Sprintf("Bearer %s", token.Token), nil
}

func ReadJSONResponse(response *http.Response, responseStruct interface{}) (*http.Response, error) {
	defer response.Body.Close()
	err := json.NewDecoder(response.Body).Decode(responseStruct)
	if err != nil && err.Error() == "EOF" {
		return response, nil
	}
	return response, nil
}
