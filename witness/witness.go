package witness
 
import (
	"net/http"
	"strconv"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/url"
	"fmt"
	"io"
	"io/ioutil"
	"flag"
	logger "log"
)

type Ret struct {
	Code int     `json:"code"` 
	Msg  string  `json:"msg"` 
	Id   string  `json:"id"` 
	Data string  `json:"data"`
}

type BlockQuery struct {
	key     string
	chain   string
	channel string
	c       int			// 这个变量防止启动多个全链检查
}

const URL_prefix  = "witness/"
const sign_url    = "/"+ URL_prefix +"sign"
const deliver_url = "/"+ URL_prefix +"deliver"
const def_config  = "witness-config.json"

var pubkeystr string
var prikey *ecdsa.PrivateKey
var configFile = flag.String("c", def_config, "Witness default Config File");
var xboson_url_base string
var c *Config
var requestBlockChan chan BlockQuery = make(chan BlockQuery, 1024)
var block_count = 0


func StartWitnessProgram() {
	defer CloseAllBlockDB()
	ls := Logset{}
	setLoggerFile(&ls)
	c = loadConfig(*configFile, setGlbKey)

	if c.Host != "" {
		if !findIpWithConfig() {
			findIpWithStdin()
		} else {
			log("Local IP not change")
		}
	} else {
		findIpWithStdin()
	}

	if c.Host == "" {
		logger.Fatalln("Cannot Find Local IP")
	} else {
		log("Local IP:", c.Host)
	}

	if c.URLxBoson == "" {
		logger.Fatalln("Config.URLxBoson cannot nil")
	}
	xboson_url_base = "http://"+ c.URLxBoson + "/xboson/witness/"

	if c.ID == "" {
		go doReg()
	} else {
		go doChange()
	}

	go readRequestThread()

	log("Witness peer start, Http Server start")
	log("Http Port", c.Port)
	log("Http Path", sign_url)
	http.HandleFunc(sign_url, sign)
	http.HandleFunc(deliver_url, deliver)
}


func StartHttpServer() {
	http.ListenAndServe(":"+ strconv.Itoa(c.Port), nil)
}


func doReg() {
	p := url.Values{}
	p.Set("algorithm",  "SHA256withECDSA")
	p.Set("publickey",  pubkeystr)
	p.Set("host", 			c.Host)
	p.Set("port", 			strconv.Itoa(c.Port))
	p.Set("urlperfix",  URL_prefix)

	log("Do register to xBoson platform")
	ret, err := callHttp("register", &p)
	if err != nil {
		return
	}
	if ret.Code != 0 {
		// 失败就退出
		logger.Fatalln("Register fail", ret.Code, ret.Msg, ret.Data)
	}

	log("Register to xBoson Success ID=", ret.Id)
	c.ID = ret.Id
	saveConfig(c, *configFile)
}


func doChange() {
	p := url.Values{}
	p.Set("host", 			c.Host)
	p.Set("port", 			strconv.Itoa(c.Port))
	p.Set("urlperfix",  URL_prefix)
	p.Set("id",					c.ID)

	log("Do change to xBoson")
	ret, err := callHttp("change", &p)
	if err != nil {
		return
	}
	log("Change ID =", c.ID, ", code =", ret.Code, ", message =", ret.Msg, ret.Data)
	if ret.Code != 0 {
		// 失败就退出
		logger.Fatalln("Change fail", ret.Msg)
	}
}


func doRequestBlock(chain, channel string, begin, end *string) {
	p := url.Values{}
	p.Set("id", 			c.ID)
	p.Set("chain", 		chain)
	p.Set("channel", 	channel)
	p.Set("begin", 		*begin)
	p.Set("end", 			*end)

	log("Do request block", *begin, "-", *end)
	ret, err := callHttp("reqb", &p)
	if err != nil {
		return
	}
	if ret.Code != 0 {
		log("Request fail", ret.Msg)
	}
}


func callHttp(api string, parm *url.Values) (*Ret, error) {
	res, err := http.Get(xboson_url_base + api +"?"+ parm.Encode())
	if err != nil {
		log("Http fail", err)
		return nil, err
	}
	defer res.Body.Close()

	var ret Ret
	dec := json.NewDecoder(res.Body)
	err = dec.Decode(&ret)
	if (err != nil) {
		log("Parse Json fail", err)
		return nil, err
	}
	return &ret, nil
}


func setGlbKey(pri *ecdsa.PrivateKey) {
	prikey = pri
	pubkey := &pri.PublicKey
	pubkeystr = getPublicKeyStr(pubkey)
	log("Public key:", pubkeystr)
}


func sign(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	
	h := sha256.New()
	_, err := io.Copy(h, r.Body)
	if err != nil {
		w.WriteHeader(500)
		fmt.Fprint(w, err)
		return
	}

	hash := h.Sum(nil)
	signature, err := prikey.Sign(rand.Reader, hash, nil)
	if (err != nil) {
		w.WriteHeader(500)
		fmt.Fprint(w, err)
		return
	}

	w.Write(signature)
	log("Sign", base64.StdEncoding.EncodeToString(signature))
}


func deliver(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	chain := r.Header.Get("chain")
	channel := r.Header.Get("channel")
	
	db, err := OpenBlockDB(chain, channel)
	if (err != nil) {
		log("Open DB fail", err)
		w.WriteHeader(500)
		return
	}

	json_bin, _ := ioutil.ReadAll(r.Body)
	obj := make(map[string]interface{})
	if err := json.Unmarshal(json_bin, &obj); err != nil {
		log("Deliver fail", err)
		w.WriteHeader(500)
		return
	}

	if err := verify(obj); err != nil {
		log("Deliver verify fail", err, string(json_bin))
		w.WriteHeader(403)
		return
	}

	id, err1 := db.Put(obj)
	if err1 != nil {
		log("DB insert fail", err1)
		w.WriteHeader(500)
		return
	}

	w.WriteHeader(200)
	log("Deliver", chain, channel, id)
	checkRelatedPutRequest(obj, db)
}


func readRequestThread() {
	sync_all := false
	show_c := 2
	for {
		b := <- requestBlockChan

		db, err := OpenBlockDB(b.chain, b.channel)
		if err != nil {
			log("Open DB with Push request fail", err)
		}

		find, err := db.Get(b.key)
		if err != nil {
			log("Push request fail", err)
			continue
		}

		//
		// BUG: 如果链中出现一段A空洞接着有一个block接着有空洞B, 空洞B无法填充
		//
		if find != nil { 
			if !sync_all && b.c == 0 {
				if pkey, ok := find["previousKey"].(string); ok && len(pkey) > 0 {
					block_count++
					if block_count % show_c == 0 {
						log("synchroized", pkey, block_count, "...")
						show_c = show_c << 1
					}
					requestBlockChan <- BlockQuery{ pkey, b.chain, b.channel, 0 }
				} else {
					sync_all = true
					log("Synchroized", block_count, "Blocks")
				}
			}
			continue 
		}

		var last *string
		if db.lastid != 0 {
			last = &b.key
		}
		doRequestBlock(b.chain, b.channel, &b.key, last)
	}
}


func checkRelatedPutRequest(b map[string]interface{}, db *Blockdb) {
	pkey, ok := b["previousKey"].(string)
	if !ok || len(pkey) <= 0 {
		return
	}
	requestBlockChan <- BlockQuery{ pkey, db.chain, db.channel, block_count }
}


func GetConfig() (*Config) {
	return c
}


func GetPublicKeyStr() string {
	return pubkeystr;
}