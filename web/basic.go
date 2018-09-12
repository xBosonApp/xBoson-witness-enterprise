package witness_web

import (
	"io/ioutil"
	"net/http"
	"os"
	"io"
	"mime"
	"path/filepath"
	"log"
	"time"
	"encoding/json"
	"github.com/kataras/go-sessions"
	"github.com/gorilla/securecookie"
	"github.com/HouzuoGuo/tiedot/db"
	"../witness"
)

const (
	base_url 			= "/"+ witness.URL_prefix +"web/"
	base_service 	= "/"+ witness.URL_prefix + "service/"
	www_path 			= "./www/"
	default_index = "index.html"
	data_page 		= "data.html"
	DEFAULT_INDEX_FULL = base_url + default_index
)

type Page struct {
}

type Msg struct {
	Code int 					`json:"code"`
	Msg  string 			`json:"msg"`
	Data interface{} 	`json:"data"`
}

type Response http.ResponseWriter
type Request  http.Request
type Session  sessions.Session

var file_mapping = make(map[string][]byte)
var sess *sessions.Sessions;


func init() {
	hashKey  := []byte("winess-2104nxk.zjfpeoq9203gh")
	blockKey := []byte("winess-f32w3nvb,x.vo;so9wghlsd;[q")
	secureCookie := securecookie.New(hashKey, blockKey)
	sess = sessions.New(sessions.Config{
		Cookie: "witnesssessionid",
		Expires: time.Hour * 2,
		DisableSubdomainPersistence: false,
		Encode: secureCookie.Encode,
		Decode: secureCookie.Decode,
	})
}


func StartWebService() {
	http.Handle("/", http.RedirectHandler(DEFAULT_INDEX_FULL, http.StatusMovedPermanently))
	http.Handle(base_url, &Page{})
	http.HandleFunc(base_service +"login", login)
	http.HandleFunc(base_service +"logout", logout)

	HandleFunc("chain_list", 		chainList)
	HandleFunc("channel_list", 	channelList)
	HandleFunc("get_block", 		block)
	HandleFunc("read_log", 			read_log)

	log.Println("Manger URL:  http://"+ witness.GetHttpHost() +
			DEFAULT_INDEX_FULL +"?pass=" + witness.GetPass())
}


//
// 带有登陆检查
//
func HandleFunc(path string, h func(Response, *Request, *Session)) {
	http.HandleFunc(base_service + path, func(w http.ResponseWriter, r *http.Request) {
		log.Println("Service", r.URL)
		//TODO: 登陆检查
		s := sess.Start(w, r)
		// if succ, err := s.GetBoolean("login"); !succ || err != nil {
		// 	wjson(w, &Msg{ 1, "未登录", err })
		// 	return
		// }
		h(Response(w), (*Request)(r), (*Session)(s))
	})
}


func (p *Page) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fileName := r.URL.Path[len(base_url):]
	if fileName == "" {
		w.Header().Set("Location", DEFAULT_INDEX_FULL)
		w.WriteHeader(http.StatusMovedPermanently)
		return
	}

	content := file_mapping[fileName]
	if content != nil {
		w.Header().Set("Content-Type", getMimeType(fileName))
		w.WriteHeader(200)
		w.Write(content)
		return;
	}

	filePath := www_path + fileName
	file, err := os.OpenFile(filePath, os.O_RDONLY, 0600)
	defer file.Close()

	if err != nil {
		w.WriteHeader(404)
		w.Write([]byte(err.Error()))
		return
	}

	w.Header().Set("Content-Type", getMimeType(fileName))
	w.WriteHeader(200)
	
	if _, err = io.Copy(w, file); err != nil {
		log.Println("Response fail", err)
	}
}


func getMimeType(fileName string) string {
	ctype := mime.TypeByExtension(filepath.Ext(fileName))
	if ctype == "" {
		ctype = "application/octet-stream"
	}
	return ctype
}


func wjson(w http.ResponseWriter, m interface{}) {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.Encode(m)
}


func login(w http.ResponseWriter, r *http.Request) {
	pass := r.URL.Query().Get("pass")

	if pass == witness.GetPass() {
		s := sess.Start(w, r)
		s.Set("login", true)
		wjson(w, &Msg{ 0, "ok", nil })
	} else {
		wjson(w, &Msg{ 1, "失败", nil })
	}
}


func logout(w http.ResponseWriter, r *http.Request) {
	s := sess.Start(w, r)
	s.Delete("login")
	wjson(w, &Msg{ 0, "ok", nil })
}


func chainList(w Response, r *Request, s *Session) {
	all := make([]string, 0, 10)
	fileinfo , err := ioutil.ReadDir(witness.DB_PATH)

	if err != nil {
		wjson(w, &Msg{ 1, "失败"+ err.Error(), nil })
		return
	}

	for _, f := range fileinfo {
		name := f.Name()
		if name != "" {
			all = append(all, name)
		}
	}
	wjson(w, &Msg{ 0, "ok", all })
}


func channelList(w Response, r *Request, s *Session) {
	chain := r.URL.Query().Get("chain")
	db, err := db.OpenDB(witness.DB_PATH + chain)
	defer db.Close()

	if err != nil {
		wjson(w, &Msg{ 1, err.Error(), nil })
		return
	}
	all := db.AllCols()
	wjson(w, &Msg{ 0, "ok", all })
}


func block(w Response, r *Request, s *Session) {
	chain   := r.URL.Query().Get("chain")
	channel := r.URL.Query().Get("channel")
	key     := r.URL.Query().Get("key")

	db, err := witness.OpenBlockDB(chain, channel)
	if err != nil {
		wjson(w, &Msg{ 1, err.Error(), nil })
		return
	}

	if key == "" {
		k, err := db.GetLastKey();
		if err != nil {
			wjson(w, &Msg{ 2, err.Error(), nil })
			return
		}
		key = *k
	}

	const count = 10
	all := make([]interface{}, 0, count)
	for i:=0; i<count; i++ {
		b, err := db.Get(key)
		if err != nil {
			wjson(w, &Msg{ 3, err.Error(), nil })
			return
		}
		if b == nil {
			break
		}

		all = append(all, b)
		key = b["previousKey"].(string)
	}
	wjson(w, &Msg{ 0, "ok", all })
}


func read_log(w Response, r *Request, s *Session) {
	var loglist [][]interface{}
	end := time.Now().Add(10 * time.Second)
	//
	// 10秒内如果没有消息就返回 null, 否则立即返回记录, 或等待10秒
	//
	for end.After(time.Now())  {
		loglist = witness.BlockMsg.Read()
		if loglist != nil && len(loglist) > 0 {
			break;
		}
		time.Sleep(100 * time.Millisecond)
	}
	
	wjson(w, &Msg{ 0, "ok", loglist })
}