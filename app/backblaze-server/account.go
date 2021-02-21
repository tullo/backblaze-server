package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/pkg/errors"
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
func authorizeAccount(app *app) (account, error) {
	// authorize account within 3 seconds.
	d := time.Now().Add(15 * time.Second)
	ctx, cancel := context.WithDeadline(context.Background(), d)
	defer cancel()

	var a account
	url := fmt.Sprintf("%s/b2api/v2/b2_authorize_account", app.api)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return a, errors.Wrap(err, "creating request")
	}

	req.SetBasicAuth(app.keyID, app.appKey)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return a, errors.Wrap(err, "processing request")
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return a, errors.New("authorizing account: response status: " + res.Status)
	}

	if err = json.NewDecoder(res.Body).Decode(&a); err != nil {
		return a, errors.Wrap(err, "decoding json")
	}

	return a, nil
}
