package main

import (
	"errors"
	"io"
	"math/big"
	"os"
	"testing"

	"chainmaker.org/chainmaker/sdk-go/v2/examples"
)

const (
	absolutePath = "/root/jwzhou/paho.mqtt.c/occlum_instance/result_json/test_test.json"
)

// 数据上链流程测试
func TestUploadData(t *testing.T) {
	err := createFile()
	if err != nil {
		t.Errorf("Create File Error:%v", err)
	}
	res := uploadData("test_test.json")
	if !res {
		t.Errorf("Upload Data Fail\n")
	}
}

// 测试ParseStr函数
func TestParse(t *testing.T) {
	str := "0xE6,0xCD,0x9C,0xDE,0x21,0xC2,0x19,0x35,0x7B,0xF7,0xC8,0xE2,0x6B,0xE4,0x3F,0x04,0xE5,0x96,0x2E,0xF5,0xB3,0x07,0xB0,0x5C,0x0C,0x4D,0xEE,0xBC,0x2A,0xEE,0x96,0xBA"
	expected := []byte{230, 205, 156, 222, 33, 194, 25, 53, 123, 247, 200, 226, 107, 228, 63, 4, 229, 150, 46, 245, 179, 7, 176, 92, 12, 77, 238, 188, 42, 238, 150, 186}
	parse := ParseStr(str)
	for i := range expected {
		if expected[i] != parse[i] {
			t.Errorf("Parse Result Not Right\n")
		}
	}
}

// 公钥验证测试
func TestVerifyPk(t *testing.T) {
	client, err := examples.CreateChainClientWithSDKConf(sdkConfigOrg1Client1Path)
	if err != nil {
		t.Errorf("%v", err)
	}
	_, res := verifyPk(client, "test_test")
	if !res {
		t.Errorf("Verify Public Key Fail\n")
	}
}

// 签名验证测试
func TestVerifySig(t *testing.T) {
	err := createFile()
	if err != nil {
		t.Errorf("Create File Error:%v", err)
	}
	client, err := examples.CreateChainClientWithSDKConf(sdkConfigOrg1Client1Path)
	if err != nil {
		t.Errorf("%v", err)
	}
	fdata, fileData := readDataFromFile(absolutePath)
	pubkey := getPkFromChain(client, fdata.PkFlag.ID)
	r := new(big.Int).SetBytes(ParseStr(fdata.Sig.SIGR))
	s := new(big.Int).SetBytes(ParseStr(fdata.Sig.SIGS))
	sig := Signature{
		R: r,
		S: s,
	}

	if !verifySig(fileData, sig, pubkey) {
		t.Errorf("Verify Signature Fail\n")
	}
}

// 检查测试用json文件，若不存在则创建
func createFile() error {
	if _, err := os.Stat(absolutePath); errors.Is(err, os.ErrNotExist) {
		source, err := os.Open("../test_data.json")
		if err != nil {
			return err
		}
		defer source.Close()
		destination, err := os.Create(absolutePath)
		if err != nil {
			return err
		}
		defer destination.Close()
		_, err = io.Copy(destination, source)
		if err != nil {
			return err
		}
	}
	return nil
}