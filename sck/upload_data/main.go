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
	claimVersion             = "2.0.0"
	claimByteCodePath        = "./contract/upload.wasm"
	dataPath                 = "/root/jwzhou/paho.mqtt.c/occlum_instance/result_json/"
	sdkConfigOrg1Client1Path = "../sdk_configs/sdk_config_org1_client1.yml"
)

type FileData struct {
	Data Data `json:"DATA"`
	Sig  Sig  `json:"SIG"`
	Pk   Pk   `json:"PK"`
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

// 经过处理的签名
type Signature struct {
	R *big.Int
	S *big.Int
}

func main() {
	uploadData(os.Args[1])
}

func uploadData(data string) {
	fmt.Println("====================== create client ======================")
	client, err := examples.CreateChainClientWithSDKConf(sdkConfigOrg1Client1Path)
	if err != nil {
		log.Fatalln(err)
	}

	// fmt.Println("====================== 创建合约 ======================")
	// usernames := []string{examples.UserNameOrg1Admin1, examples.UserNameOrg2Admin1, examples.UserNameOrg3Admin1, examples.UserNameOrg4Admin1}
	// create(client, true, usernames...)

	fmt.Println("====================== 调用合约 ======================")
	fileName, err := invoke(client, "save", true, data)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(fileName)

	// fmt.Println("====================== 执行合约查询接口 ======================")
	// kvs := []*common.KeyValuePair{
	// 	{
	// 		Key:   "file_name",
	// 		Value: []byte(fileName),
	// 	},
	// }
	// query(client, "find_by_file_name", kvs)
}

// 创建合约
func create(client *sdk.ChainClient, withSyncResult bool, usernames ...string) {
	resp, err := createUserContract(client, claimContractName, claimVersion, claimByteCodePath,
		common.RuntimeType_WASMER, []*common.KeyValuePair{}, withSyncResult, usernames...)

	if err != nil {
		if err.Error() == "contract exist" {
			fmt.Println("contract exist")
			return
		}
		log.Fatalln(err)
	}

	fmt.Printf("CREATE claim contract resp: %+v\n", resp)
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

// 调用合约，数据上链
func invoke(client *sdk.ChainClient, method string, withSyncResult bool, data string) (string, error) {
	curTime := strconv.FormatInt(time.Now().Unix(), 10)

	f, err := os.Open(dataPath + data)
	// f, err := os.Open("./test2.json")

	if err != nil {
		fmt.Printf("Cannot open file [Err:%s]", err.Error())
		return "", err
	}
	defer f.Close()
	byteValue, _ := ioutil.ReadAll(f)

	var fdata FileData
	err = json.Unmarshal(byteValue, &fdata)
	if err != nil {
		fmt.Println("Unmarshal fail", err.Error())
	}
	// fmt.Printf("fdata: %+v\n", fdata)

	bf := bytes.NewBuffer([]byte{})
	jsonEncoder := json.NewEncoder(bf)
	jsonEncoder.SetEscapeHTML(false)
	jsonEncoder.Encode(fdata.Data)
	fileData := []byte(strings.TrimSpace(bf.String()))
	// fmt.Println(fileData)

	fileName := fmt.Sprintf("file_%s", curTime)

	// fmt.Println(fdata)

	x := new(big.Int).SetBytes(ParseStr(fdata.Pk.PKR))
	y := new(big.Int).SetBytes(ParseStr(fdata.Pk.PKS))

	pubkey := ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     x,
		Y:     y,
	}

	r := new(big.Int).SetBytes(ParseStr(fdata.Sig.SIGR))
	s := new(big.Int).SetBytes(ParseStr(fdata.Sig.SIGS))
	sig := Signature{
		R: r,
		S: s,
	}

	if !verifyPk(pubkey) {
		return "", errors.New("invalid public key")
	}
	if !verifySig(fileData, sig, &pubkey) {
		return "", errors.New("invalid signature")
	}

	pubkey_str, _ := json.Marshal(pubkey)
	sig_str, _ := json.Marshal(sig)
	// fmt.Println(string(sig_str))

	kvs := []*common.KeyValuePair{
		{
			Key:   "time",
			Value: []byte(curTime),
		},
		{
			Key:   "file_sig",
			Value: []byte(sig_str),
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
			Value: []byte(pubkey_str),
		},
	}

	err = invokeUserContract(client, claimContractName, method, "", kvs, withSyncResult)
	if err != nil {
		return "", err
	}

	return fileName, nil
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
		fmt.Printf("invoke contract success, resp: [code:%d]/[msg:%s]/[txId:%s]\n", resp.Code, resp.Message, resp.ContractResult.Result)
	} else {
		fmt.Printf("invoke contract success, resp: [code:%d]/[msg:%s]/[contractResult:%s]\n", resp.Code, resp.Message, resp.ContractResult)
	}

	return nil
}

// 查询链上数据
func query(client *sdk.ChainClient, method string, kvs []*common.KeyValuePair) {
	resp, err := client.QueryContract(claimContractName, method, kvs, -1)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Printf("QUERY claim contract resp: %+v\n", resp.ContractResult)
}

// 检验公钥是否合法
func verifyPk(pubkey ecdsa.PublicKey) bool {
	return true
}

// 检验签名
func verifySig(msg []byte, sig Signature, pubkey *ecdsa.PublicKey) bool {
	r, s := sig.R, sig.S
	//计算哈希值
	hash := sha256.New()
	//填入数据
	hash.Write(msg)
	bytes := hash.Sum(nil)
	// fmt.Printf("hash: %X\n", bytes)

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
