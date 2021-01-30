#include "main.h"

void hstob(void *destination, const void *source)
{
    BYTE *dst = (BYTE *)destination;
    const BYTE *src = (BYTE *)source;

    if (!strncmp("0x", src, strlen("0x")))
        src += 2;

    const BYTE *end = src + strlen(src) - 1;
    *dst = 0;

    while (end >= src) {
        if ((end - src) % 2) {
            *dst = 0;
            if ((*end >= '0') && (*end <= '9')) {
                *dst += *end - '0';
            } else if ((*end >= 'a') && (*end <= 'f')) {
                *dst += *end - 'a' + 10;
            } else if ((*end >= 'A') && (*end <= 'A')) {
                *dst += *end - 'A' + 10;
            }
        } else {
            if ((*end >= '0') && (*end <= '9')) {
                *dst += (*end - '0') * 0x10;
            } else if ((*end >= 'a') && (*end <= 'f')) {
                *dst += (*end - 'a' + 10) * 0x10;
            } else if ((*end >= 'A') && (*end <= 'A')) {
                *dst += (*end - 'A' + 10) * 0x10;
            }
            dst += 1;
        }
        end -= 1;
    }
}

void btohs(void *destination, const void *source, int len)
{
    BYTE *dst = destination;
    const BYTE *src = (BYTE *)source;
    const BYTE *end = src + len - 1;

    while (end >= src) {
        int c = *end & 0xff;
        snprintf(dst, 3, (c < 0x10) ? "0%x" : "%x", c);
        end -= 1;
        dst += 2;
    }
}

void sha256_hash(BYTE *dst, const BYTE *src, int len)
{
    BYTE buf[SHA256_BLOCK_SIZE];

    SHA256_CTX _ctx;
    SHA256_CTX *ctx = &_ctx;

    sha256_init(ctx);
    sha256_update(ctx, src, len);
    sha256_final(ctx, buf);

    sha256_init(ctx);
    sha256_update(ctx, (BYTE *)&buf, SHA256_BLOCK_SIZE);
    sha256_final(ctx, dst);
}

void print_block(const Block *block)
{
    char buf[512];

    printf("Version       : 0x%08x\n", block->Version);

    btohs(buf, &(block->hashPrevBlock), SHA256_BLOCK_SIZE);
    printf("hashPrevBlock : 0x%s\n", buf);

    btohs(buf, &(block->hashMerkleRoot), SHA256_BLOCK_SIZE);
    printf("hashMerkleRoot: 0x%s\n", buf);

    printf("Time          : 0x%08x\n", block->Time);
    printf("Bits          : 0x%08x\n", block->Bits);
    printf("Nonce         : 0x%08x\n", block->Nonce);

    btohs(buf, block->hash, SHA256_BLOCK_SIZE);
    printf("Hash          : 0x%s\n", buf);

    printf("\n");
}

int hex256cmp(BYTE *x1, BYTE *x2)
{
    int len = SHA256_BLOCK_SIZE / sizeof(int);

    unsigned int *ix1 = (unsigned int *)x1;
    unsigned int *ix2 = (unsigned int *)x2;

    while (--len >= 0) {
        if (*(ix1 + len) < *(ix2 + len)) {
            return -1;
        } else if (*(ix1 + len) > *(ix2 + len)) {
            return 1;
        }
    }
    return 0;
}
