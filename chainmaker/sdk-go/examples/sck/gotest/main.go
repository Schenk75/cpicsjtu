package main

import (
	"crypto/ecdsa"
	"crypto/rand"

	"crypto/elliptic"
	"crypto/sha256"
	// "encoding/json"

	// "errors"
	"fmt"
	// "io/ioutil"
	"math/big"
	// "os"
	"strconv"
	"strings"
)

type FileData struct {
	Data Data `json:"DATA"`
	Sig  Sig  `json:"SIG"`
	Pk   Pk   `json:"PK"`
}

type Data struct {
	AvgTemp string `json:"avg(temp)"`
	MinTemp string `json:"min(temp)"`
	MaxTemp string `json:"max(temp)"`
	AvgHum  string `json:"avg(hum)"`
	MinHum  string `json:"min(hum)"`
	MaxHum  string `json:"max(hum)"`
	AvgLig  string `json:"avg(lig)"`
	MinLig  string `json:"min(lig)"`
	MaxLig  string `json:"max(lig)"`
}

// type Data struct {
// 	AvgAirHumidity    string `json:"avg(air_humidity)"`
// 	MinAirHumidity    string `json:"min(air_humidity)"`
// 	MaxAirHumidity    string `json:"max(air_humidity)"`
// 	AvgAirTemperature string `json:"avg(air_temperature)"`
// 	MinAirTemperature string `json:"min(air_temperature)"`
// 	MaxAirTemperature string `json:"max(air_temperature)"`
// 	AvgAtmosphere     string `json:"avg(atmosphere)"`
// 	MinAtmosphere     string `json:"min(atmosphere)"`
// 	MaxAtmosphere     string `json:"max(atmosphere)"`
// 	AvgCo             string `json:"avg(co)"`
// 	MinCo             string `json:"min(co)"`
// 	MaxCo             string `json:"max(co)"`
// 	AvgNo2            string `json:"avg(no2)"`
// 	MinNo2            string `json:"min(no2)"`
// 	MaxNo2            string `json:"max(no2)"`
// 	AvgO3             string `json:"avg(o3)"`
// 	MinO3             string `json:"min(o3)"`
// 	MaxO3             string `json:"max(o3)"`
// }

type Sig struct {
	SIGR string `json:"SIG_r"`
	SIGS string `json:"SIG_s"`
}

type Pk struct {
	PKR string `json:"PK_r"`
	PKS string `json:"PK_s"`
}

type Signature struct {
	r *big.Int
	s *big.Int
}

func main() {
	// f, err := os.Open("/root/jwzhou/paho.mqtt.c/occlum_instance/result_json/20220118T092215Z.json")
	// if err != nil {
	// 	fmt.Printf("文件打开失败 [Err:%s]\n", err.Error())
	// 	return
	// }
	// defer f.Close()

	// byteValue, _ := ioutil.ReadAll(f)
	// var data FileData
	// err = json.Unmarshal(byteValue, &data)
	// if err != nil {
	// 	fmt.Println("解码失败", err.Error())
	// } else {
	// 	fmt.Println("解码成功")
	// 	fmt.Printf("%+v\n", data)
	// }

	// pkx, pky := ParseStr(data.Pk.PKR), ParseStr(data.Pk.PKS)
	// sigr, sigs := ParseStr(data.Sig.SIGR), ParseStr(data.Sig.SIGS)

	// // fmt.Println(pkx)
	// // fmt.Println(pky)
	// // fmt.Println(string(sigr))
	// // fmt.Println(string(sigs))
	
	// x := new(big.Int).SetBytes(pkx)
	// y := new(big.Int).SetBytes(pky)
	// r := new(big.Int).SetBytes(sigr)
	// s := new(big.Int).SetBytes(sigs)

	// pubkey := ecdsa.PublicKey{
	// 	Curve: elliptic.P256(),
	// 	X:     x,
	// 	Y:     y,
	// }

	// sig1 := Signature {
	// 	r,
	// 	s,
	// }

	// // fmt.Println(pubkey)

	// // m, _ := json.Marshal(data.Data)
	// m := []byte("aaa")
	// // fmt.Println("Data: ", string(m))
	
	// origind := "0x2D, 0xD0, 0x9F, 0xEC, 0xA8, 0x3E, 0xAD, 0x86, 0x15, 0x94, 0xA6, 0x9E, 0x72, 0xAA, 0x8C, 0x35, 0xA8, 0xEB, 0xF8, 0x92, 0x05, 0x35, 0x21, 0xE4, 0x21, 0x5E, 0x74, 0xAF, 0x37, 0x40, 0xC0, 0x60"
	// parsed := ParseStr(origind)
	// d := new(big.Int).SetBytes(parsed)
	// privkey := ecdsa.PrivateKey {
	// 	PublicKey: pubkey,
	// 	D: d,
	// }
	// fmt.Println("key1 match?", keyMatch(&privkey))
	// fmt.Printf("sig1: %+v\n", sig1)
	// fmt.Println(VerifySignECC(m, sig1, &pubkey))


	// // m := []byte("aaa")
	// privateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	// // fmt.Println(len(privateKey.D.Bytes()), privateKey.D.Bytes())
	// publicKey := privateKey.PublicKey
	// sig2, _ := ECCSign(m, privateKey)
	// fmt.Println("key2 match?", keyMatch(privateKey))
	// fmt.Printf("sig2: %+v\n", sig2)
	// fmt.Println(VerifySignECC(m, sig2, &publicKey))

	xy := ParseStr("0x81, 0xDC, 0x92, 0x02, 0xBC, 0xFE, 0xD5, 0x0C, 0xF5, 0xBF, 0xA7, 0xDA, 0x49, 0x86, 0x3E, 0xAC, 0xEB, 0x84, 0x28, 0x60, 0x6D, 0xEC, 0xFF, 0xA7, 0x98, 0x3A, 0x2B, 0xE3, 0xAA, 0xCA, 0xDF, 0x7E, 0x0B, 0xD8, 0xB6, 0x56, 0x55, 0x93, 0x93, 0xB3, 0xD8, 0x1E, 0x06, 0x2C, 0x29, 0x4C, 0x4D, 0x36, 0xA7, 0x64, 0x3D, 0xBF, 0x9C, 0x08, 0xCB, 0xD9, 0x2D, 0xF5, 0xB5, 0xCD, 0xF3, 0xB2, 0x53, 0x82")
	// px3 := new(big.Int).SetBytes(ParseStr("0x4A, 0x24, 0x48, 0x5A, 0x17, 0x97, 0xDC, 0x5A, 0x27, 0x98, 0x08, 0x88, 0xC9, 0x4A, 0x2E, 0xCB, 0x8A, 0xDE, 0x9E, 0x98, 0x19, 0x42, 0xB2, 0x9A, 0xB4, 0x74, 0x33, 0xE6, 0x1E, 0x29,0xE0, 0x8D"))
	// py3 := new(big.Int).SetBytes(ParseStr("0x14, 0xE6, 0x5B, 0x8E, 0x7F, 0xF4, 0xF1, 0x52, 0x6D, 0x42, 0x7A, 0x84, 0x61, 0x39, 0x41, 0xAD, 0x07, 0xAA, 0xDA, 0xF8, 0x4A, 0xC7, 0xDD, 0x8F, 0xC5, 0x4D, 0x4F, 0xB9, 0x9D, 0x8A, 0x7E, 0xF6"))
	px3 := new(big.Int).SetBytes(xy[:len(xy)/2])
	py3 := new(big.Int).SetBytes(xy[len(xy)/2:])
	pubkey3 := ecdsa.PublicKey {
		Curve: elliptic.P256(),
		X: px3,
		Y: py3,
	}
	d3 := new(big.Int).SetBytes(ParseStr("0x2D, 0x39, 0xE8, 0x06, 0x72, 0x7C, 0xE3, 0x69, 0x13, 0xC9, 0xDC, 0x10, 0x7B, 0xC9, 0xFE, 0x50, 0xD9, 0x3F, 0xC2, 0xB8, 0xD9, 0x61, 0x9C, 0x16, 0xA1, 0x01, 0xC4, 0x8D, 0x43, 0x0F, 0x1A, 0xA8"))
	privkey3 := ecdsa.PrivateKey {
		PublicKey: pubkey3,
		D: d3,
	}
	fmt.Println("key3 match?", keyMatch(&privkey3))
	rs := ParseStr("0xCD, 0x7F, 0x22, 0x7C, 0x48, 0xC8, 0x0F, 0xC9, 0xC0, 0x76, 0xBF, 0xCD, 0xCB, 0x33, 0x22, 0x9D, 0x48, 0xC9, 0xDB, 0x72, 0x85, 0xB4, 0x5A, 0x34, 0x74, 0x99, 0xE3, 0x32, 0x60, 0xC4, 0xDF, 0xC5, 0xCE, 0xE9, 0xC5, 0x36, 0x4B, 0x31, 0x1E, 0x29, 0xE5, 0xD6, 0xDF, 0x03, 0xDC, 0x11, 0x42, 0x8D, 0x3C, 0x85, 0x51, 0x9C, 0xDE, 0x43, 0x21, 0xB3, 0xDA, 0x2D, 0xD7, 0xD0, 0x92, 0x95, 0x3C, 0x9A")
	r3 := new(big.Int).SetBytes(rs[:len(rs)/2])
	s3 := new(big.Int).SetBytes(rs[len(rs)/2:])
	sig3 := Signature {
		r: r3,
		s: s3,
	}
	fmt.Println(VerifySignECC([]byte(`{"avg(temp)":"24.8","min(temp)":"24.8","max(temp)":"24.8","avg(hum)":"26.6","min(hum)":"26.6","max(hum)":"26.6","avg(lig)":"636.0","min(lig)":"636.0","max(lig)":"636.0"}`), sig3, &pubkey3))
}

func ParseStr(str string) []byte {
	s := strings.Split(str, ",")
	res := []byte{}

	for i := range s {
		str := strings.TrimSpace(s[i])
		num, err := strconv.ParseInt(str, 0, 16)
		if err != nil {
			panic(err)
		}
		res = append(res, uint8(num))
	}

	return res
}

func VerifySignECC(msg []byte, sig Signature, pubkey *ecdsa.PublicKey) bool {
	r, s := sig.r, sig.s
	//计算哈希值
	hash := sha256.New()
	//填入数据
	hash.Write(msg)
	bytes := hash.Sum(nil)
	// fmt.Printf("hash: %v\n", bytes)

	verify := ecdsa.Verify(pubkey, bytes, r, s)
	return verify
}

func ECCSign(plainText []byte, privkey *ecdsa.PrivateKey) (Signature, error) {
	//计算哈希值
	hash := sha256.New()
	//填入数据
	hash.Write(plainText)
	bytes := hash.Sum(nil)
	// fmt.Printf("hash: %v\n", bytes)
	// sign
	r, s, _ := ecdsa.Sign(rand.Reader, privkey, bytes)
	// fmt.Println(len(r.Bytes()), r.Bytes())

	sig := Signature {
		r,
		s,
	}

	return sig, nil
}

func keyMatch(privkey *ecdsa.PrivateKey) bool {
	pubkey := privkey.PublicKey
	c := elliptic.P256()
	x, y := c.ScalarBaseMult(privkey.D.Bytes())
	return x.Cmp(pubkey.X) == 0 && y.Cmp(pubkey.Y) == 0
}