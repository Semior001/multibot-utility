package groups

import (
	"encoding/json"
	"log"

	bolt "github.com/coreos/bbolt"
	"github.com/pkg/errors"
)

const groupBotBktName = "groupbot"

// BoltDB implements store to put and get groups with specific alias
type BoltDB struct {
	fileName string
	db       *bolt.DB
}

// NewBoltDB creates new groupbot store
func NewBoltDB(fileName string, opts bolt.Options) (*BoltDB, error) {
	log.Print("[INFO] groups.BoltDB instantiated")
	db, err := bolt.Open(fileName, 0600, &opts)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to open boltdb at %s", fileName)
	}
	err = db.Update(func(tx *bolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists([]byte(groupBotBktName)); err != nil {
			return errors.Wrap(err, "failed to create groupbot bucket")
		}
		return nil
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to initialize boltdb %s buckets", fileName)
	}
	return &BoltDB{
		fileName: fileName,
		db:       db,
	}, err
}

// GetGroups returns the list of groups by chatID in form map[group_alias][]users
func (b *BoltDB) GetGroups(chatID string) (map[string][]string, error) {
	res := make(map[string][]string)
	err := b.db.View(func(tx *bolt.Tx) error {
		chatBkt := tx.Bucket([]byte(groupBotBktName)).Bucket([]byte(chatID))
		if chatBkt == nil {
			return errors.Wrapf(
				errors.New("bucket does not exist"),
				"failed to get groups of chat %s", chatID,
			)
		}
		err := chatBkt.ForEach(func(k, v []byte) error {
			var u []string
			err := json.Unmarshal(v, &u)
			if err != nil {
				return errors.Wrapf(err, "failed to get groups of chat %s", chatID)
			}
			res[string(k)] = u
			return nil
		})
		return errors.Wrapf(err, "failed to get groups of chat %s", chatID)
	})
	return res, err
}

// DeleteUserFromGroup removes user from the group
func (b *BoltDB) DeleteUserFromGroup(chatID string, alias string, user string) error {
	err := b.db.Update(func(tx *bolt.Tx) error {
		chatBkt := tx.Bucket([]byte(groupBotBktName)).Bucket([]byte(chatID))
		if chatBkt == nil {
			return errors.Wrapf(
				errors.New("group bucket does not exist"),
				"failed to delete user from group %s:%s", chatID, alias,
			)
		}

		data := chatBkt.Get([]byte(alias))
		if data == nil {
			return errors.Wrapf(
				errors.New("group does not exist"),
				"failed to delete user from group %s:%s", chatID, alias,
			)
		}
		var users []string
		err := json.Unmarshal(data, &users)
		if err != nil {
			return errors.Wrapf(err, "failed to delete user from group %s:%s", chatID, alias)
		}

		// looking for the user in the list
		idx := -1
		for i := range users {
			if users[i] == user {
				idx = i
			}
		}
		if idx == -1 {
			// if user does not exist in the list - we just do nothing
			return nil
		}

		// removing user from the list
		users = append(users[:idx], users[idx+1:]...)
		data, err = json.Marshal(users)
		if err != nil {
			return errors.Wrapf(err, "failed to delete user from group %s:%s", chatID, alias)
		}

		// replacing list inside the bucket
		err = chatBkt.Put([]byte(alias), data)
		if err != nil {
			return errors.Wrapf(err, "failed to delete user from group %s:%s", chatID, alias)
		}
		return nil
	})
	return err
}

// AddUser adds user to the specified group
func (b *BoltDB) AddUser(chatID string, alias string, user string) error {
	err := b.db.Update(func(tx *bolt.Tx) error {
		chatBkt := tx.Bucket([]byte(groupBotBktName)).Bucket([]byte(chatID))
		if chatBkt == nil {
			return errors.Wrapf(
				errors.New("chat bucket does not exist"),
				"failed to add user to group %s:%s", chatID, alias,
			)
		}

		data := chatBkt.Get([]byte(alias))
		if data == nil {
			return errors.Wrapf(
				errors.New("group does not exist"),
				"failed to add user to group %s:%s", chatID, alias,
			)
		}

		var users []string
		err := json.Unmarshal(data, &users)
		if err != nil {
			return errors.Wrapf(err, "failed to add user to group %s:%s", chatID, alias)
		}

		users = append(users, user)
		data, err = json.Marshal(users)
		if err != nil {
			return errors.Wrapf(err, "failed to add user to group %s:%s", chatID, alias)
		}

		err = chatBkt.Put([]byte(alias), data)
		if err != nil {
			return errors.Wrapf(err, "failed to add user to group %s:%s", chatID, alias)
		}
		return nil
	})
	return err
}

// DeleteGroup removes group from the database by given chatID
func (b *BoltDB) DeleteGroup(chatID string, alias string) error {
	err := b.db.Update(func(tx *bolt.Tx) error {
		chatBkt := tx.Bucket([]byte(groupBotBktName)).Bucket([]byte(chatID))
		if chatBkt == nil {
			return errors.Wrapf(
				errors.New("chat bucket does not exist"),
				"failed to delete group %s:%s", chatID, alias,
			)
		}

		err := chatBkt.Delete([]byte(alias))
		if err != nil {
			return errors.Wrapf(err, "failed to delete group %s:%s", chatID, alias)
		}
		return nil
	})
	return err
}

// PutGroup creates a new group in the database with specified users
func (b *BoltDB) PutGroup(chatID string, alias string, users []string) error {
	err := b.db.Update(func(tx *bolt.Tx) error {
		chatBkt, err := tx.Bucket([]byte(groupBotBktName)).CreateBucketIfNotExists([]byte(chatID))
		if err != nil {
			return errors.Wrapf(err, "failed to put group %s:%s into bucket", chatID, alias)
		}

		j, err := json.Marshal(users)
		if err != nil {
			return errors.Wrapf(
				errors.Wrapf(err, "failed to marshal users list"),
				"failed to put group %s:%s into bucket", chatID, alias,
			)
		}

		err = chatBkt.Put([]byte(alias), j)
		if err != nil {
			return errors.Wrapf(err, "failed to put group %s:%s into bucket", chatID, alias)
		}
		return nil
	})
	return err
}

// GetGroup returns all users of the single group
func (b *BoltDB) GetGroup(chatID string, alias string) ([]string, error) {
	var users []string
	err := b.db.View(func(tx *bolt.Tx) error {
		chatBkt := tx.Bucket([]byte(groupBotBktName)).Bucket([]byte(chatID))
		if chatBkt == nil {
			return errors.Errorf("failed to get users of group %s:%s", chatID, alias)
		}
		data := chatBkt.Get([]byte(alias))
		if data == nil {
			return errors.Wrapf(
				errors.New("group does not exist"),
				"failed to get users of group %s:%s", chatID, alias,
			)
		}
		err := json.Unmarshal(data, &users)
		if err != nil {
			return errors.Wrapf(err, "failed to get users of group %s:%s", chatID, alias)
		}
		return nil
	})
	return users, err
}

// FindAliases looks for group aliases in the database
// and returns members of groups if group alias is present
func (b *BoltDB) FindAliases(chatID string, aliases []string) ([]string, error) {
	var users []string
	err := b.db.View(func(tx *bolt.Tx) error {
		chatBkt := tx.Bucket([]byte(groupBotBktName)).Bucket([]byte(chatID))
		if chatBkt == nil {
			return errors.Wrapf(
				errors.New("chat bucket does not exist"),
				"error while looking for aliases of chat %s in boltdb", chatID,
			)
		}
		// looking for aliases in chat bucket
		for _, alias := range aliases {
			group := chatBkt.Get([]byte(alias))
			// this alias is not a group, skip
			if group == nil {
				continue
			}
			var members []string
			err := json.Unmarshal(group, &members)
			if err != nil {
				return errors.Wrapf(err, "error while looking for aliases of chat %s in boltdb", chatID)
			}
			users = append(users, members...)
		}
		return nil
	})
	return unique(users), err
}

// AddChat creates a chat bucket in the storage
func (b *BoltDB) AddChat(id string) error {
	err := b.db.Update(func(tx *bolt.Tx) error {
		_, err := tx.Bucket([]byte(groupBotBktName)).CreateBucketIfNotExists([]byte(id))
		if err != nil {
			return errors.Wrapf(err, "failed to create chat bucket with chatId %s", id)
		}
		return nil
	})
	return err
}

// unique returns slice of unique string occurrences from the source one
func unique(sl []string) []string {
	m := make(map[string]struct{})
	for _, s := range sl {
		m[s] = struct{}{}
	}
	var res []string
	for s := range m {
		res = append(res, s)
	}
	return res
}
