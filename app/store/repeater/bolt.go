package repeater

import (
	"encoding/json"
	"log"
	"time"

	bolt "github.com/coreos/bbolt"
	"github.com/pkg/errors"
)

const (
	// top level buckets
	rulesBktName   = "rules"   // rules itself
	authorsBktName = "authors" // authors to rules references
	srcChatBktName = "src"     // source chat to rules references
)

// Bolt implements Store to access repeater rules
// There are 3 top-level buckets:
// - rules, which forms a k:v pair as ruleID:rule
// - source chat to rules references, which has sourceGroupID as key and value is nested bucket with k:v as ruleID:ts
// - authors to rules references, which has authorID as key and value is nested bucket with k:v as ruleID:ts
type Bolt struct {
	fileName string
	db       *bolt.DB
}

// NewBoltDB creates new repeater store
func NewBoltDB(fileName string, opts bolt.Options) (*Bolt, error) {
	db, err := bolt.Open(fileName, 0600, &opts)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to open boltdb at %s", fileName)
	}
	err = db.Update(func(tx *bolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists([]byte(rulesBktName)); err != nil {
			return errors.Wrap(err, "failed to create rules bucket")
		}
		if _, err := tx.CreateBucketIfNotExists([]byte(authorsBktName)); err != nil {
			return errors.Wrap(err, "failed to create authors bucket")
		}
		if _, err := tx.CreateBucketIfNotExists([]byte(srcChatBktName)); err != nil {
			return errors.Wrap(err, "failed to create src bucket")
		}
		return nil
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to initialize boltdb %s buckets", fileName)
	}

	log.Print("[INFO] repeater.Bolt instantiated")
	return &Bolt{
		fileName: fileName,
		db:       db,
	}, err
}

// Add rule to the storage
func (b *Bolt) Add(r Rule) error {
	return b.db.Update(func(tx *bolt.Tx) error {
		jdata, err := json.Marshal(r)
		if err != nil {
			return errors.Wrap(err, "failed to marshal")
		}

		// saving rule itself
		if err := tx.Bucket([]byte(rulesBktName)).Put([]byte(r.ID), jdata); err != nil {
			return errors.Wrap(err, "failed to put entry to rules bucket")
		}

		// making ts to save reference
		ts := time.Now().Format(time.RFC3339Nano)

		// saving author reference to rule
		authBkt, err := b.makeAuthorBkt(tx, r.Author)
		if err != nil {
			return errors.Wrapf(err, "can't get bucket to make reference for author %s and rule %s", r.Author, r.ID)
		}
		if err := authBkt.Put([]byte(r.ID), []byte(ts)); err != nil {
			return errors.Wrap(err, "failed to put entry to authors bucket")
		}

		// saving source chat reference to rule
		srcChatBkt, err := b.makeSourceChatBkt(tx, r.Src)
		if err != nil {
			return errors.Wrapf(err, "can't get bucket to make reference for src chat %s and rule %s", r.Src, r.ID)
		}
		if err := srcChatBkt.Put([]byte(r.ID), []byte(ts)); err != nil {
			return errors.Wrap(err, "failed to put entry to src bucket")
		}

		return nil
	})
}

// Find rules in the storage by the given request
func (b *Bolt) Find(req FindReq) ([]Rule, error) {
	switch {
	case req.Src != "" && req.Author == "": // find only by source chat ID
		return b.listRulesBySrc(req.Src)
	case req.Src == "" && req.Author != "": // find only by author ID
		return b.listRulesByAuthor(req.Author)
	}
	return nil, errors.Errorf("wrong find request for %+v", req)
}

// listRulesBySrc returns list of rules that has source of repeating data as
// the given srcChatID
func (b *Bolt) listRulesBySrc(srcChatID string) ([]Rule, error) {
	var rules []Rule
	err := b.db.View(func(tx *bolt.Tx) error {
		// listing references to rules by the given srcChatID
		srcChatBkt, err := b.getSourceChatBkt(tx, srcChatID)
		if err != nil {
			return errors.Wrapf(err, "can't get %s source chat bucket", srcChatID)
		}
		var ruleIDs []string
		err = srcChatBkt.ForEach(func(k, _ []byte) error {
			ruleIDs = append(ruleIDs, string(k))
			return nil
		})
		if err != nil {
			return errors.Wrapf(err, "failed to list %s srcChat to rules references", srcChatID)
		}

		// listing rules itself
		rules, err = b.getRules(tx, ruleIDs)
		if err != nil {
			return errors.Wrap(err, "can't list rules")
		}
		return nil
	})
	return rules, err
}

// listRulesByAuthor returns list of rules that was created by the given authors
func (b *Bolt) listRulesByAuthor(authorID string) ([]Rule, error) {
	var rules []Rule
	err := b.db.View(func(tx *bolt.Tx) error {
		// listing references to rules by the given authorID
		authorBkt, err := b.getAuthorBkt(tx, authorID)
		if err != nil {
			return errors.Wrapf(err, "can't get %s author bucket", authorID)
		}
		var ruleIDs []string
		err = authorBkt.ForEach(func(k, _ []byte) error {
			ruleIDs = append(ruleIDs, string(k))
			return nil
		})
		if err != nil {
			return errors.Wrapf(err, "failed to list %s author to rules references", authorID)
		}

		// listing rules itself
		rules, err = b.getRules(tx, ruleIDs)
		if err != nil {
			return errors.Wrap(err, "can't list rules")
		}
		return nil
	})
	return rules, err
}

// Delete rule from the storage by given ID
func (b *Bolt) Delete(id string) error {
	return b.db.Update(func(tx *bolt.Tx) error {
		rulesBkt := tx.Bucket([]byte(rulesBktName))

		// taking rule itself
		var rule Rule
		jdata := rulesBkt.Get([]byte(id))
		if err := json.Unmarshal(jdata, &rule); err != nil {
			return errors.Wrap(err, "failed to unmarshal entry")
		}

		// removing from the rules bucket
		if err := rulesBkt.Delete([]byte(id)); err != nil {
			return errors.Wrap(err, "failed to remove entry from rules bucket")
		}

		// removing references
		authorBkt, err := b.getAuthorBkt(tx, rule.Author)
		if err != nil {
			return errors.Wrapf(err, "failed to remove author reference, can't get %s author bucket", rule.Author)
		}
		if err := authorBkt.Delete([]byte(rule.ID)); err != nil {
			return errors.Wrapf(err, "failed to remove entry from %s author bucket", rule.Author)
		}

		srcChatBkt, err := b.getSourceChatBkt(tx, rule.Src)
		if err != nil {
			return errors.Wrapf(err, "failed to remove src chat reference, can't get %s src chat bucket", rule.Src)
		}
		if err := srcChatBkt.Delete([]byte(rule.ID)); err != nil {
			return errors.Wrapf(err, "failed to remove entry from %s src chat bucket", rule.Src)
		}

		return nil
	})
}

// getRules lists rules by the given list of IDs
func (b *Bolt) getRules(tx *bolt.Tx, ruleIDs []string) ([]Rule, error) {
	var rules []Rule

	// listing rules for selected IDs
	rulesBkt := tx.Bucket([]byte(rulesBktName))
	var rule Rule
	for _, rID := range ruleIDs {
		jdata := rulesBkt.Get([]byte(rID))
		if err := json.Unmarshal(jdata, &rule); err != nil {
			return nil, errors.Wrapf(err, "failed to unmarshal entry with id %s", rID)
		}
		rules = append(rules, rule)
	}

	return rules, nil
}

func (b *Bolt) makeAuthorBkt(tx *bolt.Tx, authorID string) (*bolt.Bucket, error) {
	bkt, err := tx.Bucket([]byte(authorsBktName)).CreateBucketIfNotExists([]byte(authorID))
	return bkt, errors.Wrapf(err, "failed to get %s author bucket", authorID)
}

func (b *Bolt) getAuthorBkt(tx *bolt.Tx, authorID string) (*bolt.Bucket, error) {
	bkt := tx.Bucket([]byte(authorsBktName)).Bucket([]byte(authorID))
	if bkt == nil {
		return nil, errors.Errorf("no bucket %s in store", authorsBktName)
	}
	return bkt, nil
}

func (b *Bolt) makeSourceChatBkt(tx *bolt.Tx, srcChatID string) (*bolt.Bucket, error) {
	bkt, err := tx.Bucket([]byte(srcChatBktName)).CreateBucketIfNotExists([]byte(srcChatID))
	return bkt, errors.Wrapf(err, "failed to get %s source chat bucket", srcChatID)
}

func (b *Bolt) getSourceChatBkt(tx *bolt.Tx, srcChatID string) (*bolt.Bucket, error) {
	bkt := tx.Bucket([]byte(srcChatBktName)).Bucket([]byte(srcChatID))
	if bkt == nil {
		return nil, errors.Errorf("no bucket %s in store", srcChatBktName)
	}
	return bkt, nil
}
