/// 
/// Copyright (C) BABEC. All rights reserved.
/// 
/// SPDX-License-Identifier: Apache-2.0
/// 

use crate::easycodec::*;
use crate::sim_context;
use sim_context::*;

// 安装合约时会执行此方法，必须
#[no_mangle]
pub extern "C" fn init_contract() {
    // 安装时的业务逻辑，内容可为空
    sim_context::log("init_contract");
}

// 升级合约时会执行此方法，必须
#[no_mangle]
pub extern "C" fn upgrade() {
    // 升级时的业务逻辑，内容可为空
    sim_context::log("upgrade");
    let ctx = &mut sim_context::get_sim_context();
    ctx.ok("upgrade success".as_bytes());
}

struct Fact {
    file_sig: String,
    file_data: String,
    file_name: String,
    pubkey: String,
    time: i32,
    ec: EasyCodec,
}

#[allow(dead_code)]
struct Data {
    sig: String,
    data: String,
    pubkey: String,
    ec: EasyCodec,
}

#[allow(dead_code)]
impl Fact {
    fn new_fact(file_sig: String, file_data: String, file_name: String, pubkey: String, time: i32) -> Fact {
        let mut ec = EasyCodec::new();
        ec.add_string("file_sig", file_sig.as_str());
        ec.add_string("file_data", file_data.as_str());
        ec.add_string("file_name", file_name.as_str());
        ec.add_string("pubkey", pubkey.as_str());
        ec.add_i32("time", time);
        Fact {
            file_sig,
            file_data,
            file_name,
            pubkey,
            time,
            ec,
        }
    }

    fn get_emit_event_data(&self) -> Vec<String> {
        let mut arr: Vec<String> = Vec::new();
        arr.push(self.file_sig.clone());
        arr.push(self.file_data.clone());
        arr.push(self.file_name.clone());
        arr.push(self.pubkey.clone());
        arr.push(self.time.to_string());
        arr
    }

    fn to_json(&self) -> String {
        self.ec.to_json()
    }

    fn marshal(&self) -> Vec<u8> {
        self.ec.marshal()
    }

    fn unmarshal(data: &Vec<u8>) -> Fact {
        let ec = EasyCodec::new_with_bytes(data);
        Fact {
            file_sig: ec.get_string("file_sig").unwrap(),
            file_data: ec.get_string("file_data").unwrap(),
            file_name: ec.get_string("file_name").unwrap(),
            pubkey:    ec.get_string("pubkey").unwrap(),
            time: ec.get_i32("time").unwrap(),
            ec,
        }
    }
}

#[allow(dead_code)]
impl Data {
    fn new_data(sig: String, data: String, pubkey: String) -> Data {
        let mut ec = EasyCodec::new();
        ec.add_string("sig", sig.as_str());
        ec.add_string("data", data.as_str());
        ec.add_string("pubkey", pubkey.as_str());
        Data {
            sig,
            data,
            pubkey,
            ec,
        }
    }

    fn get_emit_event_data(&self) -> Vec<String> {
        let mut arr: Vec<String> = Vec::new();
        arr.push(self.sig.clone());
        arr.push(self.data.clone());
        arr.push(self.pubkey.clone());
        arr
    }

    fn to_json(&self) -> String {
        self.ec.to_json()
    }

    fn marshal(&self) -> Vec<u8> {
        self.ec.marshal()
    }

    fn unmarshal(data: &Vec<u8>) -> Data {
        let ec = EasyCodec::new_with_bytes(data);
        Data {
            sig: ec.get_string("sig").unwrap(),
            data: ec.get_string("data").unwrap(),
            pubkey: ec.get_string("pubkey").unwrap(),
            ec,
        }
    }
}

// save 保存存证数据
#[no_mangle]
pub extern "C" fn save() {
    // 获取上下文
    let ctx = &mut sim_context::get_sim_context();

    // 获取传入参数
    let file_sig = ctx.arg_as_utf8_str("file_sig");
    let file_data = ctx.arg_as_utf8_str("file_data");
    let file_name = ctx.arg_as_utf8_str("file_name");
    let pubkey = ctx.arg_as_utf8_str("pubkey");
    let time_str = ctx.arg_as_utf8_str("time");

    // 构造结构体
    let r_i32 = time_str.parse::<i32>();
    if r_i32.is_err() {
        let msg = format!("time is {:?} not int32 number.", time_str);
        ctx.log(&msg);
        ctx.error(&msg);
        return;
    }
    let time: i32 = r_i32.unwrap();
    let fact = Fact::new_fact(file_sig, file_data, file_name, pubkey, time);

    // 事件
    ctx.emit_event("topic_vx", &fact.get_emit_event_data());

    // 序列化后存储
    ctx.put_state(
        "fact_ec",
        fact.file_name.as_str(),
        fact.marshal().as_slice(),
    );
}

// 根据file_name查询存证数据
#[no_mangle]
pub extern "C" fn find_by_file_name() {
    // 获取上下文
    let ctx = &mut sim_context::get_sim_context();

    // 获取传入参数
    let file_name = ctx.arg_as_utf8_str("file_name");

    // 校验参数
    if file_name.len() == 0 {
        ctx.log("file_name is null");
        ctx.ok("".as_bytes());
        return;
    }

    // 查询
    let r = ctx.get_state("fact_ec", &file_name);

    // 校验返回结果
    if r.is_err() {
        ctx.log("get_state fail");
        ctx.error("get_state fail");
        return;
    }
    let fact_vec = r.unwrap();
    if fact_vec.len() == 0 {
        ctx.log("None");
        ctx.ok("".as_bytes());
        return;
    }

    // 查询
    let fact = Fact::unmarshal(&fact_vec);
    // let json_str = fact.to_json();
    let return_val = Data::new_data(fact.file_sig, fact.file_data, fact.pubkey);
    let json_str = return_val.to_json();

    // 返回查询结果
    ctx.ok(json_str.as_bytes());
    ctx.log(&json_str);
}