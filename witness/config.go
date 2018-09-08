package witness

import (
	"crypto/elliptic"
	"os"
	"flag"
	"encoding/json"
	"fmt"
	"encoding/base64"
	"crypto/ecdsa"
	"crypto/rand"

	x "crypto/x509"
	logger "log"
)

type Config struct {
	PrivateKey 	string
	Port 				int
	Host        string
	URLxBoson 	string
	ID          string
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