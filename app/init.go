package app

import (
	"database/sql"
	//"fmt"
	"github.com/coopernurse/gorp"
	_ "github.com/go-sql-driver/mysql"
	"github.com/revel/revel"
	//"os"
	//"os/signal"
	//"syscall"
	"github.com/wdreeveii/timecl/app/controllers"
	"github.com/wdreeveii/timecl/app/logger"
	"github.com/wdreeveii/timecl/app/models"
	"github.com/wdreeveii/timecl/app/network_manager"
)

func Init() {
	Driver, found := revel.Config.String("db.driver")
	if !found {
		revel.ERROR.Fatal("No db.driver found.")
	}
	Spec, found := revel.Config.String("db.spec")
	if !found {
		revel.ERROR.Fatal("No db.spec found.")
	}

	// Open a connection.
	var err error
	var Db *sql.DB

	Db, err = sql.Open(Driver, Spec)
	if err != nil {
		revel.ERROR.Fatal(err)
	}
	dbm := &gorp.DbMap{Db: Db, Dialect: gorp.MySQLDialect{"InnoDB", "UTF8"}}
	dbm.TraceOn("[gorp]", revel.INFO)
	log := logger.Init(dbm, models.EmailSettingsProvider{})
	log.ErrorOn("[logger]", revel.ERROR)
	network_manager.Init(dbm)
	controllers.Init(dbm)
	go log.Run()
	controllers.InitEngines(dbm)
}
func init() {
	/*go func() {
		signal_source := make(chan os.Signal)
		signal.Notify(signal_source, syscall.SIGHUP)
		for {
			<-signal_source
			fmt.Println("Terminal Disconnected")
		}
	}()*/
	// Filters is the default set of global filters.
	revel.Filters = []revel.Filter{
		revel.PanicFilter,             // Recover from panics and display an error page instead.
		revel.RouterFilter,            // Use the routing table to select the right Action
		revel.FilterConfiguringFilter, // A hook for adding or removing per-Action filters.
		revel.ParamsFilter,            // Parse parameters into Controller.Params.
		revel.SessionFilter,           // Restore and write the session cookie.
		revel.FlashFilter,             // Restore and write the flash cookie.
		revel.ValidationFilter,        // Restore kept validation errors and save new ones from cookie.
		revel.I18nFilter,              // Resolve the requested language
		revel.InterceptorFilter,       // Run interceptors around the action.
		revel.ActionInvoker,           // Invoke the action.
	}
	revel.OnAppStart(Init)
}
