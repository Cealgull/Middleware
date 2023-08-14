package models

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAssetMarshal(t *testing.T) {

	asset := Asset{
		ContentType: "image/png",
		CID:         "QmQ6h7t2JhK4Z8JvJx3rZQFbDq1Q4Z6jZkQk9H8b2X5j2H",
	}

	var _, err = json.Marshal(&asset)
	assert.NoError(t, err)

}

func TestPostMarshal(t *testing.T) {

	post := Post{
		Content:       "Hello World",
		Hash:          "QmQ6h7t2JhK4Z8JvJx3rZQFbDq1Q4Z6jZkQk9H8b2X5j2H",
		CreatorWallet: "0x12345678",
		Creator: &User{
			Username: "test",
			Wallet:   "0x12345678",
			Avatar:   "null",
			Muted:    false,
			Banned:   false,
		},

		ReplyTo: &Post{
			Content: "Hello",
		},
		BelongTo: &Topic{
			Title: "test",
		},
	}

	var _, err = json.Marshal(&post)
	assert.NoError(t, err)

}

func TestTopicMarshal(t *testing.T) {

	topic := Topic{
		Title:         "test",
		CreatorWallet: "0x12345678",
		Creator: &User{
			Username: "test",
		},
	}

	var _, err = json.Marshal(&topic)
	assert.NoError(t, err)

}

func TestProfileMarshal(t *testing.T) {

	profile := Profile{
		Signature: "hello world",
		User:      &User{},
	}

	var _, err = json.Marshal(&profile)
	assert.NoError(t, err)

}

func TestUserMarshal(t *testing.T) {
  user := User{} 
  var _, err = json.Marshal(&user)
  assert.NoError(t, err)
}
