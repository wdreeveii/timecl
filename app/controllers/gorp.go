package controllers

import (
	"code.google.com/p/go.crypto/bcrypt"
	"database/sql"
	"github.com/coopernurse/gorp"
	_ "github.com/mattn/go-sqlite3"
	r "github.com/revel/revel"
	//"timecl/app/logger"
	"timecl/app/models"
	//"timecl/app/network_manager"
)

var (
	dbm *gorp.DbMap
)

type Count struct {
	Count int64 `db:"count(*)"`
}

func setColumnSizes(t *gorp.TableMap, colSizes map[string]int) {
	for col, size := range colSizes {
		t.ColMap(col).MaxSize = size
	}
}

func Init(dbmap *gorp.DbMap) {
	dbm = dbmap

	t := dbm.AddTable(models.User{}).SetKeys(true, "UserId")
	t.ColMap("Password").Transient = true
	setColumnSizes(t, map[string]int{
		"Username": 20,
		"Name":     100,
	})

	t = dbm.AddTable(models.AppConfig{}).SetKeys(true, "ConfigId")
	t.ColMap("Key").Unique = true
	setColumnSizes(t, map[string]int{
		"Key": 100,
		"Val": 1000,
	})

	t = dbm.AddTable(models.Hotel{}).SetKeys(true, "HotelId")
	setColumnSizes(t, map[string]int{
		"Name":    50,
		"Address": 100,
		"City":    40,
		"State":   6,
		"Zip":     6,
		"Country": 40,
	})

	t = dbm.AddTable(models.Booking{}).SetKeys(true, "BookingId")
	t.ColMap("User").Transient = true
	t.ColMap("Hotel").Transient = true
	t.ColMap("CheckInDate").Transient = true
	t.ColMap("CheckOutDate").Transient = true
	setColumnSizes(t, map[string]int{
		"CardNumber": 16,
		"NameOnCard": 50,
	})
	//network_manager.InitNetworkConfigTables(dbm)
	//logger.InitLoggerTables(dbm)
	err := dbm.CreateTablesIfNotExists()
	if err != nil {
		panic(err)
	}

	results, err := dbm.Select(Count{}, `select count(*) from User`)
	if err != nil {
		panic(err)
	}
	if results[0].(*Count).Count == 0 {
		bcryptPassword, _ := bcrypt.GenerateFromPassword([]byte("demo"), bcrypt.DefaultCost)
		demoUser := &models.User{0, "Demo User", "demo", "demo", bcryptPassword}
		if err := dbm.Insert(demoUser); err != nil {
			panic(err)
		}
	}
}

type GorpController struct {
	*r.Controller
	Txn *gorp.Transaction
}

func (c *GorpController) Begin() r.Result {
	txn, err := dbm.Begin()
	if err != nil {
		panic(err)
	}
	c.Txn = txn
	return nil
}

func (c *GorpController) Commit() r.Result {
	if c.Txn == nil {
		return nil
	}
	if err := c.Txn.Commit(); err != nil && err != sql.ErrTxDone {
		panic(err)
	}
	c.Txn = nil
	return nil
}

func (c *GorpController) Rollback() r.Result {
	if c.Txn == nil {
		return nil
	}
	if err := c.Txn.Rollback(); err != nil && err != sql.ErrTxDone {
		panic(err)
	}
	c.Txn = nil
	return nil
}
