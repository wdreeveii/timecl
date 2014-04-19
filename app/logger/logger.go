package logger

import (
	"container/list"
	"database/sql"
	"errors"
	"fmt"
	"github.com/coopernurse/gorp"
	"math"
	"net/smtp"
	"strings"
	"time"
)

type Email struct {
	Addr        string
	Port        string
	SSL         string
	Username    string
	Password    string
	AuthType    string
	MaxMsgs     string
	MaxDuration string
}

type EmailProvider interface {
	ProcessEmailRateLimits(email_settings Email) (int, int, error)
	GetEmail(txn *gorp.Transaction) (Email, error)
}

type LogStream interface {
	Printf(format string, v ...interface{})
}

type LogManager struct {
	log_prefix string
	log_output LogStream
	dbm        *gorp.DbMap
	email_info EmailProvider
}

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

type ErrorData struct {
	Error          string
	Count          int64
	Timestamp      int64
	FirstTimestamp int64

	//transient
	Time  time.Time
	First time.Time
}
type ErrInfo struct {
	Count int64
	Time  time.Time
	First time.Time
}

type ErrorSlice []*ErrorListElement
type ErrorListElement struct {
	Error string
	Count int64
	Time  time.Time
	First time.Time
}

func (e ErrorSlice) Len() int {
	return len(e)
}
func (e ErrorSlice) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
}
func (e ErrorSlice) Less(i, j int) bool {
	return e[i].Time.Before(e[j].Time)
}

func initLoggerTables(dbm *gorp.DbMap) {
	t := dbm.AddTable(LoggingData{}).SetKeys(false, "Timestamp", "ObjectId")
	t.ColMap("Time").Transient = true
	t = dbm.AddTable(AlertData{}).SetKeys(true, "Id")
	t.ColMap("Time").Transient = true
	t = dbm.AddTable(ErrorData{}).SetKeys(false, "Error")
	t.ColMap("Time").Transient = true
	t.ColMap("First").Transient = true

	err := dbm.CreateTablesIfNotExists()
	if err != nil {
		panic(err)
	}
}

func (d *LoggingData) PreInsert(_ gorp.SqlExecutor) error {
	d.Timestamp = d.Time.Unix()
	return nil
}

func (d *AlertData) PreInsert(_ gorp.SqlExecutor) error {
	d.Timestamp = d.Time.Unix()
	return nil
}

func (d *ErrorData) PreInsert(_ gorp.SqlExecutor) error {
	d.Timestamp = d.Time.Unix()
	d.FirstTimestamp = d.First.Unix()
	return nil
}

func (d *ErrorData) PostGet(_ gorp.SqlExecutor) error {
	d.Time = time.Unix(d.Timestamp, 0)
	d.First = time.Unix(d.FirstTimestamp, 0)
	return nil
}
func sendOneEmail(data AlertData, email_settings Email) {
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
		PublishOneError(fmt.Errorf("Problem sending email: %v", err))
	}
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

func PublishOneError(err error) {
	var list ErrorSlice
	list = append(list, &ErrorListElement{Error: err.Error(),
		Count: 1,
		Time:  time.Now(),
		First: time.Now()})
	var event = Event{Type: "errors", Data: list}
	publish <- event
}

func Publish(event Event) {
	publish <- event
}

func saveErrorList(txn *gorp.Transaction, errlist ErrorSlice) error {
	var insert_q = `
INSERT INTO ErrorData (Error, Count, Timestamp, FirstTimestamp) VALUES (?, ?, ?, ?)
`
	var update_q = `
UPDATE ErrorData SET Timestamp = ?, Count = Count + ? WHERE Error = ?
`
	for _, v := range errlist {
		_, err := txn.Exec(insert_q, v.Error, v.Count, v.Time.Unix(), v.First.Unix())
		if err != nil {
			_, err = txn.Exec(update_q, v.Time.Unix(), v.Count, v.Error)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

const MAXINTERVAL = 10

func (m *LogManager) saveTrends(idata interface{}) {
	data, ok := idata.(LoggingData)
	if ok {
		txn, err := m.dbm.Begin()
		if err != nil {
			go PublishOneError(fmt.Errorf("Could not initiate database transaction to store trend data: %v", err))
		} else {
			err = txn.Insert(&data)
			if err != nil {
				go PublishOneError(fmt.Errorf("Could not insert trend data:", err))
			}
		}
		if txn != nil {
			err = txn.Commit()
			if err != nil {
				go PublishOneError(fmt.Errorf("Could not store trend data:", err))
			}
		}
	} else {
		go PublishOneError(errors.New("Problem converting logging data to storage format."))
	}
}

func (m *LogManager) trace(args ...interface{}) {
	if m.log_output != nil {
		m.log_output.Printf("%s %v", m.log_prefix, args)
	}
}

func (m *LogManager) saveErrors(idata interface{}) {
	data, ok := idata.(ErrorSlice)
	if ok {
		if len(data) > 0 {
			txn, err := m.dbm.Begin()
			if err != nil {
				m.trace("Could not initiate database transaction to store errors:", err)
			} else {
				err := saveErrorList(txn, data)
				if err != nil {
					m.trace("Could not save errors:", err)
				}
			}
			if txn != nil {
				if err := txn.Commit(); err != nil && err != sql.ErrTxDone {
					m.trace("Could not commit db transaction:", err)
				}
			}
		}

	} else {
		m.trace("Problem converting error data for storage.")
	}
}

func (m *LogManager) doAlert(buffer_email chan AlertData, buffering_enabled bool, email_settings Email, idata interface{}) {
	data, ok := idata.(AlertData)
	if ok {
		go func(dbm *gorp.DbMap, data AlertData) {
			txn, err := m.dbm.Begin()
			if err != nil {
				PublishOneError(fmt.Errorf("Could not create db transaction to save alert data."))
			} else {
				err = txn.Insert(&data)
				if err != nil {
					PublishOneError(fmt.Errorf("Could not insert alert into log:", err))
				} else {
					err = txn.Commit()
					if err != nil {
						PublishOneError(fmt.Errorf("Could not log alert:", err))
					}
				}
			}
		}(m.dbm, data)
		if email_settings.Addr != "" {
			if buffering_enabled {
				buffer_email <- data
			} else {
				go sendOneEmail(data, email_settings)
			}
		}
	} else {
		go PublishOneError(errors.New("Problem converting alert data for storage."))
	}
}

func (m *LogManager) Run() {
	var buffer_email = make(chan AlertData)
	subscribers := list.New()
	var email_settings Email
	var email_buffer []AlertData
	var email_ticker <-chan time.Time

	txn, err := m.dbm.Begin()
	if err != nil {
		go PublishOneError(fmt.Errorf("Problem getting email settings: %v", err))
	} else {
		email_settings, err = m.email_info.GetEmail(txn)
		if err != nil {
			go PublishOneError(fmt.Errorf("Problem getting email settings: %v", err))
		}
		maxmsgs, maxduration, err := m.email_info.ProcessEmailRateLimits(email_settings)
		if err != nil {
			go PublishOneError(fmt.Errorf("There was a problem parsing email rate limit parameters: %v", err))
		}
		if maxmsgs != 0 && maxduration != 0 {
			seconds_between_msg := time.Duration(math.Ceil(float64(maxduration) / float64(maxmsgs)))
			if seconds_between_msg > 0 {
				email_ticker = time.Tick(seconds_between_msg * time.Second)
			}
		}
	}
	if txn != nil {
		if err := txn.Commit(); err != nil && err != sql.ErrTxDone {
			go PublishOneError(fmt.Errorf("Problem getting email settings: %v", err))
		}
	}
	for {
		select {
		case ch := <-subscribe:
			subscriber := make(chan Event, 100)
			subscribers.PushBack(subscriber)
			ch <- Subscription{subscriber}

		case event := <-publish:
			for ch := subscribers.Front(); ch != nil; ch = ch.Next() {
				ch.Value.(chan Event) <- event
			}
			switch event.Type {
			case "capture":
				m.saveTrends(event.Data)
			case "errors":
				m.saveErrors(event.Data)
			case "alert":
				m.doAlert(buffer_email, email_ticker != nil, email_settings, event.Data)
			}
		case alert := <-buffer_email:
			email_buffer = append(email_buffer, alert)
		case <-email_ticker:
			if len(email_buffer) > 0 {
				email := email_buffer[0]
				email_buffer = email_buffer[1:]
				go sendOneEmail(email, email_settings)
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

func (m *LogManager) ErrorOn(prefix string, logger LogStream) {
	m.log_output = logger
	if len(prefix) > 0 {
		m.log_prefix = prefix + " "
	}
}

func Init(dbm *gorp.DbMap, email_settings EmailProvider) *LogManager {
	initLoggerTables(dbm)

	var l = LogManager{dbm: dbm, email_info: email_settings}
	return &l
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
