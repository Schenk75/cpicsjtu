## 环境依赖
使用rust的sdk编译合约依赖rust编译器和cargo

## 安装rust cargo

> 官方地址: https://www.rust-lang.org/tools/install

**windows**

>  To install Rust on Windows, download and run [rustup-init.exe](https://win.rustup.rs/), then follow the onscreen instructions.
>
> https://win.rustup.rs/

**Linux、Mac**

> curl https://sh.rustup.rs -sSf | sh

## 编译
```sh
make build
```
生成路径：target/wasm32-unknown-unknown/release/chainmaker_contract.wasm
