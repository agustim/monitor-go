package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/dgraph-io/badger"
	"github.com/gin-gonic/gin"
	"github.com/olebedev/config"
)

const CfgFile string = "monitor.config"

var DBPath string
var LogFile string
var CsvFile string
var CsvSinergiaFile string
var XApiKeyRead string
var XApiKeyWrite string
var TestEnv bool
var PortServer int
var DBPoint *badger.DB
var AssistitCounter int

func OpenDatabase() (*badger.DB, error) {
	opts := badger.DefaultOptions
	opts.Dir = DBPath
	opts.ValueDir = DBPath
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func CloseDatabase(db *badger.DB) {
	db.Close()
}

func main() {
	var err error
	PortServer = 9001
	LoadConfig()
	defer CloseDatabase(DBPoint)
	DBPoint, err = OpenDatabase()
	if err != nil {
		fmt.Println("Some problem with database: ", err)
		return
	}
	if TestEnv {
		// Develop Enviroment
		// Create Database Test
	} else {
		// Production Environment
		gin.SetMode(gin.ReleaseMode)
		// Active log
		gin.DisableConsoleColor()
		f, err := os.OpenFile(LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			fmt.Println("[monitor]: File error. ", err)
		}
		gin.DefaultWriter = io.MultiWriter(f, os.Stdout)
		fmt.Println("[monitor]: Start server.")
	}

	// Create Routes
	router := gin.Default()
	router.Use(ApiKeyMiddleware)
	v1 := router.Group("/")
	{
		v1.POST("/add", fetchAdd)
		v1.GET("/export", fetchExport)
		v1.GET("/export/:server", fetchExportServer)
	}
	router.Run(":" + strconv.Itoa(PortServer))
}

func LoadConfig() {
	// Load Config file
	cfg, err := config.ParseYamlFile(CfgFile)
	if err != nil {
		fmt.Println("Error al llegir el fitxer de configuració, potser no existeix.")
		os.Exit(1)
	}
	DBPath, err = cfg.String("server.dbpath")
	LogFile, err = cfg.String("server.logfile")
	XApiKeyWrite, err = cfg.String("server.xapikeywrite")
	TestEnv, err = cfg.Bool("server.testenv")
	PortServer, err = cfg.Int("server.port")
	if err != nil {
		fmt.Println("Reviseu els parametres, no son correctes.")
		os.Exit(2)
	}
}

func ApiKeyMiddleware(c *gin.Context) {
	// c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	// c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
	// c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-Api-Key")
	// c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")

	// if c.Request.Method == "OPTIONS" {
	// 	c.AbortWithStatus(204)
	// 	return
	// }
	token := c.Request.Header.Get("X-Api-Key")
	if token == XApiKeyWrite {
		c.Next()
	} else {
		c.AbortWithStatus(401)
	}
}

func fetchAdd(c *gin.Context) {
	var r RegistreServer

	c.BindJSON(&r)
	t := time.Now().Unix()
	r.Hora = strconv.FormatInt(t, 10)
	fmt.Println(r)
	err := r.Create(DBPoint)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": http.StatusNotFound, "message": "Error in Create!"})
		return
	}
	c.JSON(http.StatusOK, r.String())
}
func fetchExportServer(c *gin.Context) {
	serverId := c.Param("server")

	b := &bytes.Buffer{}
	w := csv.NewWriter(b)

	ExportServer(serverId, w, DBPoint)
	w.Flush()

	c.Data(http.StatusOK, "text/csv", b.Bytes())
}

func fetchExport(c *gin.Context) {

	b := &bytes.Buffer{}
	w := csv.NewWriter(b)

	Export(w, DBPoint)
	w.Flush()

	c.Data(http.StatusOK, "text/csv", b.Bytes())
}

func Export(w *csv.Writer, db *badger.DB) {
	r := &RegistreServer{}

	err := db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			v, _ := item.Value()
			json.Unmarshal(v, r)
			w.Write(r.Strings())
		}
		return nil
	})

	if err != nil {
		fmt.Println("Error Iterator: ", err)
	}

}

func ExportServer(serverId string, w *csv.Writer, db *badger.DB) {
	var h HistoryServer
	r := &RegistreServer{}

	h.Get(db, r.IdServer)

	err := db.View(func(txn *badger.Txn) error {
		var errtorn error
		for _, hora := range h.Hores {
			item, err := txn.Get([]byte(serverId + "-" + hora))
			if err != nil {
				errtorn = err
				fmt.Println("error")
			}
			valor, err := item.Value()
			if err == nil {
				json.Unmarshal([]byte(valor), r)
				w.Write(r.Strings())
			}
		}
		return errtorn
	})
	if err != nil {
		fmt.Println("err: " + err.Error())
	}
}
