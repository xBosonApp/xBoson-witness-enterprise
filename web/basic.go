package witness_web

import (
  "net/http"
  "mime"
  "path/filepath"
  "log"
  "time"
  "encoding/json"
  "github.com/kataras/go-sessions"
  "github.com/gorilla/securecookie"
  "witness-enterprise/witness"
)

const (
  base_url      = "/"+ witness.URL_prefix +"web/"
  base_service  = "/"+ witness.URL_prefix + "service/"
  www_path      = "./www/"
  default_index = "index.html"
  data_page     = "data.html"
  DEFAULT_INDEX_FULL = base_url + default_index
)

type Msg struct {
  Code int          `json:"code"`
  Msg  string       `json:"msg"`
  Data interface{}  `json:"data"`
}

type Http struct {
  R  *http.Request
  W  http.ResponseWriter
  S  *sessions.Session
}

var sess *sessions.Sessions;

func init() {
  secureCookie := securecookie.New(
    securecookie.GenerateRandomKey(32), 
    securecookie.GenerateRandomKey(16))

  sess = sessions.New(sessions.Config{
    Cookie: "witnesssessionid",
    Expires: time.Hour * 2,
    Encode: secureCookie.Encode,
    Decode: secureCookie.Decode,
  })
}


func StartWebService() {
  http.Handle("/", http.RedirectHandler(DEFAULT_INDEX_FULL, http.StatusMovedPermanently))
  http.Handle(base_url, &Page{})
  http.HandleFunc(base_service +"login", login)
  http.HandleFunc(base_service +"logout", logout)

  HandleFunc("chain_list",    chainList)
  HandleFunc("channel_list",  channelList)
  HandleFunc("get_block",     block)
  HandleFunc("read_log",      read_log)
  HandleFunc("info",          servier_info)

  log.Println("Manger URL:  http://"+ witness.GetConfig().GetHttpHost() +
      DEFAULT_INDEX_FULL +"?pass=" + witness.GetConfig().GetPass())
}


//
// 带有登陆检查
//
func HandleFunc(path string, h func(h Http)) {
  http.HandleFunc(base_service + path, func(w http.ResponseWriter, r *http.Request) {
    log.Println("Service", r.URL)
    //TODO: 登陆检查
    s := sess.Start(w, r)
    if succ, err := s.GetBoolean("login"); !succ || err != nil {
      wjson(w, &Msg{ 401, "未登录", err })
      return
    }
    h(Http{ r, w, s })
  })
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


func (h *Http) Json(m interface{}) {
  wjson(h.W, m)
}
