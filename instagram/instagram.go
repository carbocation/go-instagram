// Copyright 2013 The go-instagram AUTHORS. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package instagram provides a client for using the Instagram API.

Access different parts of the Instagram API using the various services on a Instagram
Client (second parameter is access token that likely you'll need to access most of
Instagram endpoints):

	client := instagram.NewClient(nil)

You can then optionally set ClientID, ClientSecret and AccessToken:

	client.ClientID = "8f2c0ad697ea4094beb2b1753b7cde9c"

With client object set, you can call Instagram endpoints:

	// Gets the most recent media published by a user with id "3"
	media, next, err := client.Users.RecentMedia("3", nil)

Set optional parameters for an API method by passing an Parameters object.

	// Gets user's feed.
	opt := &instagram.Parameters{Count: 3}
	media, next, err := client.Users.RecentMedia("3", opt)

The full Instagram API is documented at http://instagram.com/developer/endpoints/.
*/
package instagram

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
)

const (
	// LibraryVersion represents this library version
	LibraryVersion = "0.5"

	// BaseURL represents Instagram API base URL
	BaseURL = "https://api.instagram.com/v1/"

	// UserAgent represents this client User-Agent
	UserAgent = "github.com/carbocation/go-instagram v" + LibraryVersion
)

// A Client manages communication with the Instagram API.
type Client struct {
	// HTTP client used to communicate with the API.
	client *http.Client

	// Base URL for API requests.
	BaseURL *url.URL

	// UserAgent agent used when communicating with Instagram API.
	UserAgent string

	// Application client_id
	ClientID string

	// Application client_secret
	ClientSecret string

	// Authenticated user's access_token
	AccessToken string

	// For Authenticated endpoints, using X-Forwarded-For
	// increases events per hour permitted by Instagram
	// This value should, if not nil, be the value of
	// a user's IP address. See
	// http://instagram.com/developer/restrict-api-requests/
	// for additional detail
	XInstaForwardedFor string

	// Services used for talking to different parts of the API.
	Users         *UsersService
	Relationships *RelationshipsService
	Media         *MediaService
	Comments      *CommentsService
	Likes         *LikesService
	Tags          *TagsService
	Locations     *LocationsService
	Geographies   *GeographiesService
	Realtime      *RealtimeService

	// Temporary Response
	Response *Response
}

// Parameters specifies the optional parameters to various service's methods.
type Parameters struct {
	Count        uint64
	MinID        string
	MaxID        string
	MinTimestamp int64
	MaxTimestamp int64
	Lat          float64
	Lng          float64
	Distance     float64
}

// Ratelimit specifies API calls limit found in HTTP headers.
type Ratelimit struct {
	// Total number of possible calls per hour
	Limit int

	// How many calls are left for this particular token or client ID
	Remaining int
}

// Response specifies Instagram's response structure.
//
// Instagram's envelope structure spec: http://instagram.com/developer/endpoints/#structure
type Response struct {
	Response   *http.Response      // HTTP response
	Meta       *ResponseMeta       `json:"meta,omitempty"`
	Data       interface{}         `json:"data,omitempty"`
	Pagination *ResponsePagination `json:"pagination,omitempty"`
}

// GetMeta gets extra information about the response. If all goes well,
// only Code key with value 200 is returned. If something goes wrong,
// ErrorType and ErrorMessage keys are present.
func (r *Response) GetMeta() *ResponseMeta {
	return r.Meta
}

// GetData gets the meat of the response.
func (r *Response) GetData() interface{} {
	return &r.Data
}

// GetError gets error from meta's response.
func (r *Response) GetError() error {
	if r.Meta.ErrorType != "" || r.Meta.ErrorMessage != "" {
		return fmt.Errorf(fmt.Sprintf("%s: %s", r.Meta.ErrorType, r.Meta.ErrorMessage))
	}
	return nil
}

// GetPagination gets pagination information.
func (r *Response) GetPagination() *ResponsePagination {
	return r.Pagination
}

// Parsed rate limit information from response headers.
func (r *Response) GetRatelimit() (Ratelimit, error) {
	var rl Ratelimit
	var err error
	const (
		Limit     = `X-Ratelimit-Limit`
		Remaining = `X-Ratelimit-Remaining`
	)

	rl.Limit, err = strconv.Atoi(r.Response.Header.Get(Limit))
	if err != nil {
		return rl, err
	}

	rl.Remaining, err = strconv.Atoi(r.Response.Header.Get(Remaining))
	return rl, err
}

// NextURL gets next url which represents URL for next set of data.
func (r *Response) NextURL() string {
	p := r.GetPagination()
	return p.NextURL
}

// NextMaxID gets MaxID parameter that can be passed for next request.
func (r *Response) NextMaxID() string {
	p := r.GetPagination()
	return p.NextMaxID
}

// ResponseMeta represents information about the response. If all goes well,
// only a Code key with value 200 will present. However, sometimes things
// go wrong, and in that case ErrorType and ErrorMessage are present.
type ResponseMeta struct {
	ErrorType    string `json:"error_type,omitempty"`
	Code         int    `json:"code,omitempty"`
	ErrorMessage string `json:"error_message,omitempty"`
}

// ResponsePagination represents information to get access to more data in
// any request for sequential data.
type ResponsePagination struct {
	NextURL   string `json:"next_url,omitempty"`
	NextMaxID string `json:"next_max_id,omitempty"`
}

// NewClient returns a new Instagram API client. if a nil httpClient is
// provided, http.DefaultClient will be used.
func NewClient(httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	baseURL, _ := url.Parse(BaseURL)

	c := &Client{
		client:    httpClient,
		BaseURL:   baseURL,
		UserAgent: UserAgent,
	}
	c.Users = &UsersService{client: c}
	c.Relationships = &RelationshipsService{client: c}
	c.Media = &MediaService{client: c}
	c.Comments = &CommentsService{client: c}
	c.Likes = &LikesService{client: c}
	c.Tags = &TagsService{client: c}
	c.Locations = &LocationsService{client: c}
	c.Geographies = &GeographiesService{client: c}
	c.Realtime = &RealtimeService{client: c}

	return c
}

func (c *Client) ComputeXInstaForwardedFor() string {
	if c.XInstaForwardedFor == "" {
		return ""
	}

	mac := hmac.New(sha256.New, []byte(c.ClientSecret))
	mac.Write([]byte(c.XInstaForwardedFor))

	return fmt.Sprintf("%s|%s", c.XInstaForwardedFor, hex.EncodeToString(mac.Sum(nil)))
}

// NewRequest creates an API request. A relative URL can be provided in urlStr,
// in which case it is resolved relative to the BaseURL of the Client.
// Relative URLs should always be specified without a preceding slash. If
// specified
func (c *Client) NewRequest(method, urlStr string, body string) (*http.Request, error) {
	rel, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	u := c.BaseURL.ResolveReference(rel)
	q := u.Query()
	if c.AccessToken != "" && q.Get("access_token") == "" {
		q.Set("access_token", c.AccessToken)
	}
	if c.ClientID != "" && q.Get("client_id") == "" {
		q.Set("client_id", c.ClientID)
	}
	if c.ClientSecret != "" && q.Get("client_secret") == "" {
		q.Set("client_secret", c.ClientSecret)
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequest(method, u.String(), bytes.NewBufferString(body))
	if err != nil {
		return nil, err
	}

	if method == "POST" {
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	}

	if c.XInstaForwardedFor != "" {
		req.Header.Add("X-Insta-Forwarded-For", c.ComputeXInstaForwardedFor())
	}

	req.Header.Add("User-Agent", c.UserAgent)
	return req, nil
}

// Do sends an API request and returns the API response. The API response is
// decoded and stored in the value pointed to by v, or returned as an error if
// an API error has occurred.
func (c *Client) Do(req *http.Request, v interface{}) (*http.Response, error) {
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	err = CheckResponse(resp)
	if err != nil {
		return resp, err
	}

	r := &Response{Response: resp}
	if v != nil {
		r.Data = v
		err = json.NewDecoder(resp.Body).Decode(r)
		c.Response = r
	}
	return resp, err
}

// InstagramError represents an error recieved from instagram
type InstagramError ResponseMeta

// Error makes the InstagramError suitable for the error interface
func (err *InstagramError) Error() string {
	return fmt.Sprintf("%s (%d): %s", err.ErrorType, err.Code, err.ErrorMessage)
}

// ErrorResponse represents a Response which contains an error
type ErrorResponse Response

func (r *ErrorResponse) Error() string {
	if r == nil {
		return fmt.Sprintf("A nil error response was returned")
	}

	if r.Response == nil || r.Response.Request == nil {
		return fmt.Sprintf("A nil error response was returned on %v", r)
	}

	if r.Response.Request.URL == nil {
		return fmt.Sprintf("A nil error response was returned on %v", r.Response.Request)
	}

	if r.Meta == nil {
		return fmt.Sprintf("%v %v: %d (no metadata)", r.Response.Request.Method, r.Response.Request.URL,
			r.Response.StatusCode)
	}

	return fmt.Sprintf("%v %v: %d %v %v",
		r.Response.Request.Method, r.Response.Request.URL,
		r.Response.StatusCode, r.Meta.ErrorType, r.Meta.ErrorMessage)
}

// CheckResponse checks the API response for error, and returns it
// if present. A response is considered an error if it has non StatusOK
// code.
func CheckResponse(r *http.Response) error {
	if r.StatusCode == http.StatusOK {
		return nil
	}

	data, readErr := ioutil.ReadAll(r.Body)
	if readErr != nil {
		return readErr
	}

	// Forbidden: see http://instagram.com/developer/restrict-api-requests/
	if r.StatusCode == http.StatusForbidden {
		err := &InstagramError{}
		json.Unmarshal(data, &err)
		return err
	}

	// RateLimit: see http://instagram.com/developer/limits/
	if r.StatusCode == 429 {
		err := &InstagramError{}
		json.Unmarshal(data, &err)
		return err
	}

	// Sometimes Instagram returns 500 with plain message
	// "Oops, an error occurred.".
	if r.StatusCode == http.StatusInternalServerError {
		err := &InstagramError{
			ErrorType:    "Internal Server Error",
			Code:         http.StatusInternalServerError,
			ErrorMessage: "Oops, an error occurred.",
		}
		return err
	}

	if data != nil {
		// Unlike for successful (2XX) requests, unsuccessful
		// requests SOMETIMES have the {Meta: Error{}} format but
		// SOMETIMES they are just Error{}. From what I can tell, there is not
		// an obvious rationale behind what gets constructed in which way, so
		// we need to try both:
		err := &InstagramError{}
		json.Unmarshal(data, err)
		if *err != *new(InstagramError) {
			// Unmarshaling did something
			return err
		} else {
			// Unmarshaling did nothing for us, so the format was not Error{}.
			// We will assume the format was {Meta: Error{}}:
			temp := make(map[string]InstagramError)
			json.Unmarshal(data, &temp)

			meta := temp["meta"]

			delete(temp, "meta") // Probably uselesss
			return &meta
		}
	}

	return nil
}
