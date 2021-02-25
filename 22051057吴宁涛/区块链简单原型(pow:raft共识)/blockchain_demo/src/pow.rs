use crate::base::Block;
use num_bigint::{BigUint, BigInt};
use num_traits::{Zero, One};
use std::ops::Shl;
use std::mem;
use std::intrinsics::transmute;
use tiny_keccak::{Sha3, Hasher};
use std::cmp::Ordering;
use chrono::format::Numeric::Ordinal;

const TARGET_BITS: u32 = 24;
const MAX_NONCE: i64 = i64::max_value();

//pow工作量证明共识

pub struct ProofOfWork {
    pub block: Block,
    pub target: BigUint,
}

//实现工作量证明
impl ProofOfWork {
    //新建新的共识
    pub fn new_proof_of_work(&mut self) {
        let m: BigUint = One::one();
        let size: usize = (256 - TARGET_BITS) as usize;
        let t: BigUint = m.shl(size);
        self.target = t;
    }

    //准备数据
    pub fn prepare_data(&self, nonce: i64) -> Vec<u8> {
        let mut data: Vec<u8> = Vec::<u8>::new();
        for v in self.block.prev_block_hash.iter() {
            data.push(*v);
        }
        for v in self.block.data.as_bytes().iter() {
            data.push(*v);
        }
        for v in self.block.timestamp.to_string().as_bytes().iter() {
            data.push(*v);
        }
        unsafe {
            let n: [u8; 4] = transmute::<u32, [u8; 4]>(TARGET_BITS);
            for v in n.iter() {
                data.push(*v);
            }
            let m: [u8; 8] = transmute::<i64, [u8; 8]>(nonce);
            for v in m.iter() {
                data.push(*v);
            }
        }
        data
    }

    //运行共识
    pub fn run(&self) -> (i64, [u8; 32]) {
        let mut big_int: BigUint;
        let mut hash: [u8; 32] = [0; 32];
        let mut nonce: i64 = 0;

        loop {
            if nonce >= MAX_NONCE {
                break;
            }
            let pd = self.prepare_data(nonce);
            let data: &[u8] = pd.as_ref();
            let mut sha3 = Sha3::v256();
            sha3.update(data);
            sha3.finalize(&mut hash);
            big_int = BigUint::from_bytes_be(&hash);
            if big_int.cmp(&self.target) == Ordering::Greater {
                break;
            } else {
                nonce = nonce + 1;
            }
        }

        (nonce, hash)
    }

    //验证
    pub fn validate(&self) -> bool {
        let pd: Vec<u8> = self.prepare_data(self.block.nonce);
        let data: &[u8] = pd.as_ref();
        let mut hash: [u8; 32] = [0; 32];
        let mut sha3 = Sha3::v256();
        sha3.update(data);
        sha3.finalize(&mut hash);
        let big_int: BigUint = BigUint::from_bytes_be(&hash);
        let is_valid: bool = big_int.cmp(&self.target) == Ordering::Greater;
        is_valid
    }
}