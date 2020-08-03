package repeater

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path"
	"testing"
	"time"

	bolt "github.com/coreos/bbolt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBolt_Add(t *testing.T) {
	svc := prepareBoltDB(t)
	expected := Rule{
		ID:     "foo",
		Src:    "srcGroup",
		Re:     "[a-zA-Z0-9]",
		Dest:   "destGroup",
		Author: "semior001",
	}

	err := svc.Add(expected)
	require.NoError(t, err)

	err = svc.db.View(func(tx *bolt.Tx) error {
		var actual Rule
		jdata := tx.Bucket([]byte(rulesBktName)).Get([]byte(expected.ID))
		require.NoError(t, json.Unmarshal(jdata, &actual))
		assert.Equal(t, expected, actual, "rules entry")

		srcChatBkt := tx.Bucket([]byte(srcChatBktName)).Bucket([]byte(expected.Src))
		assert.NotNil(t, srcChatBkt, "src chat bucket is nil")
		ts := srcChatBkt.Get([]byte(expected.ID))
		assert.NotNil(t, ts, "src chat to rule reference not exists")

		authorBkt := tx.Bucket([]byte(authorsBktName)).Bucket([]byte(expected.Author))
		assert.NotNil(t, authorsBktName, "author bucket is nil")
		ts = authorBkt.Get([]byte(expected.ID))
		assert.NotNil(t, ts, "author to rule reference not exists")

		return nil
	})
	require.NoError(t, err)
}

func TestBolt_List(t *testing.T) {
	svc := prep(t)

	// checking from the repeater position
	rules, err := svc.Find(FindReq{Src: "srcGroup"})
	require.NoError(t, err)
	assert.ElementsMatch(t, []Rule{
		{
			ID:     "foo",
			Src:    "srcGroup",
			Re:     "[a-zA-Z0-9]",
			Dest:   "destGroup",
			Author: "semior001",
		},
		{
			ID:     "bar",
			Src:    "srcGroup",
			Re:     "[a-zA-Z0-9]",
			Dest:   "destGroup1",
			Author: "someuser",
		},
	}, rules)

	// checking from the user position
	rules, err = svc.Find(FindReq{Author: "semior001"})
	require.NoError(t, err)
	assert.ElementsMatch(t, []Rule{
		{
			ID:     "foo",
			Src:    "srcGroup",
			Re:     "[a-zA-Z0-9]",
			Dest:   "destGroup",
			Author: "semior001",
		},
		{
			ID:     "foo1",
			Src:    "srcGroup1",
			Re:     "[a-zA-Z0-9]",
			Dest:   "destGroup",
			Author: "semior001",
		},
	}, rules)
}
func TestBolt_Delete(t *testing.T) {
	svc := prep(t)

	err := svc.Delete("foo")
	require.NoError(t, err)

	err = svc.db.View(func(tx *bolt.Tx) error {
		jdata := tx.Bucket([]byte(rulesBktName)).Get([]byte("foo"))
		require.Nil(t, jdata, "entry still exists in rules bucket")

		jdata = tx.Bucket([]byte(srcChatBktName)).Bucket([]byte("srcGroup")).Get([]byte("foo"))
		require.Nil(t, jdata, "reference still exists in src chat bucket")

		jdata = tx.Bucket([]byte(authorsBktName)).Bucket([]byte("semior001")).Get([]byte("foo"))
		require.Nil(t, jdata, "reference still exists in authors bucket")

		return nil
	})
	require.NoError(t, err)
}

func prep(t *testing.T) *Bolt {
	svc := prepareBoltDB(t)
	lst := []Rule{
		{
			ID:     "foo",
			Src:    "srcGroup",
			Re:     "[a-zA-Z0-9]",
			Dest:   "destGroup",
			Author: "semior001",
		},
		{
			ID:     "bar",
			Src:    "srcGroup",
			Re:     "[a-zA-Z0-9]",
			Dest:   "destGroup1",
			Author: "someuser",
		},
		{
			ID:     "foo1",
			Src:    "srcGroup1",
			Re:     "[a-zA-Z0-9]",
			Dest:   "destGroup",
			Author: "semior001",
		},
		{
			ID:     "bar1",
			Src:    "srcGroup2",
			Re:     "[a-zA-Z0-9]+",
			Dest:   "someDestGroup",
			Author: "someotheruser",
		},
	}

	err := svc.db.Update(func(tx *bolt.Tx) error {
		for _, rule := range lst {
			jdata, jerr := json.Marshal(rule)
			require.NoError(t, jerr)

			// put rule itself into rules bucket
			err := tx.Bucket([]byte(rulesBktName)).Put([]byte(rule.ID), jdata)
			require.NoError(t, err)

			// making references
			ts := time.Now().Format(time.RFC3339Nano)

			authorBkt, err := tx.Bucket([]byte(authorsBktName)).CreateBucketIfNotExists([]byte(rule.Author))
			require.NoError(t, err)
			err = authorBkt.Put([]byte(rule.ID), []byte(ts))
			require.NoError(t, err)

			srcChatBkt, err := tx.Bucket([]byte(srcChatBktName)).CreateBucketIfNotExists([]byte(rule.Src))
			require.NoError(t, err)
			err = srcChatBkt.Put([]byte(rule.ID), []byte(ts))
		}
		return nil
	})
	require.NoError(t, err)

	return svc
}

func prepareBoltDB(t *testing.T) *Bolt {
	loc, err := ioutil.TempDir("", "test_repeater_multibot")
	require.NoError(t, err, "failed to make temp dir")

	svc, err := NewBoltDB(path.Join(loc, "repeater_bot_test.db"), bolt.Options{})
	require.NoError(t, err, "New bolt storage")

	t.Cleanup(func() {
		assert.NoError(t, os.RemoveAll(loc))
	})
	return svc
}
