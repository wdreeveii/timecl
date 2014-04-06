package logger

import (
	"container/list"
	"database/sql"
	"github.com/coopernurse/gorp"
	"io/ioutil"
	"log"
	"math"
	"net/smtp"
	"os"
	"strings"
	"time"
	"timecl/app/models"
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

type AlertData struct {
	Id         int
	Timestamp  int64
	ObjectId   int
	Subject    string
	EventText  string
	Recipients string

	//transient
	Time time.Time
}

func InitLoggerTables(dbm *gorp.DbMap) {
	t := dbm.AddTable(LoggingData{}).SetKeys(false, "Timestamp", "ObjectId")
	t.ColMap("Time").Transient = true
	t = dbm.AddTable(AlertData{}).SetKeys(true, "Id")
	t.ColMap("Time").Transient = true
}

func (d *LoggingData) PreInsert(_ gorp.SqlExecutor) error {
	d.Timestamp = d.Time.Unix()
	return nil
}

func (d *AlertData) PreInsert(_ gorp.SqlExecutor) error {
	d.Timestamp = d.Time.Unix()
	return nil
}
func SendOneEmail(data AlertData, email_settings models.Email) {
	rec := strings.Split(data.Recipients, ",")
	for i := 0; i < len(rec); i++ {
		rec[i] = strings.TrimSpace(rec[i])
	}
	var authImplementation smtp.Auth
	if email_settings.AuthType == "PASSWORD" {
		authImplementation = smtp.PlainAuth("",
			email_settings.Username,
			email_settings.Password,
			email_settings.Addr)
	} else if email_settings.AuthType == "MD5" {
		authImplementation = smtp.CRAMMD5Auth(email_settings.Username,
			email_settings.Password)
	}
	body := "To: " + data.Recipients + "\r\nSubject: TIMECL: " +
		data.Subject + "\r\n\r\n" + data.EventText
	err := smtp.SendMail(email_settings.Addr+":"+email_settings.Port,
		authImplementation,
		"noreply@timecl.com",
		rec,
		[]byte(body))
	if err != nil {
		LOG.Println("Problem sending email:", err)
	}
}

var lastMsg time.Time
var bufferedAlerts = list.New()

func SendOneBufferedEmail(email_settings models.Email) <-chan time.Time {
	firstAlert := bufferedAlerts.Front()
	if firstAlert != nil {
		alert, ok := firstAlert.Value.(AlertData)
		if ok {
			go SendOneEmail(alert, email_settings)
			bufferedAlerts.Remove(firstAlert)
			lastMsg = time.Now()
		}
	}
	if bufferedAlerts.Len() > 0 {
		maxmsgs, maxduration, err := models.ProcessEmailRateLimits(email_settings)
		if err != nil {
			LOG.Println("There was a problem parsing email rate limit parameters.")
			return nil
		}
		if maxmsgs != 0 && maxduration != 0 {
			seconds_between_msg := time.Duration(math.Ceil(float64(maxduration) / float64(maxmsgs)))
			callback := time.After((lastMsg.Add(seconds_between_msg * time.Second)).Sub(time.Now()))
			return callback
		}
		return nil
	}
	return nil
}
func SendEmailRateLimited(data AlertData, email_settings models.Email) <-chan time.Time {
	maxmsgs, maxduration, err := models.ProcessEmailRateLimits(email_settings)
	if err != nil {
		LOG.Println("There was a problem parsing email rate limit parameters.")
	}
	if maxmsgs != 0 && maxduration != 0 {
		seconds_between_msg := time.Duration(math.Ceil(float64(maxduration) / float64(maxmsgs)))
		if time.Now().After(lastMsg.Add(seconds_between_msg * time.Second)) {
			firstAlert := bufferedAlerts.Front()
			if firstAlert != nil {
				alert, ok := firstAlert.Value.(AlertData)
				if ok {
					go SendOneEmail(alert, email_settings)
					bufferedAlerts.Remove(firstAlert)
					lastMsg = time.Now()
				}
			}
		} else {
			bufferedAlerts.PushBack(data)
		}
		if bufferedAlerts.Len() > 0 {
			callback := time.After((lastMsg.Add(seconds_between_msg * time.Second)).Sub(time.Now()))
			return callback
		} else {
			return nil
		}
	} else {
		go SendOneEmail(data, email_settings)
	}
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
	var rate_limited_email_ready <-chan time.Time
	subscribers := list.New()

	for {
		select {
		case ch := <-subscribe:
			subscriber := make(chan Event, 100)
			subscribers.PushBack(subscriber)
			ch <- Subscription{subscriber}

		case event := <-publish:
			switch event.Type {
			case "capture":
				data, ok := event.Data.(LoggingData)
				if ok {
					err := dbm.Insert(&data)
					if err != nil {
						LOG.Println("Problem inserting logging data:", err)
					}
				} else {
					LOG.Println("Problem converting logging data for storage.")
				}
			case "alert":
				data, ok := event.Data.(AlertData)
				if ok {
					err := dbm.Insert(&data)
					if err != nil {
						LOG.Println("Problem inserting alert data:", err)
					}
					txn, err := dbm.Begin()
					if err != nil {
						LOG.Println("Problem getting email settings:", err)
					} else {
						email_settings, err := models.GetEmail(txn)
						if err != nil {
							LOG.Println("Problem getting email settings:", err)
						}
						if email_settings.Addr != "" {
							rate_limited_email_ready = SendEmailRateLimited(data, email_settings)
						}
					}
					if txn != nil {
						if err := txn.Commit(); err != nil && err != sql.ErrTxDone {
							LOG.Println("Problem getting email settings:", err)
						}
					}
				} else {
					LOG.Println("Problem converting alert data for storage.")
				}
			}
			for ch := subscribers.Front(); ch != nil; ch = ch.Next() {
				ch.Value.(chan Event) <- event
			}
		case <-rate_limited_email_ready:
			txn, err := dbm.Begin()
			if err != nil {
				LOG.Println("Problem getting email settings:", err)
			} else {
				email_settings, err := models.GetEmail(txn)
				if err != nil {
					LOG.Println("Problem getting email settings:", err)
				}
				if email_settings.Addr != "" {
					rate_limited_email_ready = SendOneBufferedEmail(email_settings)
				}
			}
			if txn != nil {
				if err := txn.Commit(); err != nil && err != sql.ErrTxDone {
					LOG.Println("Problem getting email settings:", err)
				}
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
