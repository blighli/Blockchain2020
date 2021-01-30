# Kitties

## 功能
基于Substrate实现了以太坊加密猫的功能，
具体功能如下：
- 创建一只加密猫
- 加密猫展示成一张卡片，并显示是不是属于你的 
- 转让加密猫
为了防止加密猫被无限地创建，在创建时需要质押一定数量的dot。

## 效果图
![UTOOLS_1612005289980.png](https://i.loli.net/2021/01/30/IOxpMoWYU5NLGcB.png)

## 运行方式
需要rust nightly-2020-10-29版本和前端yarn环境

### 编译
```shell 
cd frontend && yarn install 
cd node && cargo build --release 
```
### 运行
```shell
cd frontend && yarn install 
cd node && ./target/release/node-template --dev 
```
