/*
Copyright (C) BABEC. All rights reserved.
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"os"
	"strconv"
	"strings"
	"time"

	"chainmaker.org/chainmaker/common/v2/crypto"
	"chainmaker.org/chainmaker/pb-go/v2/common"
	sdk "chainmaker.org/chainmaker/sdk-go/v2"
	"chainmaker.org/chainmaker/sdk-go/v2/examples"
)

const (
	createContractTimeout    = 5
	claimContractName        = "upload001"
	pubkeyContractName		 = "pubkey003"
	claimVersion             = "2.0.0"
	claimByteCodePath        = "/root/chainmaker/sdk-go/examples/sck/contract/upload.wasm"
	dataPath                 = "/root/jwzhou/paho.mqtt.c/occlum_instance/result_json/"
	sdkConfigOrg1Client1Path = "/root/chainmaker/sdk-go/examples/sdk_configs/sdk_config_org1_client1.yml"
	printLog 				 = true
)

func PrintLog(format string, a ...interface{}) {
	if printLog {
		log.Printf(format, a...)
	}
}

type FileData struct {
	Data 		Data 		`json:"DATA"`
	Sig  		Sig  		`json:"SIG"`
	PkFlag  	PkFlag   	`json:"PK_FLAG"`
	Prediction 	Prediction	`json:"PREDICTION"`
}

type Data struct {
	AvgTemp      string `json:"avg(temp)"`
	MinTemp      string `json:"min(temp)"`
	MaxTemp      string `json:"max(temp)"`
	AvgHum       string `json:"avg(hum)"`
	MinHum       string `json:"min(hum)"`
	MaxHum       string `json:"max(hum)"`
	AvgLig       string `json:"avg(lig)"`
	MinLig       string `json:"min(lig)"`
	MaxLig       string `json:"max(lig)"`
	SerialNumber string `json:"serialNumber"`
	SensorType   string `json:"sensorType"`
	SensorModel  string `json:"sensorModel"`
	ToTeeTime    string `json:"ToTeeTime"`
}

type Sig struct {
	SIGR string `json:"SIG_r"`
	SIGS string `json:"SIG_s"`
}

type PkFlag struct {
	ID string `json:"id"`
}

type Prediction struct {
	TempAfter30Mins float64 `json:"Temp_After_30mins"`
}

// 经过处理的签名
type Signature struct {
	R *big.Int
	S *big.Int
}

// 从链上读取的已注册公钥信息
type PkFromChain struct {
	Pubkey   string `json:"pubkey"`
	PubkeyID string `json:"pubkey_id"`
	Orgid    string `json:"orgid"`
	Time     int    `json:"time"`
}

type PkCurve struct {
	Curve struct {
		P       *big.Int  	`json:"P"`
		N       *big.Int  	`json:"N"`
		B       *big.Int  	`json:"B"`
		Gx      *big.Int  	`json:"Gx"`
		Gy      *big.Int  	`json:"Gy"`
		BitSize int    		`json:"BitSize"`
		Name    string 		`json:"Name"`
	} `json:"Curve"`
	X *big.Int `json:"X"`
	Y *big.Int `json:"Y"`
}

func main() {
	uploadData(os.Args[1])
}

func uploadData(data string) bool {
	fmt.Println("====================== create client ======================")
	start := time.Now()
	client, err := examples.CreateChainClientWithSDKConf(sdkConfigOrg1Client1Path)
	if err != nil {
		log.Printf("%v", err)
		return false
	}
	fmt.Println("time: ", time.Since(start))

	// fmt.Println("====================== 创建合约 ======================")
	// start = time.Now()
	// usernames := []string{examples.UserNameOrg1Admin1, examples.UserNameOrg2Admin1, examples.UserNameOrg3Admin1, examples.UserNameOrg4Admin1}
	// create(client, true, usernames...)
	// fmt.Println("time: ", time.Since(start))

	fmt.Println("====================== 调用合约 ======================")
	start = time.Now()
	fileName, err := invoke(client, "save", true, data)
	if err != nil {
		log.Printf("%v", err)
		return false
	}
	PrintLog("file name: %v\n", fileName)
	fmt.Println("time: ", time.Since(start))

	fmt.Println("====================== 执行合约查询接口 ======================")
	start = time.Now()
	kvs := []*common.KeyValuePair{
		{
			Key:   "file_name",
			Value: []byte(fileName),
		},
	}
	res := query(client, "find_by_file_name", claimContractName, kvs)
	PrintLog(string(res.Result))
	fmt.Println("time: ", time.Since(start))
	return true
}

// 创建合约
func create(client *sdk.ChainClient, withSyncResult bool, usernames ...string) {
	resp, err := createUserContract(client, claimContractName, claimVersion, claimByteCodePath,
		common.RuntimeType_WASMER, []*common.KeyValuePair{}, withSyncResult, usernames...)

	if err != nil {
		// 合约已存在，直接执行
		if err.Error() == "contract exist" {
			PrintLog("contract exist")
			return
		}
		log.Fatalln(err)
	}

	PrintLog("CREATE claim contract resp: %+v\n", resp)
}

// 调用合约，数据上链
func invoke(client *sdk.ChainClient, method string, withSyncResult bool, data string) (string, error) {
	curTime := strconv.FormatInt(time.Now().Unix(), 10)

	fdata, fileData := readDataFromFile(dataPath + data)

	fileName := fmt.Sprintf("file_%s", curTime)
	// fmt.Println(fdata)

	// 根据PkFlag从链上读取公钥
	start := time.Now()
	pubkey := getPkFromChain(client, fdata.PkFlag.ID)
	fmt.Println("get pk time: ", time.Since(start))
	
	// fmt.Println("pubkey: ", pubkey)
	pubkeyStr, _ := json.Marshal(pubkey)

	r := new(big.Int).SetBytes(ParseStr(fdata.Sig.SIGR))
	s := new(big.Int).SetBytes(ParseStr(fdata.Sig.SIGS))
	sig := Signature{
		R: r,
		S: s,
	}
	sigStr, _ := json.Marshal(sig)
	// fmt.Println(string(sig_str))

	start = time.Now()
	if !verifySig(fileData, sig, pubkey) {
		return "", errors.New("invalid signature")
	} else {
		PrintLog("[+] Valid signature")
	}
	fmt.Println("verify sig time: ", time.Since(start))

	kvs := []*common.KeyValuePair{
		{
			Key:   "time",
			Value: []byte(curTime),
		},
		{
			Key:   "file_sig",
			Value: []byte(sigStr),
		},
		{
			Key:   "file_data",
			Value: []byte(fileData),
		},
		{
			Key:   "file_name",
			Value: []byte(fileName),
		},
		{
			Key:   "pubkey",
			Value: []byte(pubkeyStr),
		},
	}

	err := invokeUserContract(client, claimContractName, method, "", kvs, withSyncResult)
	if err != nil {
		return "", err
	}

	return fileName, nil
}

// 查询链上数据
func query(client *sdk.ChainClient, method, contractName string, kvs []*common.KeyValuePair) *common.ContractResult {
	resp, err := client.QueryContract(contractName, method, kvs, -1)
	if err != nil {
		log.Fatalln(err)
	}

	// PrintLog("QUERY claim contract resp: %+v\n", resp.ContractResult)
	return resp.ContractResult
}

// 检验公钥是否合法
func verifyPk(client *sdk.ChainClient, pkId string) (*common.ContractResult, bool) {
	kvs := []*common.KeyValuePair{
		{
			Key:   "pubkey_id",
			Value: []byte(pkId),
		},
	}

	res := query(client, "find_by_pubkey_id", pubkeyContractName, kvs)
	return res, true
}

// 检验签名
func verifySig(msg []byte, sig Signature, pubkey *ecdsa.PublicKey) bool {
	r, s := sig.R, sig.S
	//计算哈希值
	hash := sha256.New()
	//填入数据
	hash.Write(msg)
	bytes := hash.Sum(nil)
	// PrintLog("hash: %X\n", bytes)

	verify := ecdsa.Verify(pubkey, bytes, r, s)
	return verify
}

// 处理json里的原始字符串
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

// 从文件读出数据
func readDataFromFile(fileName string) (FileData, []byte) {
	f, err := os.Open(fileName)
	// f, err := os.Open("./test_data.json")

	if err != nil {
		log.Fatalf("Cannot open file [Err:%s]", err.Error())
	}
	defer f.Close()
	byteValue, _ := ioutil.ReadAll(f)

	var fdata FileData
	err = json.Unmarshal(byteValue, &fdata)
	if err != nil {
		log.Fatalf("Unmarshal fail %s", err.Error())
	}
	// PrintLog("fdata: %+v\n", fdata)

	bf := bytes.NewBuffer([]byte{})
	jsonEncoder := json.NewEncoder(bf)
	jsonEncoder.SetEscapeHTML(false)
	jsonEncoder.Encode(fdata.Data)
	fileData := []byte(strings.TrimSpace(bf.String()))
	// fmt.Println(fileData)
	return fdata, fileData
}

// 根据id从链上获取已注册的公钥
func getPkFromChain(client *sdk.ChainClient, pkId string) *ecdsa.PublicKey {
	getPkRes, _ := verifyPk(client, pkId)
	if len(getPkRes.Result) == 0 {
		log.Fatalln("invalid public key")
	}
	PrintLog("[+] Valid public key")

	var pkfc PkFromChain
	err := json.Unmarshal(getPkRes.Result, &pkfc)
	if err != nil {
		log.Fatalln("Unmarshal fail", err.Error())
	}
	// PrintLog("Pk From Chain: %+v\n", pkfc)
	// fmt.Println(pkfc.Pubkey)

	var pkCurve PkCurve
	err = json.Unmarshal([]byte(pkfc.Pubkey), &pkCurve)
	if err != nil {
		fmt.Println("Unmarshal fail", err.Error())
	}
	// PrintLog("Pk Curve: %+v\n", pkCurve)

	return &ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     pkCurve.X,
		Y:     pkCurve.Y,
	}
}

func createUserContract(client *sdk.ChainClient, contractName, version, byteCodePath string, runtime common.RuntimeType,
	kvs []*common.KeyValuePair, withSyncResult bool, usernames ...string) (*common.TxResponse, error) {

	payload, err := client.CreateContractCreatePayload(contractName, version, byteCodePath, runtime, kvs)
	if err != nil {
		return nil, err
	}

	// endorsers, err := examples.GetEndorsers(payload, usernames...)
	endorsers, err := examples.GetEndorsersWithAuthType(crypto.HashAlgoMap[client.GetHashType()],
		client.GetAuthType(), payload, usernames...)
	if err != nil {
		return nil, err
	}

	resp, err := client.SendContractManageRequest(payload, endorsers, createContractTimeout, withSyncResult)
	if err != nil {
		return nil, err
	}

	err = examples.CheckProposalRequestResp(resp, true)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func invokeUserContract(client *sdk.ChainClient, contractName, method, txId string,
	kvs []*common.KeyValuePair, withSyncResult bool) error {

	resp, err := client.InvokeContract(contractName, method, txId, kvs, -1, withSyncResult)
	if err != nil {
		return err
	}

	if resp.Code != common.TxStatusCode_SUCCESS {
		return fmt.Errorf("invoke contract failed, [code:%d]/[msg:%s]", resp.Code, resp.Message)
	}

	if !withSyncResult {
		PrintLog("invoke contract success, resp: [code:%d]/[msg:%s]/[txId:%s]\n", resp.Code, resp.Message, resp.ContractResult.Result)
	} else {
		PrintLog("invoke contract success, resp: [code:%d]/[msg:%s]/[contractResult:%s]\n", resp.Code, resp.Message, resp.ContractResult)
	}

	return nil
}