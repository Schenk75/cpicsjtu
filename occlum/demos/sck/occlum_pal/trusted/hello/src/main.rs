use std::env;

fn main() {
    let args: Vec<String> = env::args().collect();
    for (index, value) in args.iter().enumerate() {
        println!("{} => {}", index, value);
    }
    println!("Hello, world!");
}
