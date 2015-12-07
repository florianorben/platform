// Copyright (c) 2015 Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package api

import (
	"fmt"
	"net/http"
	"strings"

	l4g "code.google.com/p/log4go"
	"github.com/gorilla/mux"
	"github.com/mattermost/platform/model"
)

func InitPostUserData(r *mux.Router) {
	l4g.Debug("Initializing post_user_data api routes")

	starredRoute := r.PathPrefix("/starred").Subrouter()
	starredRoute.Handle("/{id:[A-Za-z0-9]+}", ApiAppHandler(getStarredById)).Methods("GET")
	starredRoute.Handle("/{id:[A-Za-z0-9]+}/delete", ApiAppHandler(deletePostUserData)).Methods("GET")
	starredRoute.Handle("/posts", ApiAppHandler(getStarredPosts)).Methods("GET")
	starredRoute.Handle("/posts/create", ApiAppHandler(createStarredPost)).Methods("POST")
	starredRoute.Handle("/channels", ApiAppHandler(getStarredChannels)).Methods("GET")
	starredRoute.Handle("/channels/create", ApiAppHandler(createStarredChannel)).Methods("POST")
}

func createPostUserData(c *Context, w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	data := model.PostUserDataFromJson(r.Body)

	data.Id = model.NewId()
	data.UserId = c.Session.UserId
	data.PostId = params["postId"]
	data.PreSave()
	fmt.Println(data)
}

func searchPostUserData(c *Context, w http.ResponseWriter, r *http.Request) {
	schan := Srv.Store.PostUserData().GetByPostIds([]string{"gqx7ge1agpre5d7f4rq4yym4po", "skjdfsfdnasdasd"})

	if result := <-schan; result.Err != nil {
		c.Err = result.Err
		return
	} else {
		response := result.Data.(model.Reactions)
		w.Write([]byte(response.ToJson()))
	}
}

func getStarredById(c *Context, w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	schan := Srv.Store.PostUserData().Get(params["id"])

	if result := <-schan; result.Err != nil {
		c.Err = result.Err
		return
	} else {
		response := result.Data.(model.PostUserData)
		w.Write([]byte(response.ToJson()))
	}
}

func getStarredPosts(c *Context, w http.ResponseWriter, r *http.Request) {

	fmt.Println(c.Session.UserId)
	schan := Srv.Store.PostUserData().GetStarredPosts(c.Session.UserId)

	if result := <-schan; result.Err != nil {
		c.Err = result.Err
		return
	} else {
		response := result.Data.(model.PostUserDataList)
		w.Write([]byte(response.ToJson()))
	}
}

func getStarredChannels(c *Context, w http.ResponseWriter, r *http.Request) {
	schan := Srv.Store.PostUserData().GetStarredChannels(c.Session.UserId)

	if result := <-schan; result.Err != nil {
		c.Err = result.Err
		return
	} else {
		response := result.Data.(model.PostUserDataList)
		w.Write([]byte(response.ToJson()))
	}
}

func createStarredPost(c *Context, w http.ResponseWriter, r *http.Request) {
	data := model.PostUserDataFromJson(r.Body)

	if data == nil {
		c.SetInvalidParam("createStarredPost", "post_user_data")
		return
	} else {
		data.Type = model.POST_USER_DATA_STARRED_POST
		data.UserId = c.Session.UserId
	}

	pchan := Srv.Store.Post().Get(data.PostId)
	if postResult := <-pchan; postResult.Err != nil {
		c.Err = postResult.Err
		return
	} else {
		postResult := postResult.Data.(*model.PostList)
		if len(postResult.Posts) != 1 {
			c.SetInvalidParam("createStarredPost", "post_id")
			return
		} else if data.ChannelId == "" {
			data.ChannelId = postResult.Posts[data.PostId].ChannelId
		}

		cchan := Srv.Store.Channel().CheckPermissionsTo(c.Session.TeamId, data.ChannelId, c.Session.UserId)
		if !c.HasPermissionsToChannel(cchan, "starPost") {
			return
		}

		if result := <-Srv.Store.PostUserData().Save(data); result.Err != nil {
			c.Err = result.Err
			return
		} else {
			postUserData := result.Data.(*model.PostUserData)

			message := model.NewMessage(c.Session.TeamId, postUserData.ChannelId, c.Session.UserId, model.ACTION_POST_STARRED)
			message.Add("starred_post", postUserData.ToJson())

			PublishAndForget(message)
			w.Write([]byte(postUserData.ToJson()))
		}
	}
}

func createStarredChannel(c *Context, w http.ResponseWriter, r *http.Request) {
	data := model.PostUserDataFromJson(r.Body)

	if data == nil {
		c.SetInvalidParam("createStarredChannel", "post_user_data")
		return
	} else {
		data.Type = model.POST_USER_DATA_STARRED_CHANNEL
		data.UserId = c.Session.UserId
	}

	chchan := Srv.Store.Channel().Get(data.ChannelId)
	if chanResult := <-chchan; chanResult.Err != nil {
		c.Err = chanResult.Err
		return
	} else {
		cchan := Srv.Store.Channel().CheckPermissionsTo(c.Session.TeamId, data.ChannelId, c.Session.UserId)
		if !c.HasPermissionsToChannel(cchan, "starChannel") {
			return
		}

		if result := <-Srv.Store.PostUserData().Save(data); result.Err != nil {
			c.Err = result.Err
			return
		} else {
			postUserData := result.Data.(*model.PostUserData)

			//double check, don't want to broadcast stars (and future tags?) to everyone
			message := model.NewMessage(c.Session.TeamId, postUserData.ChannelId, c.Session.UserId, model.ACTION_CHANNEL_STARRED)
			message.Add("starred_channel", postUserData.ToJson())

			PublishAndForget(message)
			w.Write([]byte(postUserData.ToJson()))
		}
	}
}

func deletePostUserData(c *Context, w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]

	pc := Srv.Store.PostUserData().Get(id)
	scm := Srv.Store.Channel().GetMember(id, c.Session.UserId)

	if presult := <-pc; presult.Err != nil {
		c.Err = presult.Err
		return
	} else if scmresult := <-scm; scmresult.Err != nil {
		c.Err = scmresult.Err
		return
	} else {
		data := presult.Data.(*model.PostUserData)
		channelMember := scmresult.Data.(model.ChannelMember)

		if (!strings.Contains(channelMember.Roles, model.CHANNEL_ROLE_ADMIN) && !strings.Contains(c.Session.Roles, model.ROLE_TEAM_ADMIN)) ||
			c.Session.UserId != data.UserId {
			c.Err = model.NewAppError("deleteChannel", "You do not have the appropriate permissions", "")
			c.Err.StatusCode = http.StatusForbidden
			return
		}

		if dresult := <-Srv.Store.PostUserData().Delete(id); dresult.Err != nil {
			c.Err = dresult.Err
			return
		}

		result := make(map[string]string)
		result["id"] = id
		w.Write([]byte(model.MapToJson(result)))
	}
}
