package bitcoindemo.transaction;


import bitcoindemo.util.CryptUtil;

import java.security.PublicKey;

/**
 * @author LiuRuiyang
 * @date 2021/1/31 3:43 下午
 */
public class TransactionOutput {
    public String id;
    public PublicKey recipient;
    public float value;
    public String parentTransactionId;

    public TransactionOutput(PublicKey recipient, float value, String parentTransactionId) {
        this.recipient = recipient;
        this.value = value;
        this.parentTransactionId = parentTransactionId;
        this.id = CryptUtil.applySha256(CryptUtil.getStringFromKey(recipient) +
                Float.toString(value) + parentTransactionId);

    }

    public boolean isMine(PublicKey publicKey) {
        return (publicKey == recipient);
    }

}
