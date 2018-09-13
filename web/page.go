package witness_web

import (
  "net/http"
  "os"
  "io"
  "log"
)

type Page struct {
}

var file_mapping = make(map[string][]byte)


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