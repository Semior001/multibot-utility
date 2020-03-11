package groups

import (
	"encoding/json"
	"io/ioutil"
	"path"
	"testing"

	bolt "github.com/coreos/bbolt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBoltDB_PutGroup(t *testing.T) {
	svc := prepareBoltDB(t)

	users := []string{"@blah", "@blah1", "@blah2"}
	err := svc.PutGroup("foo", "@bar", users)
	require.NoError(t, err)

	err = svc.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(groupBotBktName))
		assert.NotNil(t, bkt)

		chat := bkt.Bucket([]byte("foo"))
		assert.NotNil(t, chat)

		members := chat.Get([]byte("@bar"))
		assert.NotNil(t, members)
		assert.NotEmpty(t, members)

		err = json.Unmarshal(members, &users)
		require.NoError(t, err)

		assert.Contains(t, users, "@blah")
		assert.Contains(t, users, "@blah1")
		assert.Contains(t, users, "@blah2")
		return nil
	})
	require.NoError(t, err)
}

func TestBoltDB_DeleteUserFromGroup(t *testing.T) {
	svc := prepareBoltDB(t)

	users := []string{"@blah", "@blah1", "@blah2"}
	err := svc.db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(groupBotBktName))
		assert.NotNil(t, bkt)

		chat, err := bkt.CreateBucket([]byte("foo"))
		require.NoError(t, err)

		j, err := json.Marshal(users)
		require.NoError(t, err)

		err = chat.Put([]byte("@bar"), j)
		require.NoError(t, err)
		return nil
	})
	require.NoError(t, err)

	err = svc.DeleteUserFromGroup("foo", "@bar", "@blah1")
	require.NoError(t, err)

	err = svc.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(groupBotBktName))
		assert.NotNil(t, bkt)

		chat := bkt.Bucket([]byte("foo"))
		assert.NotNil(t, chat)

		j := chat.Get([]byte("@bar"))
		require.NoError(t, err)

		err = json.Unmarshal(j, &users)
		require.NoError(t, err)

		assert.Contains(t, users, "@blah")
		assert.Contains(t, users, "@blah2")
		assert.NotContains(t, users, "@blah1")
		return nil
	})

	err = svc.DeleteUserFromGroup("foo", "@bar", "@blah1")
	err = svc.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(groupBotBktName))
		assert.NotNil(t, bkt)

		chat := bkt.Bucket([]byte("foo"))
		assert.NotNil(t, chat)

		j := chat.Get([]byte("@bar"))
		require.NoError(t, err)

		err = json.Unmarshal(j, &users)
		require.NoError(t, err)

		assert.Contains(t, users, "@blah")
		assert.Contains(t, users, "@blah2")
		assert.NotContains(t, users, "@blah1")
		return nil
	})
}

func TestBoltDB_GetGroup(t *testing.T) {
	svc := prepareBoltDB(t)

	users := []string{"@blah", "@blah1", "@blah2"}
	err := svc.db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(groupBotBktName))
		assert.NotNil(t, bkt)

		chat, err := bkt.CreateBucket([]byte("foo"))
		require.NoError(t, err)

		j, err := json.Marshal(users)
		require.NoError(t, err)

		err = chat.Put([]byte("@bar"), j)
		require.NoError(t, err)
		return nil
	})
	require.NoError(t, err)

	q, err := svc.GetGroup("foo", "@bar")
	require.NoError(t, err)
	assert.Contains(t, q, "@blah")
	assert.Contains(t, q, "@blah1")
	assert.Contains(t, q, "@blah2")
}

func TestBoltDB_GetGroups(t *testing.T) {
	svc := prepareBoltDB(t)

	users := []string{"@blah", "@blah1", "@blah2"}
	err := svc.db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(groupBotBktName))
		assert.NotNil(t, bkt)

		chat, err := bkt.CreateBucketIfNotExists([]byte("foo"))
		require.NoError(t, err)

		j, err := json.Marshal(users)
		require.NoError(t, err)

		err = chat.Put([]byte("@bar"), j)
		require.NoError(t, err)

		err = chat.Put([]byte("@bar1"), j)
		require.NoError(t, err)

		err = chat.Put([]byte("@bar2"), j)
		require.NoError(t, err)
		return nil
	})
	require.NoError(t, err)

	groups, err := svc.GetGroups("foo")
	require.NoError(t, err)

	assert.Contains(t, groups, "@bar")
	assert.Contains(t, groups, "@bar1")
	assert.Contains(t, groups, "@bar2")

	assert.Contains(t, groups["@bar"], "@blah")
	assert.Contains(t, groups["@bar"], "@blah1")
	assert.Contains(t, groups["@bar"], "@blah2")

	assert.Contains(t, groups["@bar1"], "@blah")
	assert.Contains(t, groups["@bar1"], "@blah1")
	assert.Contains(t, groups["@bar1"], "@blah2")

	assert.Contains(t, groups["@bar2"], "@blah")
	assert.Contains(t, groups["@bar2"], "@blah1")
	assert.Contains(t, groups["@bar2"], "@blah2")
}

func TestBoltDB_DeleteGroup(t *testing.T) {
	svc := prepareBoltDB(t)

	users := []string{"@blah", "@blah1", "@blah2"}
	err := svc.db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(groupBotBktName))
		assert.NotNil(t, bkt)

		chat, err := bkt.CreateBucket([]byte("foo"))
		require.NoError(t, err)

		j, err := json.Marshal(users)
		require.NoError(t, err)

		err = chat.Put([]byte("@bar"), j)
		require.NoError(t, err)
		return nil
	})
	require.NoError(t, err)

	err = svc.DeleteGroup("foo", "@bar")
	require.NoError(t, err)

	err = svc.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(groupBotBktName))
		assert.NotNil(t, bkt)

		chat := bkt.Bucket([]byte("foo"))
		assert.NotNil(t, chat)

		q := chat.Get([]byte("@bar"))
		assert.Nil(t, q)
		return nil
	})
	require.NoError(t, err)
}

func TestBoltDB_AddUser(t *testing.T) {
	svc := prepareBoltDB(t)

	users := []string{"@blah", "@blah1", "@blah2"}
	err := svc.db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(groupBotBktName))
		assert.NotNil(t, bkt)

		chat, err := bkt.CreateBucket([]byte("foo"))
		require.NoError(t, err)

		j, err := json.Marshal(users)
		require.NoError(t, err)

		err = chat.Put([]byte("@bar"), j)
		require.NoError(t, err)
		return nil
	})
	require.NoError(t, err)

	err = svc.AddUser("foo", "@bar", "@blah3")

	err = svc.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(groupBotBktName))
		assert.NotNil(t, bkt)

		chat := bkt.Bucket([]byte("foo"))
		assert.NotNil(t, chat)

		j := chat.Get([]byte("@bar"))
		err = json.Unmarshal(j, &users)
		require.NoError(t, err)

		assert.Contains(t, users, "@blah")
		assert.Contains(t, users, "@blah1")
		assert.Contains(t, users, "@blah2")
		assert.Contains(t, users, "@blah3")
		return nil
	})
	require.NoError(t, err)
}

func TestBoltDB_FindAliases(t *testing.T) {
	svc := prepareBoltDB(t)

	users := []string{"@blah", "@blah1", "@blah2"}
	usersA := []string{"@blahA", "@blah1A", "@blah2A"}
	usersC := []string{"@blahC", "@blah1C", "@blah2C"}

	err := svc.db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(groupBotBktName))
		assert.NotNil(t, bkt)

		chatBkt, err := bkt.CreateBucket([]byte("foo"))
		require.NoError(t, err)
		assert.NotNil(t, chatBkt)

		j, err := json.Marshal(users)
		require.NoError(t, err)

		err = chatBkt.Put([]byte("@users"), j)
		require.NoError(t, err)

		j, err = json.Marshal(usersA)
		require.NoError(t, err)

		err = chatBkt.Put([]byte("@usersA"), j)
		require.NoError(t, err)

		j, err = json.Marshal(usersC)
		require.NoError(t, err)

		err = chatBkt.Put([]byte("@usersC"), j)
		require.NoError(t, err)
		return nil
	})
	require.NoError(t, err)

	queried, err := svc.FindAliases("foo", []string{"@usersA", "@usersC"})
	require.NoError(t, err)

	for _, u := range usersA {
		assert.Contains(t, queried, u)
	}

	for _, u := range usersC {
		assert.Contains(t, queried, u)
	}

	for _, u := range users {
		assert.NotContains(t, queried, u)
	}
}

func TestBoltDB_Unique(t *testing.T) {
	queried := unique([]string{"@blah", "@blah1", "@blah", "@blah1", "@blah3"})
	m := make(map[string]int)
	for _, s := range queried {
		if _, ok := m[s]; !ok {
			m[s] = 0
		}
		m[s]++
	}
	for _, cnt := range m {
		assert.Equal(t, cnt, 1)
	}
}

func prepareBoltDB(t *testing.T) *BoltDB {
	loc, err := ioutil.TempDir("", "test_groups_multibot")
	require.NoError(t, err, "failed to make temp dir")

	svc, err := NewBoltDB(path.Join(loc, "groups_bot_test.db"), bolt.Options{})
	require.NoError(t, err, "New bolt storage")
	return svc
}
