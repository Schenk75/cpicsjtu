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
    file_hash: String,
    file_data: String,
    file_name: String,
    time: i32,
    ec: EasyCodec,
}

impl Fact {
    fn new_fact(file_hash: String, file_data: String, file_name: String, time: i32) -> Fact {
        let mut ec = EasyCodec::new();
        ec.add_string("file_hash", file_hash.as_str());
        ec.add_string("file_data", file_data.as_str());
        ec.add_string("file_name", file_name.as_str());
        ec.add_i32("time", time);
        Fact {
            file_hash,
            file_data,
            file_name,
            time,
            ec,
        }
    }

    fn get_emit_event_data(&self) -> Vec<String> {
        let mut arr: Vec<String> = Vec::new();
        arr.push(self.file_hash.clone());
        arr.push(self.file_data.clone());
        arr.push(self.file_name.clone());
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
            file_hash: ec.get_string("file_hash").unwrap(),
            file_data: ec.get_string("file_data").unwrap(),
            file_name: ec.get_string("file_name").unwrap(),
            time: ec.get_i32("time").unwrap(),
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
    let file_hash = ctx.arg_as_utf8_str("file_hash");
    let file_data = ctx.arg_as_utf8_str("file_data");
    let file_name = ctx.arg_as_utf8_str("file_name");
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
    let fact = Fact::new_fact(file_hash, file_data, file_name, time);

    // 事件
    ctx.emit_event("topic_vx", &fact.get_emit_event_data());

    // 序列化后存储
    ctx.put_state(
        "fact_ec",
        fact.file_name.as_str(),
        fact.marshal().as_slice(),
    );
}

// find_by_file_hash 根据file_hash查询存证数据
#[no_mangle]
pub extern "C" fn find_by_file_hash() {
    // 获取上下文
    let ctx = &mut sim_context::get_sim_context();

    // 获取传入参数
    let file_hash = ctx.arg_as_utf8_str("file_hash");

    // 校验参数
    if file_hash.len() == 0 {
        ctx.log("file_hash is null");
        ctx.ok("".as_bytes());
        return;
    }

    // 查询
    let r = ctx.get_state("fact_ec", &file_hash);

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
    let r = ctx.get_state("fact_ec", &file_hash).unwrap();
    let fact = Fact::unmarshal(&r);
    let json_str = fact.to_json();

    // 返回查询结果
    ctx.ok(json_str.as_bytes());
    ctx.log(&json_str);
}

// find_by_file_name 根据file_name查询存证数据
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
    let json_str = fact.to_json();

    // 返回查询结果
    ctx.ok(json_str.as_bytes());
    ctx.log(&json_str);
}