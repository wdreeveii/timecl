package logger

import (
	"container/list"
	"github.com/coopernurse/gorp"
	"io/ioutil"
	"log"
	"os"
	"time"
)

var LOG = log.New(os.Stderr, "Logger ", log.Ldate|log.Ltime)
var DEBUG = log.New(ioutil.Discard, "Logger ", log.Ldate|log.Ltime)

type LoggingData struct {
	Timestamp int64
	ObjectId  int
	Min       float64
	Max       float64
	Avg       float64

	//transient
	Time time.Time
}

func InitLoggerTables(dbm *gorp.DbMap) {
	t := dbm.AddTable(LoggingData{}).SetKeys(false, "Timestamp", "ObjectId")
	t.ColMap("Time").Transient = true
}

func (d *LoggingData) PreInsert(_ gorp.SqlExecutor) error {
	d.Timestamp = d.Time.Unix()
	return nil
}

type Subscription struct {
	New <-chan Event
}

type Event struct {
	Type string
	Data interface{}
}

var (
	subscribe   = make(chan (chan<- Subscription), 10)
	unsubscribe = make(chan (<-chan Event), 10)
	publish     = make(chan Event, 100)
)

func Subscribe() Subscription {
	resp := make(chan Subscription)
	subscribe <- resp
	return <-resp
}

func (s Subscription) Cancel() {
	unsubscribe <- s.New
	drain(s.New)
}

func Publish(event Event) {
	publish <- event
}

func logger_pub_sub(dbm *gorp.DbMap) {
	subscribers := list.New()

	for {
		select {
		case ch := <-subscribe:
			subscriber := make(chan Event, 100)
			subscribers.PushBack(subscriber)
			ch <- Subscription{subscriber}

		case event := <-publish:
			if event.Type == "capture" {
				data, ok := event.Data.(LoggingData)
				if ok {
					err := dbm.Insert(&data)
					if err != nil {
						LOG.Println("Problem inserting capture data:", err)
					}
				} else {
					LOG.Println("Problem converting the data for storage.")
				}
			}
			for ch := subscribers.Front(); ch != nil; ch = ch.Next() {
				ch.Value.(chan Event) <- event
			}

		case unsub := <-unsubscribe:
			for ch := subscribers.Front(); ch != nil; ch = ch.Next() {
				if ch.Value.(chan Event) == unsub {
					subscribers.Remove(ch)
					break
				}
			}
		}
	}
}

func Init(dbm *gorp.DbMap) {
	InitLoggerTables(dbm)
	err := dbm.CreateTablesIfNotExists()
	if err != nil {
		panic(err)
	}
	go logger_pub_sub(dbm)
}

// Drains a given channel of any messages.
func drain(ch <-chan Event) {
	for {
		select {
		case _, ok := <-ch:
			if !ok {
				return
			}
		default:
			return
		}
	}
}
