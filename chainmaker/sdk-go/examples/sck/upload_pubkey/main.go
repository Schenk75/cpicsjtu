/*
Copyright (C) BABEC. All rights reserved.
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/json"
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
	claimContractName        = "pubkey003"
	claimVersion             = "2.0.0"
	claimByteCodePath        = "./contract/pubkey_upload.wasm"
	pkPath                 	 = "/root/jwzhou/paho.mqtt.c/occlum_instance/pb_json/"
	sdkConfigOrg1Client1Path = "../sdk_configs/sdk_config_org1_client1.yml"
)

type Pk struct {
	PKR string `json:"PK_r"`
	PKS string `json:"PK_s"`
}

func main() {
	uploadPk(os.Args[1])
}


func uploadPk(pkName string) {
	fmt.Println("====================== create client ======================")
	client, err := examples.CreateChainClientWithSDKConf(sdkConfigOrg1Client1Path)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println("====================== 创建合约 ======================")
	usernames := []string{examples.UserNameOrg1Admin1, examples.UserNameOrg2Admin1, examples.UserNameOrg3Admin1, examples.UserNameOrg4Admin1}
	create(client, true, usernames...)

	fmt.Println("====================== 调用合约 ======================")
	pkId, err := invoke(client, "save", true, pkName)
	if err != nil {
		log.Fatalln(err)
	}
	// fmt.Println(pkId)

	fmt.Println("====================== 执行合约查询接口 ======================")
	// pkId := "20220221T055011Z"
	kvs := []*common.KeyValuePair{
		{
			Key:   "pubkey_id",
			Value: []byte(pkId),
		},
	}
	res := query(client, "find_by_pubkey_id", kvs)
	fmt.Println(res)
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

// 调用合约，数据上链
func invoke(client *sdk.ChainClient, method string, withSyncResult bool, pkName string) (string, error) {
	curTime := strconv.FormatInt(time.Now().Unix(), 10)

	f, err := os.Open(pkPath + pkName)
	// f, err := os.Open("./test2.json")

	if err != nil {
		fmt.Printf("Cannot open file [Err:%s]", err.Error())
		return "", err
	}
	defer f.Close()
	byteValue, _ := ioutil.ReadAll(f)
	// fmt.Println("byteValue: ", string(byteValue))

	var pk Pk
	err = json.Unmarshal(byteValue, &pk)
	if err != nil {
		fmt.Println("Unmarshal fail", err.Error())
	}
	fmt.Printf("pk: %+v\n", pk)

	x := new(big.Int).SetBytes(ParseStr(pk.PKR))
	y := new(big.Int).SetBytes(ParseStr(pk.PKS))
	pubkey := ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     x,
		Y:     y,
	}
	fmt.Println("origin pubkey: ", pubkey)

	pubkeyStr, _ := json.Marshal(pubkey)
	orgid := examples.OrgId1

	pkId := strings.Split(pkName, ".")[0]
	fmt.Println("pkId: ", pkId)

	kvs := []*common.KeyValuePair{
		{
			Key:   "time",
			Value: []byte(curTime),
		},
		{
			Key:   "pubkey",
			Value: []byte(pubkeyStr),
		},
		{
			Key:   "pubkey_id",
			Value: []byte(pkId),
		},
		{
			Key:   "orgid",
			Value: []byte(orgid),
		},
	}

	err = invokeUserContract(client, claimContractName, method, "", kvs, withSyncResult)
	if err != nil {
		return "", err
	}

	return string(pkId), nil
}

// 查询链上数据
func query(client *sdk.ChainClient, method string, kvs []*common.KeyValuePair) *common.ContractResult {
	resp, err := client.QueryContract(claimContractName, method, kvs, -1)
	if err != nil {
		log.Fatalln(err)
	}

	// fmt.Printf("QUERY claim contract resp: %+v\n", resp.ContractResult)
	return resp.ContractResult
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
		fmt.Printf("invoke contract success, resp: [code:%d]/[msg:%s]/[txId:%s]\n", resp.Code, resp.Message, resp.ContractResult.Result)
	} else {
		fmt.Printf("invoke contract success, resp: [code:%d]/[msg:%s]/[contractResult:%s]\n", resp.Code, resp.Message, resp.ContractResult)
	}

	return nil
}