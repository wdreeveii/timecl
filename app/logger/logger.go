package logger

import (
	"container/list"
	"database/sql"
	"errors"
	"fmt"
	"github.com/coopernurse/gorp"
	"github.com/revel/revel"
	"math"
	"net/smtp"
	"strings"
	"time"
	"timecl/app/models"
)

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

func InitLoggerTables(dbm *gorp.DbMap) {
	t := dbm.AddTable(LoggingData{}).SetKeys(false, "Timestamp", "ObjectId")
	t.ColMap("Time").Transient = true
	t = dbm.AddTable(AlertData{}).SetKeys(true, "Id")
	t.ColMap("Time").Transient = true
	t = dbm.AddTable(ErrorData{}).SetKeys(false, "Error")
	t.ColMap("Time").Transient = true
	t.ColMap("First").Transient = true
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

func SaveErrorList(txn *gorp.Transaction, errlist ErrorSlice) error {
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

func SaveTrends(dbm *gorp.DbMap, idata interface{}) {
	data, ok := idata.(LoggingData)
	if ok {
		txn, err := dbm.Begin()
		if err != nil {
			fmt.Println("begin failed:", err)
			go PublishOneError(fmt.Errorf("Could not initiate database transaction to store trend data: %v", err))
		} else {
			err = txn.Insert(&data)
			if err != nil {
				fmt.Println("errror:", err)
			}
		}
		if txn != nil {
			err = txn.Commit()
			if err != nil {
				fmt.Println("errrrrr:", err)
			}
		}
	} else {
		go PublishOneError(errors.New("Problem converting logging data to storage format."))
	}
}

func SaveErrors(dbm *gorp.DbMap, idata interface{}) {
	data, ok := idata.(ErrorSlice)
	for _, v := range data {
		revel.INFO.Printf("err: %+v\n", v)
	}
	if ok {
		if len(data) > 0 {
			txn, err := dbm.Begin()
			if err != nil {
				revel.INFO.Println(fmt.Errorf("Could not initiate database transaction to store errors: %v", err))
			} else {
				err := SaveErrorList(txn, data)
				if err != nil {
					fmt.Println("err errrr", err)
				}
			}
			if txn != nil {
				if err := txn.Commit(); err != nil && err != sql.ErrTxDone {
					revel.INFO.Println(fmt.Errorf("Could not commit db transaction: %v", err))
				}
			}
		}

	} else {
		revel.INFO.Println(errors.New("Problem converting error data for storage."))
	}
}

func DoAlert(dbm *gorp.DbMap, buffer_email chan AlertData, buffering_enabled bool, email_settings models.Email, idata interface{}) {
	data, ok := idata.(AlertData)
	if ok {
		go func(dbm *gorp.DbMap, data AlertData) {
			txn, err := dbm.Begin()
			if err != nil {
				PublishOneError(fmt.Errorf("Could not create db transaction to save alert data."))
			} else {
				err = txn.Insert(&data)
				if err != nil {
					fmt.Println("Alert insert err:", err)
				} else {
					err = txn.Commit()
					if err != nil {
						fmt.Println("Alert commit err:", err)
					}
				}
			}
		}(dbm, data)
		if email_settings.Addr != "" {
			if buffering_enabled {
				buffer_email <- data
			} else {
				SendOneEmail(data, email_settings)
			}
		}
	} else {
		go PublishOneError(errors.New("Problem converting alert data for storage."))
	}
}

func Run(dbm *gorp.DbMap) {
	var buffer_email = make(chan AlertData)
	subscribers := list.New()
	var email_settings models.Email
	var email_buffer []AlertData
	var email_ticker <-chan time.Time

	txn, err := dbm.Begin()
	if err != nil {
		go PublishOneError(fmt.Errorf("Problem getting email settings: %v", err))
	} else {
		email_settings, err = models.GetEmail(txn)
		if err != nil {
			go PublishOneError(fmt.Errorf("Problem getting email settings: %v", err))
		}
		maxmsgs, maxduration, err := models.ProcessEmailRateLimits(email_settings)
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
				SaveTrends(dbm, event.Data)
			case "errors":
				SaveErrors(dbm, event.Data)
			case "alert":
				DoAlert(dbm, buffer_email, email_ticker != nil, email_settings, event.Data)
			}
		case alert := <-buffer_email:
			email_buffer = append(email_buffer, alert)
		case <-email_ticker:
			if len(email_buffer) > 0 {
				email := email_buffer[0]
				email_buffer = email_buffer[1:]
				go SendOneEmail(email, email_settings)
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
