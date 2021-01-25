use crate::pow::ProofOfWork;
use tiny_keccak::{Sha3, Hasher};
use chrono::prelude::*;
use std::borrow::Borrow;
use num_bigint::BigUint;


extern crate chrono;

//区块链系统基本骨架

#[derive(Debug)]
pub struct Block {
    pub timestamp: i64,
    pub data: String,
    pub prev_block_hash: [u8; 32],
    pub hash: [u8; 32],
    pub nonce: i64,
}

impl Block {
    //设置哈希
    pub fn set_hash(block: &mut Block) -> &Block {
        let mut sha3 = Sha3::v256();
        sha3.update(block.timestamp.to_string().as_bytes());
        sha3.update(block.data.as_ref());
        sha3.update(block.prev_block_hash.as_ref());
        let mut hash: [u8; 32] = [0; 32];
        sha3.finalize(&mut hash);
        block.hash = hash;
        block
    }

    //创建新的区块
    pub fn new_block(data: String, prev_block_hash: [u8; 32]) -> Block {
        let dt = Local::now();
        let mut block = Block {
            timestamp: dt.timestamp_millis(),
            data,
            prev_block_hash,
            hash: [0;32],
            nonce: 0,
        };
        let mut pow = ProofOfWork {
            block: Block {
                timestamp: 0,
                data: "".to_string(),
                prev_block_hash: [0; 32],
                hash: [0; 32],
                nonce: 0,
            },
            target: BigUint::new(Vec::new()),
        };
        pow.new_proof_of_work();
        let (nonce, hash) = pow.run();
        block.hash = hash;
        block.nonce = nonce;
        pow.block.timestamp = block.timestamp;
        pow.block.data = block.data;
        pow.block.prev_block_hash = block.prev_block_hash;
        pow.block.hash = block.hash;
        pow.block.nonce = block.nonce;
        pow.block
    }

    pub fn new_genesis_block() -> Block {
        return Block::new_block("Genesis Block".to_string(), [0; 32]);
    }
}

pub struct Blockchain {
    pub blocks: Vec<Block>,
}

impl Blockchain {
    //添加区块
    pub fn add_block(&mut self, data: String) {
        let mut prev_block= self.blocks.get(self.blocks.len()-1);
        let block;
        if prev_block.is_none() {
            block = Block::new_block(data, [0; 32]);
        } else {
            block = Block::new_block(data, prev_block.unwrap().hash);
        }
        self.blocks.push(block);
    }

    //创建新的链
    pub fn new_blockchain(&mut self) {
        let block = Block::new_genesis_block();
        self.blocks.push(block);
    }
}