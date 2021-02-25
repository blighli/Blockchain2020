package bitcoindemo.element;




import bitcoindemo.transaction.Transaction;
import bitcoindemo.util.CryptUtil;
import bitcoindemo.util.TreeUtil;

import java.util.*;

/**
 * @author LiuRuiyang
 * @date 2021/1/31 3:41 下午
 * 区块信息
 *
 */
public class Block {

    //区块的hash，即数字签名
    public String hash;
    //前一个区块的hash
    public String previousHash;
    //hash根
    public String merkleRoot;
    //交易列表
    public List<Transaction> transactions = new ArrayList<>();
    //时间戳
    private final long timeStamp;
    //随机数，使得区块hash的前几位按照要求为0
    private int nonce;

    public Block(String previousHash) {
        this.previousHash = previousHash;
        this.timeStamp = new Date().getTime();
        this.hash = calculateHash();
    }

    public String calculateHash() {
        String calculatedHash = CryptUtil.applySha256(
                previousHash +
                        Long.toString(timeStamp) +
                        Integer.toString(nonce) +
                        merkleRoot
        );
        return calculatedHash;
    }

    /**
     * 不停的重复计算hash，直到值小于目标值
     *
     * @param difficulty
     */
    public void mineBlock(int difficulty) {
        merkleRoot = TreeUtil.getMerkleRoot(transactions);
        String target = new String(new char[difficulty]).replace('\0', '0');
        Random random = new Random();
        while (!hash.substring(0, difficulty).equals(target)) {
            nonce = random.nextInt();
            hash = calculateHash();
        }
        System.out.println("出块成功，哈希值为 " + hash);
    }


    public boolean addTransaction(Transaction transaction) {
        if (Objects.isNull(transaction)) {
            return false;
        }

        if (!Objects.equals(previousHash, "0")) {
            if (!transaction.processTransaction()) {
                System.out.println("交易处理失败！");
                return false;
            }
        }
        transactions.add(transaction);
        System.out.println("交易已成功打包");
        return true;
    }

}
