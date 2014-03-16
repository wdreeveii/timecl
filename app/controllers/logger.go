package controllers

import (
	//"code.google.com/p/go.net/websocket"
	"fmt"
	"github.com/revel/revel"
	"timecl/app/logger"
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

type LogDataClientFormat struct {
	Timestamp int64
	Min       float64
	Max       float64
	Avg       float64
}

func (c Logger) GetData() revel.Result {
	ids, err := c.Txn.Select(struct{ ObjectId int }{}, `SELECT DISTINCT ObjectId FROM LoggingData`)
	if err != nil {
		fmt.Println("Error getting keys:", err)
		return c.RenderError(err)
	}
	var objects = make(map[string][]LogDataClientFormat)
	if len(ids) > 0 {
		var query string
		for k, v := range ids {
			obj := v.(*struct{ ObjectId int })
			if k != 0 {
				query += "UNION ALL\n"
			}
			query += "SELECT * FROM (SELECT * FROM LoggingData WHERE ObjectId = " + fmt.Sprintf("%d", obj.ObjectId) + " ORDER BY Timestamp DESC) a" + fmt.Sprintf("%d\n", k)
		}
		query += "ORDER BY ObjectId, Timestamp ASC"
		results, err := c.Txn.Select(logger.LoggingData{}, query)
		if err != nil {
			fmt.Println("Error getting datalogs:", err)
		}

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
	}
	return c.RenderJson(objects)
}

/*func (c Engine) NewEngine() revel.Result {
	InitEngine()
	engine.Save()
	return c.RenderJson(1)
}*/

/*func (c Engine) EngineSocket(ws *websocket.Conn) revel.Result {
	subscription := engine.Subscribe()
	defer subscription.Cancel()
	init := engine.ListObjects()
	if err := websocket.JSON.Send(ws, &init); err != nil {
		revel.INFO.Println(err)
		return nil
	}
	ports := engine.ListPorts()
	if err := websocket.JSON.Send(ws, &ports); err != nil {
		revel.INFO.Println(err)
		return nil
	}
	revel.INFO.Println(ports)

	newMessages := make(chan logic_engine.Event)
	go func() {
		var msg logic_engine.Event
		for {
			err := websocket.JSON.Receive(ws, &msg)
			if err != nil {
				revel.INFO.Println("Error receiving msg from client:", err)
				close(newMessages)
				return
			}
			fmt.Println("msg from client")
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
			engine.Publish(msg)
		}
	}
	return nil
}*/
