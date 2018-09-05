package main
 
import (
	"net/http"
	"strconv"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"flag"
	"os"
	"encoding/json"
	"net/url"
	"net"
	"fmt"
	"io"
	"io/ioutil"

	x 			"crypto/x509"
	logger 	"log"
)

type Config struct {
	PrivateKey 	string
	Port 				int
	Host        string
	URLxBoson 	string
	ID          string
}

type Ret struct {
	Code int     `json:"code"` 
	Msg  string  `json:"msg"` 
	Id   string  `json:"id"` 
	Data string  `json:"data"`
}

const url_perfix  = "witness/"
const sign_url    = "/"+ url_perfix +"sign"
const deliver_url = "/"+ url_perfix +"deliver"
const def_config  = "witness-config.json"

var pubkeystr string
var prikey *ecdsa.PrivateKey
var configFile = flag.String("c", def_config, "Witness default Config File");
var xboson_url_base string
var c *Config


func main() {
	defer CloseAllBlockDB()
	ls := Logset{}
	setLoggerFile(&ls)
	c = new(Config)
	
	loadConfig()

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

	log("Witness peer start, Http Server start")
	log("Http Port", c.Port)
	log("Http Path", sign_url)
	http.HandleFunc(sign_url, sign)
	http.HandleFunc(deliver_url, deliver)
	http.ListenAndServe(":"+ strconv.Itoa(c.Port), nil)
}


func doReg() {
	p := url.Values{}
	p.Set("algorithm",  "SHA256withECDSA")
	p.Set("publickey",  pubkeystr)
	p.Set("host", 			c.Host)
	p.Set("port", 			strconv.Itoa(c.Port))
	p.Set("urlperfix",  url_perfix)

	log("Do register to xBoson platform")
	ret := callHttp("register", &p)
	if ret.Code != 0 {
		logger.Fatalln("Register fail", ret.Code, ret.Msg, ret.Data)
	}

	log("Register to xBoson Success ID=", ret.Id)
	c.ID = ret.Id
	saveConfig()
}


func doChange() {
	p := url.Values{}
	p.Set("host", 			c.Host)
	p.Set("port", 			strconv.Itoa(c.Port))
	p.Set("urlperfix",  url_perfix)
	p.Set("id",					c.ID)

	log("Do change to xBoson")
	ret := callHttp("change", &p)
	log("Change ID =", c.ID, ", code =", ret.Code, ", message =", ret.Msg, ret.Data)
	if ret.Code != 0 {
		logger.Fatalln("Change fail")
	}
}


func callHttp(api string, parm *url.Values) *Ret {
	res, err := http.Get(xboson_url_base + api +"?"+ parm.Encode())
	if err != nil {
		logger.Fatalln("Http fail", err)
	}
	defer res.Body.Close()

	var ret Ret
	dec := json.NewDecoder(res.Body)
	err = dec.Decode(&ret)
	if (err != nil) {
		logger.Fatalln("Parse Json fail", err)
	}
	return &ret
}


func loadConfig() {
	flag.Parse()
	log("Load config From:", *configFile)
	file, err := os.Open(*configFile)
	if err != nil {
		log(err)
		setConfigFromUser()
		genKey()
		saveConfig()
		return
	}
	defer file.Close()

	dec := json.NewDecoder(file)
	err = dec.Decode(c)
	if err != nil {
		logger.Fatalln("Parse Config fail", err)
	}
	loadKey()
}


func saveConfig() {
	file, err := os.OpenFile(*configFile, os.O_RDWR|os.O_CREATE, 0700)
	if err != nil {
		log("Open Config File fail", err)
		return;
	}
	defer file.Close()

	enc := json.NewEncoder(file)
	enc.SetIndent("", "  ")
	enc.Encode(c)
	log("Save config to", *configFile)
}


func setConfigFromUser() {
	fmt.Print("Input Http Port: ");
	fmt.Scanf("%d\n", &c.Port)
	fmt.Print("Input xBoson platform Adderss, [host:port]: ");
	fmt.Scanf("%s\n", &c.URLxBoson)

	if c.Port <= 0 { c.Port = 10080 }
	if c.URLxBoson == "" { c.URLxBoson = "localhost:8080" }
	log("xBoson Address:", c.URLxBoson);
}


func loadKey() {
	bin, err := base64.StdEncoding.DecodeString(c.PrivateKey)
	if err != nil {
		logger.Fatalln("Decode Private Key fail")
	}
	pri, err := x.ParseECPrivateKey(bin)
	if err != nil {
		logger.Fatalln("Parse Private key fail")
	}
	setGlbKey(pri)
	log("Load Key from config file")
}


func genKey() {
	pri, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		logger.Fatalln("Cannot create key pair", err)
	}
	setGlbKey(pri)
	log("Generate New Key")

	bin, err := x.MarshalECPrivateKey(pri)
	if err != nil {
		logger.Fatalln("Cannot Marshal EC private key", err)
	}
	c.PrivateKey = base64.StdEncoding.EncodeToString(bin)
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

	json, _ := ioutil.ReadAll(r.Body)
	id, err1 := db.Put(json)
	if err1 != nil {
		log("DB insert fail", err1)
		w.WriteHeader(500)
		return
	}

	w.WriteHeader(200)
	log("Deliver", chain, channel, id)
}


func findIpWithConfig() bool {
	var isfind bool

	getLocalIp(func (ip *net.IP) bool {
		if c.Host == ip.String() {
			isfind = true
			return true
		}
		return false
	})
	return isfind
}


func findIpWithStdin() {
	getLocalIp(func (ip *net.IP) bool {
		var cf int
		fmt.Print("Local IP is: ", ip, " ? (y/N) ")
		fmt.Scanf("%c\n", &cf)
		if cf == 'y' {
			c.Host = ip.String()
			saveConfig()
			return true;
		}
		return false
	})
}


/**
 * 如果 setter 返回 true, 则终止 ip 地址的便利
 */
func getLocalIp(setter func(*net.IP) bool) {
	ifaces, err := net.Interfaces()
	if err != nil {
		logger.Fatalln("Cannot get Network Interfaces", err)
	}
	
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			logger.Fatalln("Cannot get Network Address", err)
		}
		
		for _, addr := range addrs {
			var ip net.IP
			// log("addr=", addr)
			switch v := addr.(type) {
				case *net.IPNet:
								ip = v.IP
				case *net.IPAddr:
								ip = v.IP
			}
			if ip != nil 				&&
				 ip.To4() != nil 	&&
				 !ip.IsLoopback() &&
				 !ip.IsUnspecified() {
				if setter(&ip) {
					return
				}
			}
		}
	}
}


/**
 * http://www.ietf.org/rfc/rfc5480.txt
 * http://www.rfc-base.org/txt/rfc-5480.txt
 */
func getPublicKeyStr(key *ecdsa.PublicKey) string {
	bin, err := x.MarshalPKIXPublicKey(key)
	if err != nil {
		logger.Fatalln("Cannot Stringify Public Key,", err)
	}
	return base64.StdEncoding.EncodeToString(bin)
}