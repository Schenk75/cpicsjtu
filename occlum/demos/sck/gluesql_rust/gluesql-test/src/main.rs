extern crate gluesql;
extern crate serde;
extern crate serde_json;

use gluesql::prelude::*;
use serde::{Deserialize, Serialize};
use serde_json::from_str;
use std::time::{Duration, Instant};

const DATAFILE: &str = "/bin/data-FRqwI11L_LlTU-GWNV0sz.json";

#[derive(Debug, Serialize, Deserialize)]
#[allow(non_snake_case)]
struct EnvironmentData {
    DATE: String,
    TIME: String,
    air_humidity: f32,
    air_temperature: f32,
    atmosphere: f32,
    co: f32,
    no2: f32,
    o3: f32,
    pm10: f32,
    pm25: f32,
    rainfall: f32,
    so2: f32,
    soil_ec: f32,
    soil_humidity: f32,
    soil_ph: f32,
    soil_temperature: f32,
    wind_direction: f32,
    wind_speed: f32,
}

fn speedtest_memory() {
    let storage = MemoryStorage::default();
    let mut glue = Glue::new(storage);
    let sql_stmt = "CREATE TABLE myTable ( \
                        id integer, \
                        DATE text, \
                        TIME text, \
                        air_humidity float, \
                        air_temperature float, \
                        atmosphere float, \
                        co float, \
                        no2 float, \
                        o3 float, \
                        pm10 float, \
                        pm25 float, \
                        rainfall float, \
                        so2 float, \
                        soil_ec float, \
                        soil_humidity float, \
                        soil_ph float, \
                        soil_temperature float, \
                        wind_direction float, \
                        wind_speed float);";
    let _res = glue.execute(sql_stmt);
    // println!("{:?}", _res);

    // Insert 500
    let raw_data = std::fs::read_to_string(DATAFILE).unwrap();
    let vals: Vec<EnvironmentData> = from_str(&raw_data).unwrap();
    // println!("{:?}", vals);
    let mut time = Duration::new(0, 0);
    for (idx, val) in vals.into_iter().enumerate() {
        let sql = format!(
            "INSERT INTO myTable VALUES ({}, \"{}\", \"{}\", {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {})",
            idx, val.DATE, val.TIME, val.air_humidity, val.air_temperature,
            val.atmosphere, val.co, val.no2, val.o3, val.pm10, val.pm25, val.rainfall,
            val.so2, val.soil_ec, val.soil_humidity, val.soil_ph, val.soil_temperature,
            val.wind_direction, val.wind_speed
        );
        // println!("{}", &sql);
        let start = Instant::now();
        let _res = glue.execute(&sql).unwrap();
        time += start.elapsed();
        // println!("{:?}", _res);
    }
    println!("Insert time: {:?}", time);

    // Aggregate
    let sql = "SELECT avg(rainfall) FROM myTable;";
    let start = Instant::now();
    let _res = glue.execute(&sql).unwrap();
    println!("Avg time: {:?}", start.elapsed());

    let sql = "SELECT min(rainfall) FROM myTable;";
    let start = Instant::now();
    let _res = glue.execute(&sql).unwrap();
    println!("Min time: {:?}", start.elapsed());

    let sql = "SELECT max(rainfall) FROM myTable;";
    let start = Instant::now();
    let _res = glue.execute(&sql).unwrap();
    println!("Max time: {:?}", start.elapsed());

    // Case
    let sql = "SELECT *, \
                CASE WHEN air_temperature > 0 \
                THEN 'Positive' \
                ELSE 'Negative' \
                END AS T_status \
                FROM myTable;";
    let start = Instant::now();
    let _res = glue.execute(&sql).unwrap();
    println!("Case time: {:?}", start.elapsed());

    // Between
    let sql = "SELECT * FROM myTable WHERE air_temperature BETWEEN 4 AND 5;";
    let start = Instant::now();
    let _res = glue.execute(&sql).unwrap();
    println!("Between time: {:?}", start.elapsed());
}

fn speedtest_sled() {
    let sled_storage = "/bin/sled";

    let storage = SledStorage::new(sled_storage).unwrap();
    let mut glue = Glue::new(storage);
    let sql_stmt = "DROP TABLE IF EXISTS myTable;";
    let _res = glue.execute(sql_stmt);

    let sql_stmt = "CREATE TABLE myTable ( \
                        id integer, \
                        DATE text, \
                        TIME text, \
                        air_humidity float, \
                        air_temperature float, \
                        atmosphere float, \
                        co float, \
                        no2 float, \
                        o3 float, \
                        pm10 float, \
                        pm25 float, \
                        rainfall float, \
                        so2 float, \
                        soil_ec float, \
                        soil_humidity float, \
                        soil_ph float, \
                        soil_temperature float, \
                        wind_direction float, \
                        wind_speed float);";
    let _res = glue.execute(sql_stmt);
    // println!("{:?}", _res);

    // Insert 500
    let raw_data = std::fs::read_to_string(DATAFILE).unwrap();
    let vals: Vec<EnvironmentData> = from_str(&raw_data).unwrap();
    // println!("{:?}", vals);
    let mut time = Duration::new(0, 0);
    for (idx, val) in vals.into_iter().enumerate() {
        let sql = format!(
        "INSERT INTO myTable VALUES ({}, \"{}\", \"{}\", {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {})",
        idx, val.DATE, val.TIME, val.air_humidity, val.air_temperature,
        val.atmosphere, val.co, val.no2, val.o3, val.pm10, val.pm25, val.rainfall,
        val.so2, val.soil_ec, val.soil_humidity, val.soil_ph, val.soil_temperature,
        val.wind_direction, val.wind_speed
    );
        // println!("{}", &sql);
        let start = Instant::now();
        let _res = glue.execute(&sql).unwrap();
        time += start.elapsed();
        // println!("{:?}", _res);
    }
    println!("Insert time: {:?}", time);

    // Aggregate
    let sql = "SELECT avg(rainfall) FROM myTable;";
    let start = Instant::now();
    let _res = glue.execute(&sql).unwrap();
    println!("Avg time: {:?}", start.elapsed());

    let sql = "SELECT min(rainfall) FROM myTable;";
    let start = Instant::now();
    let _res = glue.execute(&sql).unwrap();
    println!("Min time: {:?}", start.elapsed());

    let sql = "SELECT max(rainfall) FROM myTable;";
    let start = Instant::now();
    let _res = glue.execute(&sql).unwrap();
    println!("Max time: {:?}", start.elapsed());

    // Case
    let sql = "SELECT *, \
                CASE WHEN air_temperature > 0 \
                THEN 'Positive' \
                ELSE 'Negative' \
                END AS T_status \
                FROM myTable;";
    let start = Instant::now();
    let _res = glue.execute(&sql).unwrap();
    println!("Case time: {:?}", start.elapsed());

    // Between
    let sql = "SELECT * FROM myTable WHERE air_temperature BETWEEN 4 AND 5;";
    let start = Instant::now();
    let _res = glue.execute(&sql).unwrap();
    println!("Between time: {:?}", start.elapsed());
}

fn main() {
    println!("[+] Gluesql Memory");
    speedtest_memory();

    // println!("\n[+] Gluesql Sled");
    // speedtest_sled();
}
