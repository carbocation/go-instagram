package instagram

import (
	"fmt"
	"net/http"
	"net/url"
)

type RealtimeService struct {
	client *Client
}

// Realtime represents a realtime subscription on Instagram's service.
type Realtime struct {
	ID          string `json:"id,omitempty"`
	Type        string `json:"type,omitempty"`
	Object      string `json:"object,omitempty"`
	ObjectID    string `json:"object_id,omitempty"`
	Aspect      string `json:"aspect,omitempty"`
	CallbackURL string `json:"callback_url,omitempty"`
}

type RealtimeResponse struct {
	SubscriptionID int64  `json:"subscription_id,omitempty"`
	Object         string `json:"object,omitempty"`
	ObjectID       string `json:"object_id,omitempty"`
	ChangedAspect  string `json:"changed_aspect,omitempty"`
	Time           int64  `json:"time,omitempty"`
}

//ListSubscriptions ists the realtime subscriptions that are already active for your account
func (s *RealtimeService) ListSubscriptions() ([]Realtime, error) {
	u := "subscriptions/"

	req, err := s.client.NewRequest("GET", u, "")
	if err != nil {
		return nil, err
	}

	realtime := new([]Realtime)

	_, err = s.client.Do(req, realtime)
	if err != nil {
		return nil, err
	}

	return *realtime, err
}

// SubscribeToTag initiates the subscription to realtime updates about tag `tag`
//
// Instagram API docs: http://instagram.com/developer/realtime/
func (s *RealtimeService) SubscribeToTag(tag, callbackURL, verifyToken string) (*Realtime, error) {
	u := "subscriptions/"

	params := url.Values{
		"aspect":        {"media"},
		"object":        {"tag"},
		"object_id":     {tag},
		"callback_url":  {callbackURL},
		"client_id":     {s.client.ClientID},
		"client_secret": {s.client.ClientSecret},
		"verify_token":  {verifyToken},
	}

	req, err := s.client.NewRequest("POST", u, params.Encode())
	if err != nil {
		return nil, err
	}

	realtime := new(Realtime)

	_, err = s.client.Do(req, realtime)
	if err != nil {
		return nil, err
	}

	return realtime, err
}

// DeleteAllSubscriptions deletes all active subscriptions for an account.
//
// Instagram API docs: http://instagram.com/developer/realtime/
func (s *RealtimeService) DeleteAllSubscriptions() (*Realtime, error) {
	u := "subscriptions/"

	params := url.Values{
		"object":        {"all"},
		"client_id":     {s.client.ClientID},
		"client_secret": {s.client.ClientSecret},
	}

	u += "?" + params.Encode()

	req, err := s.client.NewRequest("DELETE", u, "")
	if err != nil {
		return nil, err
	}

	realtime := new(Realtime)

	_, err = s.client.Do(req, realtime)
	if err != nil {
		return nil, err
	}

	return realtime, err
}

// UnsubscribeFrom unsubscribes you from a specific subscription.
//
// Instagram API docs: http://instagram.com/developer/realtime/
func (s *RealtimeService) UnsubscribeFrom(sid string) (*Realtime, error) {
	u := "subscriptions/"

	params := url.Values{
		"id":            {sid},
		"client_id":     {s.client.ClientID},
		"client_secret": {s.client.ClientSecret},
	}

	u += "?" + params.Encode()

	req, err := s.client.NewRequest("DELETE", u, "")
	if err != nil {
		return nil, err
	}

	realtime := new(Realtime)

	_, err = s.client.Do(req, realtime)
	if err != nil {
		return nil, err
	}

	return realtime, err
}

//An example RealTimeSubscribe ResponseWriter. This can be plugged directly into
// any standard http server. Note, however, that this particular implementation does
// no checking that the verifyToken is correct.
func ServeInstagramRealtimeSubscribe(w http.ResponseWriter, r *http.Request) {
	verify := r.FormValue("hub.challenge")

	fmt.Fprintf(w, verify)
}
