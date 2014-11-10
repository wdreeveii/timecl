package controllers

import (
	"code.google.com/p/go.net/websocket"
	"fmt"
	"github.com/revel/revel"
	"strings"
	"time"
	"github.com/wdreeveii/timecl/app/logger"
)

type Logger struct {
	Application
}

func (c Logger) checkUser() revel.Result {
	if user := c.connected(); user == nil {
		c.Flash.Error("Please log in first")
		return c.Redirect(Application.Index)
	}
	return nil
}

func (c Logger) Index() revel.Result {
	return c.Render()
}

func (c Logger) Alerts() revel.Result {
	var alertlist []logger.AlertData
	_, err := c.Txn.Select(&alertlist, "SELECT * FROM AlertData ORDER BY Timestamp DESC")
	if err != nil {
		return c.RenderError(err)
	}
	for i, _ := range alertlist {
		alertlist[i].Time = time.Unix(alertlist[i].Timestamp, 0)
	}
	return c.Render(alertlist)
}

func (c Logger) Errors() revel.Result {
	var errlist []logger.ErrorData
	_, err := c.Txn.Select(&errlist, "SELECT Error, Count, Timestamp, FirstTimestamp FROM ErrorData ORDER BY Timestamp DESC")
	if err != nil {
		return c.RenderError(err)
	}
	for i, _ := range errlist {
		errlist[i].Time = time.Unix(errlist[i].Timestamp, 0)
		errlist[i].First = time.Unix(errlist[i].FirstTimestamp, 0)
	}
	return c.Render(errlist)
}

type LogDataClientFormat struct {
	Timestamp int64
	Min       float64
	Max       float64
	Avg       float64
}

func (c Logger) GetData(start string, end string) map[string][]LogDataClientFormat {
	var queryArgs []interface{}
	var queryPieces []string
	if start != "start" {
		queryPieces = append(queryPieces, "Timestamp >= strftime('%s', ?, 'utc')")
		queryArgs = append(queryArgs, start)
	}
	if end != "now" {
		queryPieces = append(queryPieces, "Timestamp <= strftime('%s', ?, 'utc')")
		queryArgs = append(queryArgs, end)
	}
	var whereclause = strings.Join(queryPieces, " AND ")
	var query = "SELECT * FROM LoggingData"
	if whereclause != "" {
		query += " WHERE " + whereclause
	}
	query += " ORDER BY ObjectId, Timestamp ASC"

	c.Begin()

	var objects = make(map[string][]LogDataClientFormat)

	results, err := c.Txn.Select(logger.LoggingData{}, query, queryArgs...)
	if err != nil {
		fmt.Println("Error getting datalogs:", err)
	}
	c.Commit()
	var groupsize = len(results) / 5000
	var group = make(map[int][]*logger.LoggingData)
	for _, v := range results {
		m := v.(*logger.LoggingData)
		key := fmt.Sprintf("%v", m.ObjectId)
		if groupsize > 0 {
			if len(group[m.ObjectId]) < groupsize-1 {
				group[m.ObjectId] = append(group[m.ObjectId], m)
			} else {
				var time = m.Timestamp
				var min = m.Min
				var max = m.Max
				var avg = m.Avg
				for _, d := range group[m.ObjectId] {
					if d.Timestamp < time {
						time = d.Timestamp
					}
					if d.Min < min {
						min = d.Min
					}
					if d.Max > max {
						max = d.Max
					}
					avg += d.Avg
				}
				avg = avg / float64(groupsize)
				objects[key] = append(objects[key],
					LogDataClientFormat{Timestamp: time,
						Min: min,
						Max: max,
						Avg: avg})
				group[m.ObjectId] = nil
			}

		} else {
			objects[key] = append(objects[key],
				LogDataClientFormat{Timestamp: m.Timestamp,
					Min: m.Min,
					Max: m.Max,
					Avg: m.Avg})
		}
	}
	return objects
}

/*func (c Engine) NewEngine() revel.Result {
	InitEngine()
	engine.Save()
	return c.RenderJson(1)
}*/

func (c Logger) LoggerSocket(ws *websocket.Conn) revel.Result {
	c.Commit()
	subscription := logger.Subscribe()
	defer subscription.Cancel()

	newMessages := make(chan logger.Event)
	go func() {
		var msg logger.Event
		for {
			err := websocket.JSON.Receive(ws, &msg)
			if err != nil {
				revel.INFO.Println("Error receiving msg from client:", err)
				close(newMessages)
				return
			}
			newMessages <- msg
		}
	}()

	for {
		select {
		case event := <-subscription.New:
			if err := websocket.JSON.Send(ws, &event); err != nil {
				revel.INFO.Println("Error sending msg to client:", err)
				return nil
			}
		case msg, ok := <-newMessages:
			if !ok {
				return nil
			}
			switch msg.Type {
			case "get_new_range":
				var start = "start"
				var end = "end"
				var data, ok = msg.Data.(map[string]interface{})
				if ok {
					start, _ = data["start"].(string)
					end, _ = data["end"].(string)
				}

				update_data := c.GetData(start, end)
				var update_data_event = logger.Event{Type: "update_data",
					Data: update_data}
				if err := websocket.JSON.Send(ws, &update_data_event); err != nil {
					revel.INFO.Println(err)
					return nil
				}
			}
		}
	}
	return nil
}
