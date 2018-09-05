package main

import (
	"github.com/HouzuoGuo/tiedot/db"
	"encoding/json"
	"strconv"
	"io/ioutil"
	"os"
	"sync"
)


const (
	DB_PATH = "./blockdb/"
	NUM_PARTS = 2
)


type Blockdb struct {
	chain string
	channel string
	db *db.DB
	col *db.Col
}


var dblock *sync.Mutex = &sync.Mutex{}
var dbcache map[string]*Blockdb = make(map[string]*Blockdb)


//
// 该方法会从 db 池中返回已经缓存的对象.
//
func OpenBlockDB(chain string, channel string) (*Blockdb, error) {
	dblock.Lock()
	defer dblock.Unlock()

	if ret := dbcache[chain]; ret != nil {
		return ret, nil
	}

	if err := writeDBConfig(chain); err != nil {
		return nil, err
	}

	db, err := db.OpenDB(DB_PATH + chain)
	if (err != nil) {
		return nil, err
	}

	col := db.Use(channel)
	if col == nil {
		if err := db.Create(channel); err != nil {
			return nil, err
		}
		col = db.Use(channel)
		if err := col.Index([]string{"key", "previousKey"}); err != nil {
			return nil, err
		}
	}
	ret := &Blockdb{ chain, channel, db, col }
	dbcache[chain] = ret
	return ret, nil
}


//
// 预先写入并发配置, 防止过多的硬盘预分配
//
func writeDBConfig(chain string) (err error) {
	num := []byte(strconv.Itoa(NUM_PARTS))
	numFile := DB_PATH + chain +"/"+ db.PART_NUM_FILE

	if _, err := os.Stat(numFile); err == nil {
		return nil
	}
	if err := ioutil.WriteFile(numFile, num, 0600); err != nil {
		return err
	}
	return nil
}


func (b* Blockdb) Put(json_bin []byte) (int, error) {
	obj := make(map[string]interface{})
	err := json.Unmarshal(json_bin, &obj) 
	if err != nil { 
		return 0, err
	} 

	docID, err1 := b.col.Insert(obj)
	if err1 != nil {
		return 0, err1
	}
	return docID, nil
}


//
//TODO: 返回结构
//
func (b* Blockdb) Get(key string) (*string, error) {
	queryResult := make(map[int]struct{})
	query := map[string]interface{}{ "key": key	}
	if err := db.EvalQuery(query, b.col, &queryResult); err != nil {
		return nil, err
	}
	by, err := json.Marshal(queryResult)
	if (err != nil) {
		return nil, err
	}
	s := string(by)
	return &s, nil
}


func (b* Blockdb) Close() error {
	err := b.db.Close()
	b.db = nil
	b.col = nil
	return err
}


func CloseAllBlockDB() {
	for id := range dbcache {
		dbcache[id].Close()
		delete(dbcache, id)
	}
}