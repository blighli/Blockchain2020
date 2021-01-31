import hashlib
import json
from time import time
from urllib.parse import urlparse
from uuid import uuid4

import requests
from flask import Flask, jsonify, request


class Blockchain:
    def __init__(self):
        self.current_transactions = [{'sender': "0", 'recipient': node_identifier, 'amount': 100}]
        self.chain = []
        self.nodes = set()

        # 创世区块
        self.new_block(previous_hash='1', proof=100)

    def register_node(self, address):
        """
        添加节点

        :param address: 节点地址，如'http://127.0.0.1:5000'
        """

        parsed_url = urlparse(address)
        if parsed_url.netloc:
            self.nodes.add(parsed_url.netloc)
        elif parsed_url.path:
            # ip:port形式的输入 '127.0.0.1:5000'.
            self.nodes.add(parsed_url.path)
        else:
            raise ValueError('无效的URL')

    def valid_chain(self, chain):
        """
        验证一个链是否合法

        :param chain: 链
        :return: 合法返回True，否则返回False
        """

        last_block = chain[0]
        current_index = 1

        while current_index < len(chain):
            block = chain[current_index]
            print(f'{last_block}')
            print(f'{block}')
            print("\n-----------\n")
            # 校验last_block的哈希值
            last_block_hash = self.hash(last_block)
            if block['previous_hash'] != last_block_hash:
                return False

            # 校验PoW的proof
            if not self.valid_proof(last_block['proof'], block['proof'], last_block_hash):
                return False

            last_block = block
            current_index += 1

        return True

    def resolve_conflicts(self):
        """
        共识算法，保留最长链

        :return: 如果本地链被替换返回True，否则返回False
        """

        neighbours = self.nodes
        new_chain = None

        max_length = len(self.chain)

        for node in neighbours:
            response = requests.get(f'http://{node}/chain')

            if response.status_code == 200:
                length = response.json()['length']
                chain = response.json()['chain']

                if length > max_length and self.valid_chain(chain):
                    max_length = length
                    new_chain = chain

        if new_chain:
            self.chain = new_chain
            return True

        return False

    def new_block(self, proof, previous_hash):
        """
        创建区块

        :param proof: PoW求出的proof
        :param previous_hash: 前一区块的哈希
        :return: 新区块
        """

        block = {
            'index': len(self.chain) + 1,
            'timestamp': time(),
            'transactions': self.current_transactions,
            'proof': proof,
            'previous_hash': previous_hash or self.hash(self.chain[-1]),
        }

        self.current_transactions = []

        self.chain.append(block)
        return block

    def new_transaction(self, sender, recipient, amount):
        """
        创建交易

        :param sender: 发送者地址
        :param recipient: 接受者地址
        :param amount: 金额
        :return: 交易应该被放入的区块的索引
        """
        self.current_transactions.append({
            'sender': sender,
            'recipient': recipient,
            'amount': amount,
        })

        return self.last_block['index'] + 1

    @property
    def last_block(self):
        return self.chain[-1]

    @staticmethod
    def hash(block):
        """
        SHA-256求区块哈希

        :param block: 区块
        """

        # 字段按字典序排列，保证哈希值确定
        block_string = json.dumps(block, sort_keys=True).encode()
        return hashlib.sha256(block_string).hexdigest()

    def proof_of_work(self, last_block):
        """
        简易PoW：
         - 求一个p'使得hash(pp')前4位为0
         - 其中pui是前一块的proof，p'是新的proof

        :param last_block: 最后一个区块
        :return: <int> 求得的proof
        """

        last_proof = last_block['proof']
        last_hash = self.hash(last_block)

        proof = 0
        while self.valid_proof(last_proof, proof, last_hash) is False:
            proof += 1

        return proof

    @staticmethod
    def valid_proof(last_proof, proof, last_hash):
        """
        校验proof

        :param last_proof: <int> 最后一块的proof
        :param proof: <int> 当前得proof
        :param last_hash: <str> 最后一块的哈希
        :return: <bool> 符合要求返回True，否则返回False
        """

        guess = f'{last_proof}{proof}{last_hash}'.encode()
        guess_hash = hashlib.sha256(guess).hexdigest()
        return guess_hash[:4] == "0000"


app = Flask(__name__)

node_identifier = str(uuid4()).replace('-', '')

blockchain = Blockchain()


@app.route('/mine', methods=['GET'])
def mine():
    # 用PoW求出一个proof
    last_block = blockchain.last_block
    proof = blockchain.proof_of_work(last_block)

    # 挖矿成功奖励一个币
    # sender是0，意味着这个币是新挖出来的
    blockchain.new_transaction(
        sender="0",
        recipient=node_identifier,
        amount=1,
    )

    previous_hash = blockchain.hash(last_block)
    block = blockchain.new_block(proof, previous_hash)

    response = {
        'message': "新块创建成功",
        'index': block['index'],
        'transactions': block['transactions'],
        'proof': block['proof'],
        'previous_hash': block['previous_hash'],
    }
    return jsonify(response), 200


@app.route('/transactions/new', methods=['POST'])
def new_transaction():
    values = request.get_json()

    # 校验必填字段
    required = ['sender', 'recipient', 'amount']
    if not all(k in values for k in required):
        return '参数不足', 400

    # 创建交易
    index = blockchain.new_transaction(values['sender'], values['recipient'], values['amount'])

    response = {'message': f'交易会被打包进区块{index}'}
    return jsonify(response), 201


@app.route('/chain', methods=['GET'])
def full_chain():
    response = {
        'chain': blockchain.chain,
        'length': len(blockchain.chain),
    }
    return jsonify(response), 200


@app.route('/nodes/register', methods=['POST'])
def register_nodes():
    values = request.get_json()

    nodes = values.get('nodes')
    if nodes is None:
        return "请提供有效的节点列表", 400

    for node in nodes:
        blockchain.register_node(node)

    response = {
        'message': '新节点添加成功',
        'total_nodes': list(blockchain.nodes),
    }
    return jsonify(response), 201


@app.route('/nodes/list', methods=['GET'])
def get_nodes():
    response = {
        "nodes": list(blockchain.nodes)
    }
    return jsonify(response), 200


@app.route('/nodes/resolve', methods=['GET'])
def consensus():
    replaced = blockchain.resolve_conflicts()

    if replaced:
        response = {
            'message': '本地链已被替换',
            'new_chain': blockchain.chain
        }
    else:
        response = {
            'message': '本地链有效',
            'chain': blockchain.chain
        }

    return jsonify(response), 200


if __name__ == '__main__':
    from argparse import ArgumentParser

    parser = ArgumentParser()
    parser.add_argument('-p', '--port', default=5000, type=int, help='port to listen on')
    args = parser.parse_args()
    port = args.port

    app.run(host='0.0.0.0', port=port)
