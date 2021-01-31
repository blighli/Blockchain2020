package bitcoindemo.test;

import bitcoindemo.element.Block;
import bitcoindemo.element.Wallet;
import bitcoindemo.transaction.Transaction;
import bitcoindemo.transaction.TransactionInput;
import bitcoindemo.transaction.TransactionOutput;

import java.io.BufferedWriter;
import java.io.File;
import java.io.FileWriter;
import java.io.IOException;
import java.security.Security;
import java.util.*;

/**
 * @author LiuRuiyang
 * @date 2021/1/31 3:48 下午
 */
public class Main {


    //区块链，存储挖出的区块
    public static List<Block> blockChain = new ArrayList<>();

    //为了提高交易效率，使用额外的数据记录输出信息。
    public static Map<String, TransactionOutput> UTXOs = new HashMap<>();

    //可以通过设置不同的值来测试挖矿时间（不要设置太大，否则后果自负^_^）
    public static int difficulty = 4;

    //设置最小交易额
    public static final float minimumTransaction = 0.1f;
    //钱包
    public static Wallet wallet1;
    public static Wallet wallet2;


    //创世交易
    public static Transaction genesisTransaction;

    //目标时间
    public static Long targetTime = 200L;

    //前一个区块的hash
    public static String preHash;

    //每个区块打包时间
    public static List<Integer> timeCost = new ArrayList<>();

    //当前区块打包时间
    public static Integer nowTime;

    //交易金额
    public static Float cost = 8.0F;

    //总耗时
    public static Integer sumTime = 0;

    public static void main(String args[]) throws IOException {
        Security.addProvider(new org.bouncycastle.jce.provider.BouncyCastleProvider());
        genesisTest();
    }


    //构建区块链系统完整流程测试
    public static void genesisTest() throws IOException {

        wallet1 = new Wallet();
        wallet2 = new Wallet();
        Wallet baseCoin = new Wallet();
        //构建创世区块
        Block genesis = genesis(baseCoin, wallet1);
        preHash = genesis.hash;
        System.out.println("\n当前区块编号：0");
        int num = 1;
        //创建几个交易和区块进行测试
        do {

            System.out.println("\n当前区块编号：" + num++);
            preHash = trans(preHash, difficulty, wallet1, "1", wallet2, "2");
            getDifficulty();
            System.out.println("当前难度为：" + difficulty);
            System.out.println("当前时间为：" + nowTime);
            isBlockChainValid();

            System.out.println("\n当前区块编号：" + num++);
            preHash = trans(preHash, difficulty, wallet2, "2", wallet1, "1");
            getDifficulty();
            System.out.println("当前难度为：" + difficulty);
            System.out.println("当前时间为：" + nowTime);
            isBlockChainValid();

        } while (blockChain.size() <= 200);

        System.out.println("\n总耗时：" + sumTime);

    }

    private static void getDifficulty() {
        Integer sum = 0;

        if(timeCost.size() == 5){

            for (Integer time : timeCost) {

                sum += time;
            }
            int per = sum/5;
            System.out.println("当前平均出块速度为："+ per);
            int newDifficult =  (int)((targetTime * 1.0) / (per * 1.0) * (difficulty));
            System.out.println("当前raw难度为："+ newDifficult);
            if(newDifficult >= 5 ){
                difficulty = 5;
            }else if(newDifficult <= 3){
                difficulty = 3;
            }else {
                difficulty = newDifficult;
            }
        }
    }

    public static String trans(String preHash, int nowDifficulty, Wallet pay, String num1, Wallet get, String num2) {
        Block block = new Block(preHash);
        System.out.println("钱包" + num1 + "余额:" + pay.getBalance());
        System.out.println("钱包" + num1 + "转账向钱包" + num2 + "转账" + cost + "比特币");
        block.addTransaction(pay.sendFunds(get.publicKey, cost));

        Long beginTime = System.currentTimeMillis();
        addBlock(block, nowDifficulty);
        Long endTime = System.currentTimeMillis();

        nowTime = Math.toIntExact((endTime - beginTime));
        if (timeCost.size() == 5) {
            timeCost.remove(0);
        }
        sumTime += nowTime;
        timeCost.add((int) (endTime - beginTime));


        System.out.println("钱包" + num1 + "的余额: " + pay.getBalance());
        System.out.println("钱包" + num2 + "的余额: " + get.getBalance());

        return block.hash;
    }

    public static Block genesis(Wallet baseCoin, Wallet wallet) {

        //创世交易不需要任何input
        genesisTransaction = new Transaction(baseCoin.publicKey, wallet.publicKey, 10f, null);
        genesisTransaction.generateSignature(baseCoin.privateKey);
        genesisTransaction.txHash = "0";
        genesisTransaction.outputs.add(new TransactionOutput(genesisTransaction.recipient, genesisTransaction.value, genesisTransaction.txHash));
        UTXOs.put(genesisTransaction.outputs.get(0).id, genesisTransaction.outputs.get(0));

        //开始构建创世区块
        System.out.println("构建创世区块");
        Block genesisBlock = new Block("0");
        //向创世区块中加入第一笔交易
        genesisBlock.addTransaction(genesisTransaction);
        addBlock(genesisBlock, difficulty);
        return genesisBlock;
    }

    /**
     * 校验区块链中的信息是否有效
     *
     * @return
     */
    public static Boolean isBlockChainValid() {
        Block currentBlock;
        Block previousBlock;
        Map<String, TransactionOutput> tempUTXOs = new HashMap<>();
        tempUTXOs.put(genesisTransaction.outputs.get(0).id, genesisTransaction.outputs.get(0));

        String hashTarget = new String(new char[difficulty]).replace('\0', '0');

        for (int i = 1; i < blockChain.size(); i++) {
            currentBlock = blockChain.get(i);
            previousBlock = blockChain.get(i - 1);

            //首先校验当前区块是否有效
            if (!currentBlock.hash.equals(currentBlock.calculateHash())) {
                System.out.println("当前hash不相等");
                return false;
            }

            if (!previousBlock.hash.equals(currentBlock.previousHash)) {
                System.out.println("前一个hash不相等");
                return false;
            }

            if (!currentBlock.hash.substring(0, difficulty).equals(hashTarget)) {
                System.out.println("该区块还未成功出块");
                return false;
            }

            //校验当前区块中的每笔交易是否有效
            for (int j = 0; j < currentBlock.transactions.size(); j++) {
                Transaction currentTransaction = currentBlock.transactions.get(j);

                if (!currentTransaction.verifySignature()) {
                    System.out.println("签名" + j + "无效");
                    return false;
                }

                if (currentTransaction.getInputValue() != currentTransaction.getOutputValue()) {
                    System.out.println("交易金额" + j + "不相等");
                    return false;
                }

                //校验一笔交易中的每一个输入是否有效
                TransactionOutput tempOutput;
                for (TransactionInput input : currentTransaction.inputs) {
                    tempOutput = tempUTXOs.get(input.transactionOutputId);

                    if (tempOutput == null) {
                        System.out.println("无法找到输入交易：" + j);
                        return false;
                    }

                    if (input.UTXO.value != tempOutput.value) {
                        System.out.println("输入输出交易不匹配：" + j);
                        return false;
                    }

                    tempUTXOs.remove(input.transactionOutputId);
                }
                //将每笔交易得到的输出加入到临时余额集合中。
                for (TransactionOutput output : currentTransaction.outputs) {
                    tempUTXOs.put(output.id, output);
                }

                //校验当前交易的输出是否有效
                if (!Objects.equals(currentTransaction.outputs.get(0).recipient, currentTransaction.recipient)) {
                    System.out.println("当前输出无效");
                    return false;
                }
                //检验交易结余是否返回给sender
                if (!Objects.equals(currentTransaction.outputs.get(1).recipient, currentTransaction.sender)) {
                    System.out.println("交易结余未返还");
                    return false;
                }

            }
        }
        //校验通过
        System.out.println("校验通过");
        return true;
    }

    private static void addBlock(Block block, int difficulty) {

        block.mineBlock(difficulty);
        //完成工作量证明，上链
        blockChain.add(block);
    }


}
