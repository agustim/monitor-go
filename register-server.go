package main

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/dgraph-io/badger"
)

type RegistreServer struct {
	IdServer      string `json:"idServer"`
	Hora          string `json:"Hora"`
	CPU           int    `json:"CPU"`
	NSockets      int    `json:"NSockets"`
	Memory        int    `json:"Memroy"`
	TotalBytesIn  int    `json:"TotalBytesIn"`
	TotalBytesOut int    `json:"TotalBytesOut"`
}

func (r *RegistreServer) String() string {
	out, _ := json.Marshal(r)
	return (string(out))
}

func (r *RegistreServer) Strings() []string {
	record := make([]string, 7)
	record[0] = r.IdServer
	record[1] = r.Hora
	record[2] = strconv.Itoa(r.CPU)
	record[3] = strconv.Itoa(r.NSockets)
	record[4] = strconv.Itoa(r.Memory)
	record[5] = strconv.Itoa(r.TotalBytesIn)
	record[6] = strconv.Itoa(r.TotalBytesOut)

	return record
}

func (r *RegistreServer) Create(db *badger.DB) error {
	err := db.Update(func(txn *badger.Txn) error {
		err := txn.Set([]byte(r.IdServer+"-"+r.Hora), []byte(r.String()))
		return err
	})
	if err == nil {
		var h HistoryServer
		fmt.Println(r)
		h.Add(db, r.IdServer, r.Hora)
		h.Get(db, r.IdServer)
	}
	return err
}
