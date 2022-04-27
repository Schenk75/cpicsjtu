package main

import (
	"errors"
	"io"
	"os"
	"testing"
)

// 公钥注册流程测试
func TestUploadKey(t *testing.T) {
	absolutePath := "/root/jwzhou/paho.mqtt.c/occlum_instance/pb_json/test_test.json"
	if _, err := os.Stat(absolutePath); errors.Is(err, os.ErrNotExist) {
		source, err := os.Open("../test_pk.json")
		if err != nil {
			t.Errorf("%v", err)
		}
		defer source.Close()
		destination, err := os.Create(absolutePath)
		if err != nil {
			t.Errorf("%v", err)
		}
		defer destination.Close()
		_, err = io.Copy(destination, source)
		if err != nil {
			t.Errorf("%v", err)
		}
	}
	res := uploadPk("test_test.json")
	if !res {
		t.Errorf("Upload Public Key Fail")
	}
}

// 测试ParseStr函数
func TestParse(t *testing.T) {
	str := "0xE6,0xCD,0x9C,0xDE,0x21,0xC2,0x19,0x35,0x7B,0xF7,0xC8,0xE2,0x6B,0xE4,0x3F,0x04,0xE5,0x96,0x2E,0xF5,0xB3,0x07,0xB0,0x5C,0x0C,0x4D,0xEE,0xBC,0x2A,0xEE,0x96,0xBA"
	expected := []byte{230, 205, 156, 222, 33, 194, 25, 53, 123, 247, 200, 226, 107, 228, 63, 4, 229, 150, 46, 245, 179, 7, 176, 92, 12, 77, 238, 188, 42, 238, 150, 186}
	parse := ParseStr(str)
	for i := range expected {
		if expected[i] != parse[i] {
			t.Errorf("Parse result not right")
		}
	}
}