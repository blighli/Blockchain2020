package bitcoindemo.transaction;



import bitcoindemo.test.Main;
import bitcoindemo.util.CryptUtil;

import java.security.PrivateKey;
import java.security.PublicKey;
import java.util.ArrayList;
import java.util.List;

/**
 * @author LiuRuiyang
 * @date 2021/1/31 3:43 下午
 */
public class Transaction {

    //也是本次交易的hash值
    public String txHash;

    //发送者和接受者的公钥
    public PublicKey sender;
    public PublicKey recipient;

    //交易金额
    public float value;

    //数字签名保证两点：1.持有者的所有权 2.安全性，交易内容不会被篡改
    public byte[] signature;

    public List<TransactionInput> inputs = new ArrayList<>();
    public List<TransactionOutput> outputs = new ArrayList<>();

    //统计交易频次
    private static int sequence = 0;

    public Transaction(PublicKey from, PublicKey to, float value, List<TransactionInput> inputs) {
        this.sender = from;
        this.recipient = to;
        this.value = value;
        this.inputs = inputs;
    }

    /**
     * 生成签名，签名内容为发送者和接受者公钥以及交易金额
     * @param privateKey 发送者私钥
     */
    public void generateSignature(PrivateKey privateKey) {
        String data = CryptUtil.getStringFromKey(sender) + CryptUtil.getStringFromKey(recipient) +
                Float.toString(value);
        signature = CryptUtil.applyECDSASig(privateKey, data);
    }

    /**
     * 验证签名
     * @return 验签结果
     */
    public boolean verifySignature() {
        String data = CryptUtil.getStringFromKey(sender) + CryptUtil.getStringFromKey(recipient) +
                Float.toString(value);
        return CryptUtil.verifyECDSASig(sender,data,signature);
    }

    /**
     * 处理交易
     * @return 处理结果
     */
    public boolean processTransaction(){

        //验证签名
        if(!verifySignature()){
            System.out.println("验签失败！");
            return false;
        }


        for(TransactionInput input:inputs){
            input.UTXO = Main.UTXOs.get(input.transactionOutputId);
        }

        if(getInputValue()< Main.minimumTransaction){
            System.out.println("交易额小于最小数量: " + getInputValue());
            return false;
        }

        txHash = calculateHash();

        float leftOver = getInputValue() - value;
        if(leftOver < 0.0){
            System.out.println("小于最小交易数量" + getInputValue()+",value:"+value);
            return false;
        }

        outputs.add(new TransactionOutput(this.recipient,value, txHash));
        outputs.add(new TransactionOutput(this.sender,leftOver, txHash));

        //将所有output加入到hashmap
        for(TransactionOutput output:outputs){
            Main.UTXOs.put(output.id,output);
        }

        //将已经被使用了的input从hashmap中删除
        for(TransactionInput input:inputs){
            if(input.UTXO!=null){
                Main.UTXOs.remove(input.UTXO.id);
            }
        }

        return true;
    }

    public float getInputValue(){
        float total = 0;
        for(TransactionInput input:inputs ){
            if(input.UTXO == null){
                continue;
            }
            total += input.UTXO.value;
        }
        return total;
    }

    public float getOutputValue(){
        float total = 0;
        for(TransactionOutput output: outputs){
            total += output.value;
        }
        return total;
    }

    private String calculateHash() {
        sequence++;
        return CryptUtil.applySha256(
                CryptUtil.getStringFromKey(sender) +
                        CryptUtil.getStringFromKey(recipient) +
                        Float.toString(value) + sequence
        );
    }
}
