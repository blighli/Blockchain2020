# 简易区块链原型——python实现

本项目用python实现了一个简易的区块链节点，对外暴露HTTP接口。节点使用PoW算法进行共识，使用最长链原则解决分歧。

## 节点的主要功能：
- 添加节点
- 查看节点列表
- 创建交易
- 挖矿（打包区块）
- 查看本地链
- 共识，冲突解决

## 测试
### 环境准备
- python3.5+
- flask
- requests

### 启动
打开终端，进入本路径，```python blockchain.py -p 5000```启动节点，其中5000是端口号，可以任意更改。启动多个节点更换端口即可。
启动成功后，直接使用curl或者postman调用接口即可。

### 接口
#### 添加节点
- POST 
- /nodes/register
- 参数
  ``` json
  {
    "nodes": [""]
  }
  ```
- curl示例
  ```shell
  curl --location --request POST 'http://localhost:5000/nodes/register' \
  --header 'Content-Type: application/json' \
  --data '{
    "nodes": [
    "localhost:5000",
    "localhost:5001",
    "localhost:5002",
    "localhost:5003"
    ]
  }'
  ```
#### 查看节点列表
- GET
- /nodes/list
- curl示例
  ```shell
  curl --location --request GET 'http://localhost:5000/nodes/list'
  ```
#### 创建交易
- POST
- /transactions/new
- 参数
  ```json
  {
    "sender": "",
    "recipient": "",
    "amount": 0
  }
  ```
- curl示例
  ```shell
  curl --location --request POST 'http://localhost:5000/transactions/new' \
  --header 'Content-Type: application/json' \
  --data '{
    "sender": "1a96bdeff5ff4e0fb94081540d10614b",
    "recipient": "b2b70489dc57484c978f01e863cab43f",
    "amount": 1
  }'
  ```
#### 挖矿
- GET
- /mine
- curl示例
  ```shell
  curl --location --request GET 'http://localhost:5000/mine'
  ```
#### 查看本地链
- GET
- /chain
- curl示例
  ```shell
  curl --location --request GET 'http://localhost:5000/chain'
  ```
#### 冲突解决
- GET
- /nodes/resolve
- curl示例
  ```shell
  curl --location --request GET 'http://localhost:5000/nodes/resolve'
  ```