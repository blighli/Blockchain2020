#include "main.h"
#include <time.h>
#include <glib-2.0/gmodule.h>
#include <arpa/inet.h>
#include <sys/types.h>
#include <sys/socket.h>
#include <sys/wait.h>

static unsigned int BITS = 0x1f00ffff;//选择挖矿难度
static char *node;

void block_init(Block *block,
                WORD version,
                const char *hashPrevBlock,
                const char *hashMerkleRoot,
                WORD time,
                WORD bits,
                WORD nonce)
{
    if (!block)
        return;

    char *empty = "0000000000000000000000000000000000000000000000000000000000000000";

    block->Version = version;
    block->Time = time;
    block->Bits = bits;
    block->Nonce = nonce;

    hstob(&(block->hashPrevBlock), (hashPrevBlock) ? hashPrevBlock : empty);
    hstob(&(block->hashMerkleRoot), (hashMerkleRoot) ? hashMerkleRoot : empty);
}

void set_target(BYTE *dst, WORD *bits)
{
    memset(dst, 0, SHA256_BLOCK_SIZE);

    BYTE *_bits = (BYTE *)bits + 3;
    if (*_bits > 0x20)
        return;

    BYTE *ptr = dst + *_bits;

    for (int i = 1; i <= 3; i++)
        *(ptr - i) = *(_bits - i);
}

void set_hash(Block *block)
{
    sha256_hash(block->hash, (BYTE *)block, sizeof(Block) - SHA256_BLOCK_SIZE);
}

const Block *genesis_init(void)
{
    Block *block = calloc(1, sizeof(Block));

    // Bitcoin genesis block
    block_init(block, 1,
               NULL,
               "4a5e1e4baab89f3a32518a88c31bc87f618f76673e2cc77ab2127b7afdeda33b",
               1231006505,
               0x1d00ffff,
               2083236893);

    BYTE target[SHA256_BLOCK_SIZE];

    set_target(target, &(block->Bits));
    set_hash(block);

    while (hex256cmp((BYTE *)block->hash, target) >= 0)
    {
        block->Nonce += 1;
        set_hash(block);
    }

    return block;
}

void mine(const Block *genesis, int baseNonce, int step, int sockfd)
{
    BYTE target[SHA256_BLOCK_SIZE];
    char hashPrevBlock[64];
    Block buffer;
    int len;

    GList *chain = g_list_append(NULL, (gpointer)genesis);

    while (g_list_length(chain) < 10)
    {
        Block *block = calloc(1, sizeof(Block));
        int success = 0;

        btohs(hashPrevBlock, ((Block *)g_list_last(chain)->data)->hash, SHA256_BLOCK_SIZE);
        block_init(block, 1, hashPrevBlock, NULL, (unsigned int)time(NULL), BITS, baseNonce);

        set_target(target, &(block->Bits));
        set_hash(block);

        while (hex256cmp((BYTE *)block->hash, target) >= 0)
        {
            if ((len = read(sockfd, &buffer, sizeof(Block))) > 0)
            {
                free(block);
                block = &buffer;
                goto fail;
            }

            block->Nonce += step;
            set_hash(block);
        }

        success = 1;
        write(sockfd, block, sizeof(Block));

    fail:
        chain = g_list_append(chain, (gpointer)block);

        if (success)
        {
            printf("[%02d] %s\n", g_list_length(chain), node);
            print_block(block);
        }
    }

    close(sockfd);
}

void error(char *message)
{
    fputs(message, stderr);
    fputc('\n', stderr);
    exit(1);
}

int main()
{
    const Block *genesis;

    genesis = genesis_init();
    printf("[01]\n");
    print_block(genesis);

    int pid = fork();

    if (pid > 0)
    {
        node = "parent";

        int sockfd, newsockfd;
        struct sockaddr_in serv_addr;
        struct sockaddr_in clnt_addr;
        int clnt_addr_size;

        sockfd = socket(PF_INET, SOCK_STREAM, 0);
        if (sockfd == -1)
            error("socket() error");

        memset(&serv_addr, 0, sizeof(serv_addr));
        serv_addr.sin_family = AF_INET;
        serv_addr.sin_addr.s_addr = htonl(INADDR_ANY);
        serv_addr.sin_port = htons(9000);

        if (bind(sockfd, (struct sockaddr *)&serv_addr, sizeof(serv_addr)) == -1)
            error("bind() error");

        if (listen(sockfd, 5) == -1)
            error("listen() error");

        clnt_addr_size = sizeof(clnt_addr);
        newsockfd = accept(sockfd, (struct sockaddr *)&clnt_addr, &clnt_addr_size);
        if (newsockfd == -1)
            error("accept() error");

        int opt;
        opt = fcntl(newsockfd, F_GETFL, 0);
        fcntl(newsockfd, F_SETFL, opt | O_NONBLOCK);

        mine(genesis, 0, 1, newsockfd);

        int statloc;
        wait(&statloc);
    }
    else if (pid == 0)
    {
        node = "child";

        int sockfd;
        struct sockaddr_in serv_addr;
        char message[sizeof(Block)];
        int len;

        sockfd = socket(PF_INET, SOCK_STREAM, 0);
        if (sockfd == -1)
            error("socket() error");

        memset(&serv_addr, 0, sizeof(serv_addr));
        serv_addr.sin_family = AF_INET;
        serv_addr.sin_addr.s_addr = inet_addr("127.0.0.1");
        serv_addr.sin_port = htons(9000);

        if (connect(sockfd, (struct sockaddr *)&serv_addr, sizeof(serv_addr)) == -1)
            error("connect() error");

        int opt;
        opt = fcntl(sockfd, F_GETFL, 0);
        fcntl(sockfd, F_SETFL, opt | O_NONBLOCK);

        mine(genesis, 10000, -1, sockfd);
    }
    else
    {
        error("fork() error");
    }
    return 0;
}
