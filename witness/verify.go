package witness
 
import (
	"encoding/asn1"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"errors"
	"crypto/ecdsa"
	"math/big"
)

var B64Decode = base64.URLEncoding.WithPadding(base64.NoPadding)

type Sign struct {
	id string
	si []byte
}

type Block struct {
	_type int
	create uint64
	key, data, previousHash, previousKey, hash, chaincodeKey []byte
	userid, apiPath, apiHash string
	sign []Sign
}

// ecdsa 签名分为两个部分, 并使用 asn1 格式保存.
type ecdsaSignature struct {
	R, S *big.Int
}

const (
	FAIL_BLOCK        		  = 0;
	GENESIS_BLOCK           = 1;
	NORM_DATA_BLOCK         = 2;
	ENCRYPTION_DATA_BLOCK   = 3;
	CHAINCODE_CONTENT_BLOCK = 4;
	MESSAGE_BLOCK           = 5;
)


func verify(b map[string]interface{}) error {
	var block Block;
	if err := readBlockFrom(b, &block); err != nil {
		return err
	}

	if err := verifyHash(&block); err != nil {
		return err
	}

	if err := verifySign(&block); err != nil {
		return err
	}

	if err := verifyBusiness(&block); err != nil {
		return err
	}
	// log("Success verify", b["key"])
	return nil
}


func verifySign(block *Block) error {
	h := sha256.New()
	h.Write(block.key)
	h.Write(block.data)
	h.Write([]byte(block.userid))
	h.Write(int64bytes(block.create))

	if block._type != GENESIS_BLOCK {
		h.Write(block.previousKey)
		h.Write(block.previousHash)
	}

	switch (block._type) {
	case CHAINCODE_CONTENT_BLOCK:
		h.Write([]byte(block.apiPath))
		h.Write([]byte(block.apiHash))

	case GENESIS_BLOCK:
		// 不验证创世块
		return nil;

	case NORM_DATA_BLOCK, ENCRYPTION_DATA_BLOCK:
		h.Write(block.chaincodeKey)
	}

	// 如果是创世块, 该方法一定返回错误
	es, err := findSignMyself(block)
	if err != nil {
		return err
	}

	if !ecdsa.Verify(&prikey.PublicKey, h.Sum(nil), es.R, es.S) {
		return errors.New("verfiy signature fail")
	}
	return nil
}


func findSignMyself(block *Block) (*ecdsaSignature, error) {
	var sign []byte
	for _, s := range block.sign {
		if s.id == c.ID {
			sign = s.si
		}
	}
	if len(sign) <= 0 {
		return nil, errors.New("cannot find signature with verify")
	}
	var es ecdsaSignature
	if _, err := asn1.Unmarshal(sign, &es); err != nil {
		return nil, err
	}
	return &es, nil;
}


func verifyHash(block *Block) error {
	h := sha256.New()
	h.Write(block.key)
	h.Write(block.data)
	h.Write([]byte(block.userid))
	h.Write([]byte{ (byte)(block._type & 0xFF), 0, 0, 0 })
	h.Write(int64bytes(block.create))

	h.Write(block.chaincodeKey)
	h.Write([]byte(block.apiPath))
	h.Write([]byte(block.apiHash))

	if len(block.sign) > 0 {
		for _, s := range block.sign {
			h.Write([]byte(s.id))
			h.Write(s.si)
		}
	}
	
	h.Write(block.previousHash)
	h.Write(block.previousKey)
	hash := h.Sum(nil)

	if !eq(hash, block.hash) {
		fmt.Println("H", B64Decode.EncodeToString(hash), 
			B64Decode.EncodeToString(block.hash))
		return errors.New("bad hash")
	}
	return nil
}


func eq(a []byte, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := len(a)-1; i>=0; i-- {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}


func int64bytes(l uint64) []byte {
	return []byte {
		(byte) (l & 0xFF),
		(byte) ((l>>8 ) & 0xFF),
		(byte) ((l>>16) & 0xFF),
    (byte) ((l>>24) & 0xFF),

    (byte) ((l>>32) & 0xFF),
    (byte) ((l>>40) & 0xFF),
    (byte) ((l>>48) & 0xFF),
    (byte) ((l>>56) & 0xFF),
	}
}


func ReadBlockFrom(b map[string]interface{}, block *Block) (error) {
	return readBlockFrom(b, block)
}


func readBlockFrom(b map[string]interface{}, block *Block) (error) {
	var e error
	var ok bool
	var _tmp float64

	if _tmp, ok = b["type"].(float64); !ok {
		return errors.New("bad attribute 'type'")
	}
	block._type = int(_tmp)
	if e = decode64(b["key"], &block.key); e != nil {
		return e
	}
	if e = decode64(b["data"], &block.data); e != nil {
		return e
	}

	if b["previousHash"] != nil {
		if e = decode64(b["previousHash"], &block.previousHash); e != nil {
			return e
		}
		if e = decode64(b["previousKey"], &block.previousKey); e != nil {
			return e
		}
	}

	if e = decode64(b["hash"], &block.hash); e != nil {
		return e
	}
	if block.userid, ok = b["userid"].(string); !ok {
		return errors.New("bad attribute 'userid'")
	}
	if _tmp, ok = b["create"].(float64); !ok {
		return errors.New("bad attribute 'create'")
	}
	block.create = uint64(_tmp)

	if b["apiPath"] != nil {
		if block.apiPath, ok = b["apiPath"].(string); !ok {
			return errors.New("bad attribute 'apiPath'")
		}
		if block.apiHash, ok = b["apiHash"].(string); !ok {
			return errors.New("bad attribute 'apiHash'")
		}
	}

	if b["sign"] != nil {
		var signList []interface{}
		if signList, ok = b["sign"].([]interface{}); !ok {
			return errors.New("bad attribute 'sign'")
		}
		
		for _, s := range signList {
			var sign Sign
			item := s.(map[string]interface{})
			sign.id = item["id"].(string)
			if e = decode64(item["si"], &sign.si); e != nil {
				return errors.New("bad public key")
			}
			block.sign = append(block.sign, sign)
		}
	}

	if b["chaincodeKey"] != nil {
		if e = decode64(b["chaincodeKey"], &block.chaincodeKey); e != nil {
			return e;
		}
	}
	return nil
}


func decode64(v interface{}, r *[]byte) error {
	s, ok := v.(string)
	if !ok {
		return errors.New("decode64 value not string")
	}
	var e error
	*r, e = B64Decode.DecodeString(s) 
	return e
}