package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/packago/config"
)

// Account holds backblaze specific data.
type account struct {
	AbsoluteMinimumPartSize int    `json:"absoluteMinimumPartSize"`
	AccountID               string `json:"accountId"`
	Allowed                 struct {
		BucketID     string      `json:"bucketId"`
		BucketName   string      `json:"bucketName"`
		Capabilities []string    `json:"capabilities"`
		NamePrefix   interface{} `json:"namePrefix"`
	} `json:"allowed"`
	APIURL              string `json:"apiUrl"`
	AuthorizationToken  string `json:"authorizationToken"`
	DownloadURL         string `json:"downloadUrl"`
	RecommendedPartSize int    `json:"recommendedPartSize"`
}

// AuthorizeAccount retrieves backblaze account data connected to a
// configured keyID and applicationKey
func authorizeAccount() (account, error) {
	var a account
	api := config.File().GetString("backblaze.rootUrl")
	url := fmt.Sprintf("%s/b2api/v2/b2_authorize_account", api)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return a, err
	}

	keyID := config.File().GetString("backblaze.keyID")
	appKey := config.File().GetString("backblaze.applicationKey")
	req.SetBasicAuth(keyID, appKey)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return a, err
	}
	defer resp.Body.Close()

	if err = json.NewDecoder(resp.Body).Decode(&a); err != nil {
		return a, err
	}
	return a, nil
}
