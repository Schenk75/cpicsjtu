/*
Copyright (C) BABEC. All rights reserved.
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"fmt"
	"log"
	"strconv"
	"time"
	"crypto/sha256"
	"io/ioutil"
	"os"

	"chainmaker.org/chainmaker/common/v2/crypto"
	"chainmaker.org/chainmaker/pb-go/v2/common"
	sdk "chainmaker.org/chainmaker/sdk-go/v2"
	"chainmaker.org/chainmaker/sdk-go/v2/examples"
)

const (
	createContractTimeout = 5
	claimContractName     = "upload001"
	claimVersion          = "2.0.0"
	claimByteCodePath     = "./contract/upload.wasm"

	sdkConfigOrg1Client1Path = "../sdk_configs/sdk_config_org1_client1.yml"
)

func main() {
	uploadData()
}

func uploadData() {
	fmt.Println("====================== create client ======================")
	client, err := examples.CreateChainClientWithSDKConf(sdkConfigOrg1Client1Path)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println("====================== 创建合约 ======================")
	usernames := []string{examples.UserNameOrg1Admin1, examples.UserNameOrg2Admin1, examples.UserNameOrg3Admin1, examples.UserNameOrg4Admin1}
	create(client, true, usernames...)

	fmt.Println("====================== 调用合约 ======================")
	fileName, err := invoke(client, "save", true)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println("====================== 执行合约查询接口 ======================")
	kvs := []*common.KeyValuePair{
		{
			Key:   "file_name",
			Value: []byte(fileName),
		},
	}
	query(client, "find_by_file_name", kvs)
}

func create(client *sdk.ChainClient, withSyncResult bool, usernames ...string) {

	resp, err := createUserContract(client, claimContractName, claimVersion, claimByteCodePath,
		common.RuntimeType_WASMER, []*common.KeyValuePair{}, withSyncResult, usernames...)
	if err != nil {
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

	//endorsers, err := examples.GetEndorsers(payload, usernames...)
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

func invoke(client *sdk.ChainClient, method string, withSyncResult bool) (string, error) {
	curTime := strconv.FormatInt(time.Now().Unix(), 10)

	f, err := os.Open("./test.json")
    if err != nil {
        fmt.Printf("Cannot open file [Err:%s]", err.Error())
        return "", err
    }
    defer f.Close()
	byteValue, _ := ioutil.ReadAll(f)

	fileHash := fmt.Sprintf("%x", sha256.Sum256(byteValue))
	fileData := string(byteValue)
	fileName := fmt.Sprintf("file_%s", curTime)
	kvs := []*common.KeyValuePair{
		{
			Key:   "time",
			Value: []byte(curTime),
		},
		{
			Key:   "file_hash",
			Value: []byte(fileHash),
		},
		{
			Key:   "file_data",
			Value: []byte(fileData),
		},
		{
			Key:   "file_name",
			Value: []byte(fileName),
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

func query(client *sdk.ChainClient, method string, kvs []*common.KeyValuePair) {
	resp, err := client.QueryContract(claimContractName, method, kvs, -1)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Printf("QUERY claim contract resp: %+v\n", resp)
}
