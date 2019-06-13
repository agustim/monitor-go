package main

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/dgraph-io/badger"
)

const NumHistoric int = 10

type HistoryServer struct {
	Hores []string `json:"Hores"`
}

func (h *HistoryServer) String() string {
	out, _ := json.Marshal(h)
	return (string(out))
}

func (h *HistoryServer) Sort() {
	sort.Strings(h.Hores)
}
func (h *HistoryServer) Get(db *badger.DB, idServer string) {
	err := db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(idServer))
		if err != nil {
			return err
		}
		valor, err := item.Value()
		if err == nil {
			json.Unmarshal([]byte(valor), h)
		}
		return err
	})
	if err != nil {
		fmt.Println("Error: ", err.Error())
	} else {
		fmt.Println(h.String())
	}
}
func (h *HistoryServer) Add(db *badger.DB, idServer string, hora string) {

	err := db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(idServer))
		if err != nil {
			return err
		}
		valor, err := item.Value()
		if err == nil {
			json.Unmarshal([]byte(valor), h)
		}
		return err
	})

	if err != nil && err.Error() == "Key not found" {
		h.Init()
	}
	h.Hores[0] = hora
	h.Sort()
	err = db.Update(func(txn *badger.Txn) error {
		err := txn.Set([]byte(idServer), []byte(h.String()))
		return err
	})
	if err != nil {
		fmt.Println("Error Update!")
	}

}
func (h *HistoryServer) Init() {
	h.Hores = make([]string, NumHistoric)
}
