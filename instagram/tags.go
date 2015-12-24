// Copyright 2013 The go-instagram AUTHORS. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package instagram

import (
	// "errors"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
)

// TagsService handles communication with the tag related
// methods of the Instagram API.
//
// Instagram API docs: http://instagram.com/developer/endpoints/tags/
type TagsService struct {
	client *Client
}

// Tag represents information about a tag object.
type Tag struct {
	MediaCount float64 `json:"media_count,omitempty"`
	Name       string  `json:"name,omitempty"`
}

// Get information aout a tag object.
//
// Instagram API docs: http://instagram.com/developer/endpoints/tags/#get_tags
func (s *TagsService) Get(tagName string) (*Tag, error) {
	u := fmt.Sprintf("tags/%v", tagName)
	req, err := s.client.NewRequest("GET", u, "")
	if err != nil {
		return nil, err
	}

	tag := new(Tag)
	_, err = s.client.Do(req, tag)
	return tag, err
}

// RecentMedia Get a list of recently tagged media.
//
// Instagram API docs: http://instagram.com/developer/endpoints/tags/#get_tags_media_recent
func (s *TagsService) RecentMedia(tagName string, opt *Parameters) ([]Media, *ResponsePagination, error) {
	valid, err := validTagName(tagName)
	if err != nil {
		return nil, nil, err
	}

	if !valid {
		//For now, I don't think this is an error but instead an early catch of an input that won't give a result
		//but it's not clearly defined (as far as I can tell) in the Instagram spec that it *couldn't* give a result
		//In future, this might change to give an error, though
		//return nil, nil, errors.New(`go-instagram Tag.RecentMedia error: Tag names must contain only alphabetical and numerical characters.`)
		return []Media{}, &ResponsePagination{}, nil
	}

	u := fmt.Sprintf("tags/%v/media/recent", tagName)
	if opt != nil {
		params := url.Values{}
		if opt.Count != 0 {
			params.Add("count", strconv.FormatUint(opt.Count, 10))
		}
		if opt.MinID != "" {
			params.Add("min_tag_id", opt.MinID)
		}
		if opt.MaxID != "" {
			params.Add("max_tag_id", opt.MaxID)
		}
		u += "?" + params.Encode()
	}
	req, err := s.client.NewRequest("GET", u, "")
	if err != nil {
		return nil, nil, err
	}

	media := new([]Media)

	_, err = s.client.Do(req, media)
	if err != nil {
		return nil, nil, err
	}
	/*
		if err != nil {
			if req != nil && req.URL != nil {
				return nil, nil, errors.New(fmt.Sprintf("go-instagram Tag.RecentMedia error:%s on URL %s", err.Error(), req.URL.String()))
			} else {
				return nil, nil, errors.New(fmt.Sprintf("go-instagram Tag.RecentMedia error:%s on nil URL", err.Error()))
			}
		}
	*/

	page := new(ResponsePagination)
	if s.client.Response.Pagination != nil {
		page = s.client.Response.Pagination
	}

	return *media, page, err
}

// Search for tags by name.
//
// Instagram API docs: http://instagram.com/developer/endpoints/tags/#get_tags_search
func (s *TagsService) Search(q string) ([]Tag, *ResponsePagination, error) {
	u := "tags/search?q=" + q
	req, err := s.client.NewRequest("GET", u, "")
	if err != nil {
		return nil, nil, err
	}

	tags := new([]Tag)

	_, err = s.client.Do(req, tags)
	if err != nil {
		return nil, nil, err
	}

	page := new(ResponsePagination)
	if s.client.Response.Pagination != nil {
		page = s.client.Response.Pagination
	}

	return *tags, page, err
}

// Strip out things we know Instagram won't accept. For example, hyphens.
func validTagName(tagName string) (bool, error) {
	//\W matches any non-word character
	reg, err := regexp.Compile(`\W`)
	if err != nil {
		return false, err
	}

	if reg.MatchString(tagName) {
		return false, nil
	}

	return true, nil
}
