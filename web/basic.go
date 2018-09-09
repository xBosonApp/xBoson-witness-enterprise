package witness_web

import (
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

	log.Println("Manger URL:  http://"+ witness.GetHttpHost() +
			DEFAULT_INDEX_FULL +"?pass=" + witness.GetPass())
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
		w.WriteHeader(200)
		w.Header().Set("Content-Type", getMimeType(fileName))
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

	w.WriteHeader(200)
	w.Header().Set("Content-Type", getMimeType(fileName))
	
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


func wjson(w io.Writer, m interface{}) {
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