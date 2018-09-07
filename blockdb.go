package main

import (
	"github.com/HouzuoGuo/tiedot/db"
	"strconv"
	"io/ioutil"
	"os"
	"sync"
	logger "log"
)


const (
	DB_PATH   = "./block-db/"
	NUM_PARTS = 2
	FILE_MODE = 0600
	META_FILE = "meta.conf"
)


type Blockdb struct {
	chain 	string
	channel string
	db 			*db.DB
	col 		*db.Col
	meta    *os.File
	lastid  int
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
		// 如果两个属性写在一个数组中是组合主键
		if err := col.Index([]string{"key"}); err != nil {
			return nil, err
		}
	}

	metaFull := DB_PATH + chain +"/"+ channel +"-"+ META_FILE
	meta, err := os.OpenFile(metaFull, os.O_RDWR|os.O_CREATE, FILE_MODE)
	if err != nil {
		return nil, err
	}
	var buf [20]byte
	var lastid int = 0
	if n, _ := meta.ReadAt(buf[:], 0); n > 0 {
		lastid, err = strconv.Atoi(string(buf[:n]))
	}

	ret := &Blockdb{ chain, channel, db, col, meta, lastid }
	dbcache[chain] = ret
	return ret, nil
}


//
// 预先写入并发配置, 防止过多的硬盘预分配
//
func writeDBConfig(chain string) (err error) {
	num := []byte(strconv.Itoa(NUM_PARTS))
	dir := DB_PATH + chain +"/"
	numFile := dir + db.PART_NUM_FILE

	if _, err := os.Stat(numFile); err == nil {
		return nil
	}
	if err := os.MkdirAll(dir, FILE_MODE); err != nil {
		return err
	}
	if err := ioutil.WriteFile(numFile, num, FILE_MODE); err != nil {
		return err
	}
	return nil
}


func (b* Blockdb) Put(obj map[string]interface{}) (int, error) {
	id, err := b.Find(obj["key"].(string))
	if err != nil || id != 0 {
		return id, err
	}

	docID, err1 := b.col.Insert(obj)
	if err1 != nil {
		return 0, err1
	}
	b.lastid = docID
	if _, err := b.meta.WriteAt([]byte(strconv.Itoa(b.lastid)), 0); err != nil {
		log("Write last id fail", err)
	}
	return docID, nil
}


func (b* Blockdb) Get(key string) (map[string]interface{}, error) {
	id, err := b.Find(key)
	if err != nil || id == 0 {
		return nil, err
	}
	return b.col.Read(id)
}


func (b *Blockdb) GetLastKey() (*string, error) {
	if b.lastid != 0 {
		block, err := b.col.Read(b.lastid)
		if err != nil {
			return nil, err
		}
		if key, ok := block["key"].(string); ok {
			return &key, nil
		}
	}
	return nil, nil
} 


func (b* Blockdb) Find(key string) (int, error) {
	queryResult := make(map[int]struct{})
	query := map[string]interface{}{
		"eq":    key,
		"in":    []interface{}{"key"},
 	}
	if err := db.EvalQuery(query, b.col, &queryResult); err != nil {
		return 0, err
	}
	
	l := len(queryResult)
	if l == 0 {
		return 0, nil
	} else if l != 1 {
		// 退出系统
		logger.Fatalln("System fail, One key Two Block !")
	}
	for id := range queryResult {
		return id, nil
	}
	return 0, nil
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