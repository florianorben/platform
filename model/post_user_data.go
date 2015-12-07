// Copyright (c) 2015 Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package model

import (
	"encoding/json"
	"io"
	"unicode/utf8"
)

const (
	POST_USER_DATA_STARRED_CHANNEL = "c"
	POST_USER_DATA_STARRED_POST    = "s"
	POST_USER_DATA_REACTION        = "r"
	POST_USER_DATA_PINNED_POST     = "p"
)

const (
	POST_USER_DATA_TYPE_DATA_GLOBAL = "global"
	POST_USER_DATA_TYPE_DATA_USER   = "user"
)

type Reactions []PostUserDataReactions
type PostUserDataList []PostUserData

type PostUserDataReactions struct {
	PostUserData
	Count    int64  `json:"count"`
	DataType string `json:"data_type"`
}

type PostUserData struct {
	Id        string `json:"id"`
	CreateAt  int64  `json:"create_at"`
	PostId    string `json:"post_id"`
	UserId    string `json:"user_id"`
	ChannelId string `json:"channel_id"`
	Type      string `json:"type"`
	Content   string `json:"content"`
}

func (o *Reactions) ToJson() string {
	b, err := json.Marshal(o)
	if err != nil {
		return ""
	} else {
		return string(b)
	}
}

func (o *PostUserDataList) ToJson() string {
	b, err := json.Marshal(o)
	if err != nil {
		return ""
	} else {
		return string(b)
	}
}

func (o *PostUserData) ToJson() string {
	b, err := json.Marshal(o)
	if err != nil {
		return ""
	} else {
		return string(b)
	}
}

func PostUserDataListFromJson(data io.Reader) *PostUserDataList {
	decoder := json.NewDecoder(data)
	var o PostUserDataList
	err := decoder.Decode(&o)
	if err == nil {
		return &o
	} else {
		return nil
	}
}

func ReactionsFromJson(data io.Reader) *Reactions {
	decoder := json.NewDecoder(data)
	var o Reactions
	err := decoder.Decode(&o)
	if err == nil {
		return &o
	} else {
		return nil
	}
}

func PostUserDataFromJson(data io.Reader) *PostUserData {
	decoder := json.NewDecoder(data)
	var o PostUserData
	err := decoder.Decode(&o)
	if err == nil {
		return &o
	} else {
		return nil
	}
}

func (o *PostUserData) IsValid() *AppError {
	if len(o.Id) != 26 {
		return NewAppError("PostUserData.IsValid", "Invalid Id", "")
	}

	if o.CreateAt == 0 {
		return NewAppError("PostUserData.IsValid", "Create at must be a valid time", "id="+o.Id)
	}

	if !(o.Type == POST_USER_DATA_STARRED_CHANNEL || o.Type == POST_USER_DATA_STARRED_POST ||
		o.Type == POST_USER_DATA_REACTION || o.Type == POST_USER_DATA_PINNED_POST) {
		return NewAppError("PostUserData.IsValid", "Invalid type", "id="+o.Id)
	}

	if (o.Type == POST_USER_DATA_REACTION || o.Type == POST_USER_DATA_PINNED_POST || o.Type == POST_USER_DATA_STARRED_POST) && len(o.PostId) != 26 {
		return NewAppError("PostUserData.IsValid", "Invalid post id", "id="+o.Id)
	} else if o.Type == POST_USER_DATA_STARRED_CHANNEL && len(o.ChannelId) != 26 {
		return NewAppError("PostUserData.IsValid", "Invalid channel id", "id="+o.Id)
	}

	if len(o.UserId) != 26 {
		return NewAppError("PostUserData.IsValid", "Invalid user id", "")
	}

	if utf8.RuneCountInString(o.Content) > 32 {
		return NewAppError("PostUserData.IsValid", "Invalid content", "id="+o.Id)
	}

	return nil
}

func (o *PostUserData) PreSave() {
	if o.Id == "" {
		o.Id = NewId()
	}

	if o.CreateAt == 0 {
		o.CreateAt = GetMillis()
	}
}
