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
    pubkey: String,
    pubkey_hash: String,
    orgid: String,
    time: i32,
    ec: EasyCodec,
}

#[allow(dead_code)]
impl Fact {
    fn new_fact(pubkey: String, pubkey_hash: String, orgid: String, time: i32) -> Fact {
        let mut ec = EasyCodec::new();
        ec.add_string("pubkey", pubkey.as_str());
        ec.add_string("pubkey_hash", pubkey.as_str());
        ec.add_string("orgid", orgid.as_str());
        ec.add_i32("time", time);
        Fact {
            pubkey,
            pubkey_hash,
            orgid,
            time,
            ec,
        }
    }

    fn get_emit_event_data(&self) -> Vec<String> {
        let mut arr: Vec<String> = Vec::new();
        arr.push(self.pubkey.clone());
        arr.push(self.pubkey_hash.clone());
        arr.push(self.orgid.clone());
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
            pubkey:         ec.get_string("pubkey").unwrap(),
            pubkey_hash:    ec.get_string("pubkey_hash").unwrap(),
            orgid:          ec.get_string("orgid").unwrap(),
            time:           ec.get_i32("time").unwrap(),
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
    let pubkey = ctx.arg_as_utf8_str("pubkey");
    let pubkey_hash = ctx.arg_as_utf8_str("pubkey_hash");
    let orgid = ctx.arg_as_utf8_str("orgid");
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
    let fact = Fact::new_fact(pubkey, pubkey_hash, orgid, time);

    // 事件
    ctx.emit_event("topic_vx", &fact.get_emit_event_data());

    // 序列化后存储
    ctx.put_state(
        "fact_ec",
        fact.pubkey_hash.as_str(),
        fact.marshal().as_slice(),
    );
}

// 根据pubkey_hash查询该公钥是否已经在链上注册
#[no_mangle]
pub extern "C" fn find_by_pubkey_hash() {
    // 获取上下文
    let ctx = &mut sim_context::get_sim_context();

    // 获取传入参数
    let pubkey_hash = ctx.arg_as_utf8_str("pubkey_hash");

    // 校验参数
    if pubkey_hash.len() == 0 {
        ctx.log("pubkey hash is null");
        ctx.ok("".as_bytes());
        return;
    }

    // 查询
    let r = ctx.get_state("fact_ec", &pubkey_hash);

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
    let json_str = fact.to_json();

    // 返回查询结果
    ctx.ok(json_str.as_bytes());
    ctx.log(&json_str);
}