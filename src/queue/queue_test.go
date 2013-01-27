package queue

import (
	"document"
	. "launchpad.net/gocheck"
	"posting"
	"strings"
	"testing"
	"testutils"
	"time"
)

func Test(t *testing.T) { TestingT(t) }

type QuerySuite struct {
	testutils.DBSuite
}

var _ = Suite(&QuerySuite{})

func (s *QuerySuite) TestQueue(c *C) {
	go posting.Serve(s.Registry)
	go Start(s.Registry)
	for i := uint32(1); i <= 20; i++ {
		target := document.DocumentID{Doctype: 1, Docid: i}
		item, err := NewQueueItem(s.Registry, "Add Document", nil, &target, nil, nil, strings.NewReader("title=Payload&text=PayloadWithsometextlongerthanwindowsize"))
		c.Check(item, NotNil)
		c.Check(err, IsNil)
	}
	time.Sleep(4 * time.Second) // Nasty!
	stats, err := Stats(s.Registry)
	c.Check(err, IsNil)
	c.Check(stats["Completed"], Equals, 20)
	c.Check(stats["Queued"], Equals, 0)
	c.Check(stats["Failed"], Equals, 0)
	count, err := s.Registry.C("documents").Count()
	c.Check(err, IsNil)
	c.Check(count, Equals, 20)
	for i := uint32(1); i <= 20; i++ {
		target := document.DocumentID{Doctype: 1, Docid: i}
		item, err := NewQueueItem(s.Registry, "Delete Document", nil, &target, nil, nil, strings.NewReader(""))
		c.Check(item, NotNil)
		c.Check(err, IsNil)
	}
	time.Sleep(4 * time.Second) // Nasty!
	stats, err = Stats(s.Registry)
	c.Check(err, IsNil)
	c.Check(stats["Completed"], Equals, 40)
	c.Check(stats["Queued"], Equals, 0)
	c.Check(stats["Failed"], Equals, 0)
	count, err = s.Registry.C("documents").Count()
	c.Check(err, IsNil)
	c.Check(count, Equals, 0)
}

func (s *QuerySuite) TestPayload(c *C) {
	go Start(s.Registry)
	go posting.Serve(s.Registry)
	item, err := NewQueueItem(s.Registry, "test", nil, nil, nil, nil, strings.NewReader("I am the payload"))
	c.Check(err, IsNil)
	var q QueueItem
	s.Registry.C("queue").FindId(item.Id).One(&q)
	c.Check(q.Payload, NotNil)
	p, err := q.getPayload()
	c.Check(p, Equals, "I am the payload")
	c.Check(err, IsNil)
}
