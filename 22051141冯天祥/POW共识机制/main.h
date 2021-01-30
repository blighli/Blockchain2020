#ifndef _MAIN_H_
#define _MAIN_H_

#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>
#include <fcntl.h>
#include <string.h>
#include <malloc.h>
#include "sha256.h"

typedef struct _Block {
    WORD Version; /**< 区块版本号 */
    BYTE hashPrevBlock[SHA256_BLOCK_SIZE]; /**< 前一块区块的256-bit哈希值 */
    BYTE hashMerkleRoot[SHA256_BLOCK_SIZE]; /**< 所有区块的哈希值的256-bit哈希值 */
    WORD Time; /**< 从 1970-01-01T00:00 UTC 起始的当前交易时间(s)*/
    WORD Bits; /**< Current target in compact format */
    WORD Nonce; /**< 32-bit  */
    BYTE hash[SHA256_BLOCK_SIZE]; /**< 当前区块256-bit哈希值 */
} Block;

/**
 * @brief 16进制转为2进制
 *
 * @param[out] destination memory address to be written
 * @param[in] source memory address to be read (ends with '\0')
 */
void hstob(void *destination, const void *source);

/**
 * @brief 2进制转为16进制
 *
 * @param[out] destination memory address to be written
 * @param[in] source memory address to be read
 * @param[in] len size of memory to be copied
 *
 * @return string encoded in hex
 */
void btohs(void *destination, const void *source, int len);

/**
 * @brief 二进制数据转为sha256
 *
 * @param[out] destination memory address to be written
 * @param[in] source memory address to be read
 * @param[in] len size of memory to be copied
 */
void sha256_hash(BYTE *destination, const BYTE *source, int len);

/**
 * @brief 打印块信息
 *
 * @param[in] block Given block
 */
void print_block(const Block *block);

/**
 * @brief 比较二进制信息
 *
 * @param[in] x1 256-bit binary
 * @param[in] x2 256-bit binary
 *
 * @return (x1 < x2) return -1, (x1 > x2) return 1, (x1 == x2) return 0
 */
int hex256cmp(BYTE *x1, BYTE *x2);

#endif
