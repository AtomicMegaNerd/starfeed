package freshrss

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/charmbracelet/log"
)

type FreshRSSFeedPublisher struct {
	baseUrl   string
	user      string
	apiToken  string // WARNING: Do not log this value as it is a secret
	authToken string // WARNING: Do not log this value as it is a secret
	client    *http.Client
	logger    *log.Logger
}

func NewFreshRSSFeedPublisher(
	baseUrl string,
	user string,
	apiToken string,
	client *http.Client,
	logger *log.Logger,
) *FreshRSSFeedPublisher {
	return &FreshRSSFeedPublisher{
		baseUrl:  baseUrl,
		user:     user,
		apiToken: apiToken,
		client:   client,
		logger:   logger,
	}
}

func (f *FreshRSSFeedPublisher) Authenticate() error {
	logger := f.logger

	reqUrl := fmt.Sprintf("%s/api/greader.php/accounts/ClientLogin", f.baseUrl)
	logger.Debugf("Authenticating with FreshRSS at %s", reqUrl)

	formData := url.Values{
		"Email":  {f.user},
		"Passwd": {f.apiToken},
	}

	data, err := f.doApiRequest(reqUrl, []byte(formData.Encode()), false)
	if err != nil {
		return err
	}

	err = f.parsePlainTextAuthResponse(data)
	if err != nil {
		return err
	}

	logger.Info("Authenticated with FreshRSS")
	return nil
}

func (f *FreshRSSFeedPublisher) AddFeed(feed, name, category string) error {
	logger := f.logger

	addUrl := fmt.Sprintf(
		"%s/api/greader.php/reader/api/0/subscription/quickadd", f.baseUrl,
	)

	formData := url.Values{
		"quickadd": {feed},
	}

	logger.Debugf("Adding feed to FreshRSS %s: %s", addUrl, formData.Encode())

	res, err := f.doApiRequest(addUrl, []byte(formData.Encode()), true)
	if err != nil {
		return err
	}

	logger.Debugf("Response from FreshRSS: %s", res)

	var resData FreshRSSAddFeedResponse
	err = json.Unmarshal(res, &resData)

	if err != nil {
		logger.Error("Unable to parse FreshRSS response", err)
		return err
	}

	// Add the feed to the category
	err = f.editFeed(resData.StreamId, name, category)
	if err != nil {
		logger.Error("Unable to add feed to category", err)
		return err
	}

	logger.Infof("Successfully added feed %s to FreshRSS", feed)
	return nil
}

func (f *FreshRSSFeedPublisher) editFeed(streamId, name, category string) error {
	logger := f.logger

	addUrl := fmt.Sprintf(
		"%s/api/greader.php/reader/api/0/subscription/edit", f.baseUrl,
	)

	formData := url.Values{
		"ac": {"edit"},
		"s":  {streamId},
		"t":  {name},
		"a":  {fmt.Sprintf("user/%s/label/%s", f.user, category)},
	}

	logger.Debugf("Adding feed to FreshRSS %s: %s", addUrl, formData.Encode())

	_, err := f.doApiRequest(addUrl, []byte(formData.Encode()), true)
	if err != nil {
		return err
	}

	return nil
}

func (f *FreshRSSFeedPublisher) doApiRequest(
	url string, payload []byte, authHeader bool) ([]byte, error) {
	logger := f.logger

	// Set headers
	headers := map[string]string{
		"Content-type": "application/x-www-form-urlencoded",
	}
	if authHeader {
		headers["Authorization"] = fmt.Sprintf("GoogleLogin auth=%s", f.authToken)
	}

	// Create request
	reader := bytes.NewReader(payload)
	req, err := http.NewRequest("POST", url, reader)
	if err != nil {
		logger.Error("Unable to create request to FreshRSS", err)
		return nil, err
	}

	// Process headers
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	// Make request
	res, err := f.client.Do(req)
	if err != nil {
		logger.Error("Unable to make request to FreshRSS", err)
		return nil, err
	}
	defer res.Body.Close()
	data, err := io.ReadAll(res.Body)

	if err != nil {
		logger.Error("Unable to get response data from FreshRSS", err)
		return nil, err
	}

	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusCreated {
		logger.Errorf("Response from FreshRSS %s", data)
		return nil, fmt.Errorf("freshrss returned an http error code %d", res.StatusCode)
	}

	return data, nil
}

func (f *FreshRSSFeedPublisher) parsePlainTextAuthResponse(respData []byte) error {

	logger := f.logger

	lines := strings.Split(string(respData), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "Auth=") {
			f.authToken = strings.TrimPrefix(line, "Auth=")
		}
	}

	if f.authToken == "" {
		logger.Error("Unable to parse FreshRSS auth response")
		return fmt.Errorf("unable to parse FreshRSS auth response")
	}

	return nil
}
