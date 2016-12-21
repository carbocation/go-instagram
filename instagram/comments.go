// Copyright 2013 The go-instagram AUTHORS. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package instagram

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
)

// CommentsService handles communication with the comments related
// methods of the Instagram API.
//
// Instagram API docs: http://instagram.com/developer/endpoints/comments/
type CommentsService struct {
	client *Client
}

// Comment represents a comment on Instagram's media.
type Comment struct {
	CreatedTime int64  `json:"created_time,string,omitempty"`
	Text        string `json:"text,omitempty"`
	From        *User  `json:"from,omitempty"`
	ID          string `json:"id,omitempty"`
}

// MediaComments gets a full list of comments on a media.
//
// Instagram API docs: http://instagram.com/developer/endpoints/comments/#get_media_comments
func (s *CommentsService) MediaComments(mediaID string) ([]Comment, error) {
	u := fmt.Sprintf("media/%v/comments", mediaID)
	req, err := s.client.NewRequest("GET", u, "")
	if err != nil {
		return nil, err
	}

	comments := new([]Comment)
	_, err = s.client.Do(req, comments)
	return *comments, err
}

// Add a comment on a media.
//
// Instagram API docs: http://instagram.com/developer/endpoints/comments/#post_media_comments
func (s *CommentsService) Add(mediaID string, text []string) error {
	u := fmt.Sprintf("media/%v/comments", mediaID)
	params := url.Values{
		"text": text,
	}

	req, err := s.client.NewRequest("POST", u, params.Encode())
	if err != nil {
		return err
	}
	log.Print("")
	_, err = s.client.Do(req, nil)
	return err
}

//SendSignedRequest comment on a media
//
//https://www.instagram.com/developer/endpoints/comments/
func (s *CommentsService) SendSignedRequest(mediaID string, text string) (*http.Response, error) {
	urlStr := fmt.Sprintf("media/%v/comments", mediaID)
	params, err := s.client.setParams(urlStr, text)
	if err != nil {
		return nil, err
	}
	resp, err := http.PostForm(BaseURL+urlStr, params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	err = CheckResponse(resp)
	if err != nil {
		return resp, err
	}
	var v interface{}
	r := &Response{Response: resp}
	if v != nil {
		r.Data = v
		err = json.NewDecoder(resp.Body).Decode(r)
		s.client.Response = r
	}
	return resp, err
}

// Delete a comment either on the authenticated user's media or authored by
// the authenticated user.
//
// Instagram API docs: http://instagram.com/developer/endpoints/comments/#delete_media_comments
func (s *CommentsService) Delete(mediaID, commentID string) error {
	u := fmt.Sprintf("media/%v/comments/%v", mediaID, commentID)
	req, err := s.client.NewRequest("DELETE", u, "")
	if err != nil {
		return err
	}

	_, err = s.client.Do(req, nil)
	return err
}
