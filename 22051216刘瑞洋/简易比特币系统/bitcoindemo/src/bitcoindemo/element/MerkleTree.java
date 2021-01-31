package bitcoindemo.element;

import bitcoindemo.transaction.Transaction;
import bitcoindemo.util.CryptUtil;

import java.util.ArrayList;
import java.util.List;

/**
 * @author LiuRuiyang
 * @date 2021/1/31 3:46 下午
 */
public class MerkleTree {

    private List<String> txs;
    private String root;

    public MerkleTree(List<Transaction> transactions) {
        txs = new ArrayList<>();
        for (int i = 0; i < transactions.size(); i++) {
            txs.add(transactions.get(i).txHash);
        }
        root = "";
    }

    public String buildTree() {
        if (txs.isEmpty()) {
            return null;
        }
        if (txs.size() == 1) {
            return txs.get(0);
        }
        List<String> tempTxs = new ArrayList<>();
        tempTxs.addAll(txs);

        List<String> newTxs = getNewTxList(tempTxs);
        while (newTxs.size() > 1) {
            newTxs = getNewTxList(newTxs);
        }

        root = newTxs.get(0);
        return root;
    }

    private List<String> getNewTxList(List<String> txs) {
        List<String> newTxs = new ArrayList<>();
        int index = 0;
        while (index < txs.size()) {
            String left = txs.get(index);
            index++;

            String right = "";
            if (index != txs.size()) {
                right = txs.get(index);
            }
            newTxs.add(CryptUtil.applySha256(left + right));
            index++;
        }
        return newTxs;
    }


}
