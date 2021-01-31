package bitcoindemo.util;

import bitcoindemo.element.MerkleTree;
import bitcoindemo.transaction.Transaction;

import java.util.List;

/**
 * @author LiuRuiyang
 * @date 2021/1/31 3:47 下午
 * 树工具
 * 1.用于获取merkle tree的根哈希
 */
public class TreeUtil {

    /**
     * 获取交易的merkle树的根节点，当一个区块中包含大量的transaction时，计算所有的hash值是不可取的
     * 所以使用merkel tree对全部hash值进行一个计算。
     *
     * @param transactions 交易列表
     * @return merkle tree 根哈希
     */
    public static String getMerkleRoot(List<Transaction> transactions) {
        MerkleTree merkleTree = new MerkleTree(transactions);
        return merkleTree.buildTree();
    }


}
