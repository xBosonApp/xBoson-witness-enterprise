package witness_web

import (
  "io/ioutil"
  "github.com/HouzuoGuo/tiedot/db"
  "net/http"
  "witness-enterprise/witness"
  "time"
)


func login(w http.ResponseWriter, r *http.Request) {
  pass := r.URL.Query().Get("pass")

  if pass == witness.GetConfig().GetPass() {
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


func chainList(h Http) {
  all := make([]string, 0, 10)
  fileinfo , err := ioutil.ReadDir(witness.DB_PATH)

  if err != nil {
    wjson(h.W, &Msg{ 1, "失败"+ err.Error(), nil })
    return
  }

  for _, f := range fileinfo {
    name := f.Name()
    if name != "" {
      all = append(all, name)
    }
  }
  wjson(h.W, &Msg{ 0, "ok", all })
}


func channelList(h Http) {
  chain := h.R.URL.Query().Get("chain")
  db, err := db.OpenDB(witness.DB_PATH + chain)
  defer db.Close()

  if err != nil {
    wjson(h.W, &Msg{ 1, err.Error(), nil })
    return
  }
  all := db.AllCols()
  wjson(h.W, &Msg{ 0, "ok", all })
}


func block(h Http) {
  chain   := h.R.URL.Query().Get("chain")
  channel := h.R.URL.Query().Get("channel")
  key     := h.R.URL.Query().Get("key")

  db, err := witness.OpenBlockDB(chain, channel)
  if err != nil {
    wjson(h.W, &Msg{ 1, err.Error(), nil })
    return
  }  

  if key == "" {
    k, err := db.GetLastKey();
    if err != nil {
      h.Json(&Msg{ 2, err.Error(), nil })
      return
    }
    key = *k
  }

  const count = 10
  all := make([]interface{}, 0, count)
  for i:=0; i<count; i++ {
    b, err := db.Get(key)
    if err != nil {
      h.Json(&Msg{ 3, err.Error(), nil })
      return
    }
    if b == nil {
      break
    }

    all = append(all, b)
    key = b["previousKey"].(string)
  }
  h.Json(&Msg{ 0, "ok", all })
}


func read_log(h Http) {
  var loglist [][]interface{}
  end := time.Now().Add(10 * time.Second)
  rc, _ := h.S.GetInt("rc")

  //
  // 10秒内如果没有消息就返回 null, 否则立即返回记录, 或等待10秒
  //
  for end.After(time.Now())  {
    loglist, rc = witness.BlockMsg.Read(rc)
    if loglist != nil && len(loglist) > 0 {
      break;
    }
    time.Sleep(100 * time.Millisecond)
  }
  
  h.S.Set("rc", rc)
  h.Json(&Msg{ 0, "ok", loglist })
}


func servier_info(h Http) {
  c := witness.GetConfig()
  h.Json(&Msg{ 0, "ok", map[string]interface{}{
    "id"   			: c.ID,
    "host" 			: c.Host,
    "xboson" 		: c.URLxBoson,
    "publicKey" : witness.GetPublicKeyStr(),
  } })
}