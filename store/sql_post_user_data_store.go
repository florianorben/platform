// Copyright (c) 2015 Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package store

import (
	"strconv"

	"github.com/mattermost/platform/model"
	//"github.com/mattermost/platform/utils"
	"strings"
)

type SqlPostUserDataStore struct {
	*SqlStore
}

func NewSqlPostUserDataStore(sqlStore *SqlStore) PostUserDataStore {
	s := &SqlPostUserDataStore{sqlStore}

	for _, db := range sqlStore.GetAllConns() {
		table := db.AddTableWithName(model.PostUserData{}, "PostUserData").SetKeys(false, "Id")
		table.ColMap("Id").SetMaxSize(26)
		table.ColMap("UserId").SetMaxSize(26)
		table.ColMap("PostId").SetMaxSize(26)
		table.ColMap("ChannelId").SetMaxSize(26)
		table.ColMap("Type").SetMaxSize(1)
		table.ColMap("Content").SetMaxSize(32)
	}

	return s
}

func (s SqlPostUserDataStore) UpgradeSchemaIfNeeded() {}

func (s SqlPostUserDataStore) CreateIndexesIfNotExists() {
	s.CreateIndexIfNotExists("idx_post_user_data_create_at", "PostUserData", "CreateAt")
	s.CreateIndexIfNotExists("idx_post_user_data_user_id", "PostUserData", "UserId")
	s.CreateIndexIfNotExists("idx_post_user_data_post_id", "PostUserData", "PostId")
	s.CreateIndexIfNotExists("idx_post_user_data_channel_id", "PostUserData", "ChannelId")
	s.CreateIndexIfNotExists("idx_post_user_data_type", "PostUserData", "Type")
}

func (s SqlPostUserDataStore) Get(id string) StoreChannel {
	storeChannel := make(StoreChannel)

	go func() {
		result := StoreResult{}

		var data model.PostUserData
		err := s.GetReplica().SelectOne(&data,
			`SELECT
				*
			FROM
				PostUserData
			WHERE
				Id = :Id`,
			map[string]interface{}{"Id": id})

		if err != nil {
			result.Err = model.NewAppError("SqlPostUserDataStore.Get", "We couldn't get the post_user_data", "Id="+id+" "+err.Error())
		} else {
			result.Data = data
		}

		storeChannel <- result
		close(storeChannel)
	}()

	return storeChannel
}

func (s SqlPostUserDataStore) GetByPostIds(postIds []string) StoreChannel {
	storeChannel := make(StoreChannel)

	go func() {
		result := StoreResult{}

		inClause := ":PostId0"
		queryParams := map[string]interface{}{"PostId0": postIds[0]}

		for i := 1; i < len(postIds); i++ {
			paramName := "PostId" + strconv.FormatInt(int64(i), 10)
			inClause += ", :" + paramName
			queryParams[paramName] = postIds[i]
		}

		var postUserData model.Reactions
		_, err := s.GetReplica().Select(&postUserData,
			`
			SELECT
				'' AS Id, 0 AS CreateAt, PostId, '' AS UserId, '' AS ChannelId,
				Type, Content, COUNT(DISTINCT(UserId)) AS Count, '`+model.POST_USER_DATA_TYPE_DATA_GLOBAL+`' AS DataType
			FROM
				PostUserData
			WHERE
				PostId IN (`+inClause+`)
				AND Type = '`+model.POST_USER_DATA_REACTION+`'
			GROUP BY
				Content
			UNION SELECT
				*, 1 AS Count, '`+model.POST_USER_DATA_TYPE_DATA_GLOBAL+`' AS DataType
			FROM
				PostUserData
			WHERE
				UserId = 'cj8ihor9tfrsjreacjf93bnrho'
				AND Type = '`+model.POST_USER_DATA_REACTION+`'
				AND PostId IN (`+inClause+`)`,
			queryParams)
		if err != nil {
			result.Err = model.NewAppError("SqlPostUserDataStore.GetByPostIds", "We couldn't get the post_user_data items", "Ids="+strings.Join(postIds, ", ")+" "+err.Error())
		} else {
			result.Data = postUserData
		}

		storeChannel <- result
		close(storeChannel)
	}()

	return storeChannel
}

func (s SqlPostUserDataStore) GetStarredChannels(id string) StoreChannel {
	return getStarredItems(s, id, model.POST_USER_DATA_STARRED_CHANNEL)
}

func (s SqlPostUserDataStore) GetStarredPosts(id string) StoreChannel {
	return getStarredItems(s, id, model.POST_USER_DATA_STARRED_POST)
}

func getStarredItems(s SqlPostUserDataStore, UserId string, t string) StoreChannel {
	storeChannel := make(StoreChannel)
	go func() {

		result := StoreResult{}
		var data model.PostUserDataList

		_, err := s.GetReplica().Select(&data,
			`
			SELECT
				*
			FROM
				PostUserData
			WHERE
				UserId = :UserId
				AND Type = :Type`,
			map[string]interface{}{"UserId": UserId, "Type": t})

		if err != nil {
			result.Err = model.NewAppError("SqlPostUserDataStore.getFavorites - "+t, "We couldn't get the favorite items for the team", "UserId="+UserId+" "+err.Error())
		} else {
			result.Data = data
		}

		storeChannel <- result
		close(storeChannel)
	}()
	return storeChannel
}

func (s SqlPostUserDataStore) GetReactionsByPostId(postId string) StoreChannel {
	storeChannel := make(StoreChannel)
	return storeChannel
}

func (s SqlPostUserDataStore) GetPinnedPostsByChannelId(channelId string) StoreChannel {
	storeChannel := make(StoreChannel)
	return storeChannel
}

func (s SqlPostUserDataStore) Save(postUserData *model.PostUserData) StoreChannel {
	storeChannel := make(StoreChannel)

	go func() {
		result := StoreResult{}

		postUserData.PreSave()
		if result.Err = postUserData.IsValid(); result.Err != nil {
			storeChannel <- result
			close(storeChannel)
			return
		}

		if err := s.GetMaster().Insert(postUserData); err != nil {
			result.Err = model.NewAppError("SqlPostUserDataStore.Save", "We couldn't save the UserData", "id="+postUserData.Id+", "+err.Error())
		} else {
			result.Data = postUserData
		}

		storeChannel <- result
		close(storeChannel)
	}()

	return storeChannel
}

func (s SqlPostUserDataStore) Delete(id string) StoreChannel {
	storeChannel := make(StoreChannel)

	go func() {
		result := StoreResult{}

		if _, err := s.GetMaster().Exec(
			`DELETE FROM PostUserData WHERE Id = :Id`, map[string]interface{}{"Id": id}); err != nil {
			result.Err = model.NewAppError("SqlPostUserDataStore.Delete", "We encountered an error while deleteing a post user data", err.Error())
		}

		storeChannel <- result
		close(storeChannel)
	}()

	return storeChannel
}
