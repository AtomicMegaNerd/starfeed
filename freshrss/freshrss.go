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

type FreshRSSFeedManager struct {
	baseUrl   string
	user      string
	apiToken  string // WARNING: Do not log this value as it is a secret
	authToken string // WARNING: Do not log this value as it is a secret
	client    *http.Client
}

func NewFreshRSSSubManager(
	baseUrl string,
	user string,
	apiToken string,
	client *http.Client,
) *FreshRSSFeedManager {
	return &FreshRSSFeedManager{
		baseUrl:  baseUrl,
		user:     user,
		apiToken: apiToken,
		client:   client,
	}
}

func (f *FreshRSSFeedManager) Authenticate() error {

	reqUrl := fmt.Sprintf("%s/api/greader.php/accounts/ClientLogin", f.baseUrl)
	log.Debugf("Authenticating with FreshRSS at %s", reqUrl)

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

	log.Info("Authenticated with FreshRSS")
	return nil
}

func (f *FreshRSSFeedManager) AddFeed(sub, name, category string) error {

	addUrl := fmt.Sprintf(
		"%s/api/greader.php/reader/api/0/subscription/quickadd", f.baseUrl,
	)
	formData := url.Values{
		"quickadd": {sub},
	}

	log.Debugf("Adding feed to FreshRSS %s: %s", addUrl, formData.Encode())
	res, err := f.doApiRequest(addUrl, []byte(formData.Encode()), true)
	if err != nil {
		return err
	}

	// Parse the response
	var resData FreshRSSAddFeedResponse
	err = json.Unmarshal(res, &resData)
	if err != nil {
		log.Error("Unable to parse FreshRSS response", err)
		return err
	}

	// Add the sub to the category
	err = f.addFeedToCategory(resData.StreamId, name, category)
	if err != nil {
		log.Error("Unable to add feed to category", err)
		return err
	}

	log.Infof("Successfully added feed %s to FreshRSS", sub)
	return nil
}

func (f *FreshRSSFeedManager) addFeedToCategory(streamId, name, category string) error {

	addUrl := fmt.Sprintf(
		"%s/api/greader.php/reader/api/0/subscription/edit", f.baseUrl,
	)

	formData := url.Values{
		"ac": {"edit"},
		"s":  {streamId},
		"t":  {name},
		"a":  {fmt.Sprintf("user/%s/label/%s", f.user, category)},
	}

	_, err := f.doApiRequest(addUrl, []byte(formData.Encode()), true)
	if err != nil {
		return err
	}

	return nil
}

func (f *FreshRSSFeedManager) GetExistingFeeds() (map[string]struct{}, error) {

	getUrl := fmt.Sprintf(
		"%s/api/greader.php/reader/api/0/subscription/list?output=json", f.baseUrl,
	)
	log.Debug("Querying feeds in FreshRSS %s", getUrl)

	// Perform the request
	res, err := f.doApiRequest(getUrl, nil, true)
	if err != nil {
		return nil, err
	}

	// Parse the response
	var feeds RSSFeedList
	err = json.Unmarshal(res, &feeds)
	if err != nil {
		return nil, err
	}

	var feedMap map[string]struct{} = make(map[string]struct{})
	for _, feed := range feeds.Feeds {
		feedMap[feed.Url] = struct{}{}
	}
	return feedMap, nil
}

func (f *FreshRSSFeedManager) doApiRequest(
	url string, payload []byte, authHeader bool) ([]byte, error) {

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
		log.Error("Unable to create request to FreshRSS", err)
		return nil, err
	}

	// Process headers
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	// Make request
	res, err := f.client.Do(req)
	if err != nil {
		log.Error("Unable to make request to FreshRSS", err)
		return nil, err
	}
	defer res.Body.Close()
	data, err := io.ReadAll(res.Body)
	if err != nil {
		log.Error("Unable to get response data from FreshRSS", err)
		return nil, err
	}

	log.Debugf("FreshRSS response %s", data)
	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("FreshRSS returned an http error code %d", res.StatusCode)
	}

	return data, nil
}

func (f *FreshRSSFeedManager) parsePlainTextAuthResponse(respData []byte) error {

	lines := strings.Split(string(respData), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "Auth=") {
			f.authToken = strings.TrimPrefix(line, "Auth=")
		}
	}

	if f.authToken == "" {
		log.Error("Unable to parse FreshRSS auth response")
		return fmt.Errorf("unable to parse FreshRSS auth response")
	}

	return nil
}
