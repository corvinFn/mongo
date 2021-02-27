package mongo

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/globalsign/mgo"
	"github.com/pkg/errors"
)

const dbWorkerMax = 8000

// Client ...
type Client struct {
	init             sync.Once
	session          *mgo.Session
	worker           chan struct{}
	addresses        []string
	err              error
	mode             mgo.Mode
	batchSize        int
	ensureReplicated bool
}

func (c *Client) doDial() error {
	c.init.Do(func() {
		if len(c.addresses) == 0 {
			c.err = errors.New("addresses empty")
		}
		addrUrl := fmt.Sprintf("mongodb://%s?connect=direct", strings.Join(c.addresses, ","))
		session, err := mgo.Dial(addrUrl)
		if err != nil {
			c.err = err
			return
		}
		c.session = session
		c.worker = make(chan struct{}, dbWorkerMax)
		c.refreshSetting()
	})
	return c.err
}

// Open ...
func (c *Client) Open(db, collection string) *Collection {
	if err := c.doDial(); err != nil {
		log.Fatal("Failed to dial", db, collection)
	}
	select {
	case c.worker <- struct{}{}:
	default:
		log.Fatal("Mongodb opened too many connections.", db, collection)
	}
	session := c.session.Copy()
	colle := session.DB(db).C(collection)
	return &Collection{
		Collection: colle,
		session:    session,
		client:     c,
	}
}

// OpenWithLongTimeout ...
func (c *Client) OpenWithLongTimeout(db, collection string) *Collection {
	colle := c.Open(db, collection)
	colle.session.SetSocketTimeout(time.Minute * 30)
	colle.session.SetCursorTimeout(0)
	return colle
}

func (c *Client) refreshSetting() {
	if c.session == nil {
		return
	}
	if c.batchSize > 0 {
		c.session.SetMode(c.mode, true)
	}
	if c.ensureReplicated {
		c.session.SetSafe(&mgo.Safe{W: len(c.addresses)})
	} else {
		c.session.SetSafe(&mgo.Safe{})
	}
}

func (c *Client) SetBatch(n int) {
	c.batchSize = n
	c.refreshSetting()
}

func (c *Client) SetMode(mode mgo.Mode) {
	c.mode = mode
	c.refreshSetting()
}

// EnsurReplicated make sure all data written are replicated to all secondary servers.
func (c *Client) EnsureReplicated() {
	c.ensureReplicated = true
	c.refreshSetting()
}

// Collection ...
type Collection struct {
	*mgo.Collection
	session   *mgo.Session
	client    *Client
	closeOnce sync.Once
}

// Close session
func (colle *Collection) Close() {
	colle.closeOnce.Do(func() {
		colle.session.Close()
		<-colle.client.worker
	})
}

// WithContext horors input context, and close session when context is done.
func (colle *Collection) WithContext(ctx context.Context) *Collection {
	go func() {
		select {
		case <-ctx.Done():
			colle.Close()
		}
	}()
	return colle
}

func (colle *Collection) SetBatch(n int) *Collection {
	colle.session.SetBatch(n)
	return colle
}

func (colle *Collection) SetMode(mode mgo.Mode) *Collection {
	colle.session.SetMode(mode, true)
	return colle
}

// EnsureReplicated make sure all data written are replicated to all secondary servers.
func (colle *Collection) EnsureReplicated() *Collection {
	colle.session.SetSafe(&mgo.Safe{W: len(colle.client.addresses)})
	return colle
}
