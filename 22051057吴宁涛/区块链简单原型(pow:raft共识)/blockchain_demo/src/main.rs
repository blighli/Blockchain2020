pub mod base;
pub mod pow;

use std::thread;
use std::time::Duration;
use num_bigint::BigUint;
use crate::pow::ProofOfWork;
use crate::base::{Block, Blockchain};

//主函数测试

fn main() {
    let mut bc: Blockchain = Blockchain {
        blocks: Vec::<Block>::new(),
    };
    bc.new_blockchain();
    //睡两秒并模拟发生交易
    thread::sleep(Duration::from_secs(2));
    bc.add_block("Send 1 coin to TateBrown".to_string());
    thread::sleep(Duration::from_secs(2));
    bc.add_block("Send 2 more coin to TateBrown".to_string());

    for mut block in bc.blocks {
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
        //生成新区块并共识
        pow.new_proof_of_work();
        let (nonce, hash) = pow.run();
        block.hash = hash;
        block.nonce = nonce;
        pow.block.timestamp = block.timestamp;
        pow.block.data = block.data;
        pow.block.prev_block_hash = block.prev_block_hash;
        pow.block.hash = block.hash;
        pow.block.nonce = block.nonce;
        let is_valid: bool = pow.validate();
        println!("{:?}", pow.block);
        println!("{}", is_valid);
    }
}
