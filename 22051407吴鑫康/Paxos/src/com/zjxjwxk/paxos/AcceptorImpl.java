package com.zjxjwxk.paxos;

/**
 * 接受者实现类
 * @author Xinkang Wu
 * @date 2021/1/18 14:18
 */
public class AcceptorImpl implements Acceptor {
    /**
     * 接受者ID
     */
    private int id;
    /**
     * 承诺ID
     * 承诺忽略小于该ID的提议
     */
    private long promiseId;
    /**
     * 已接受的最大ID
     */
    private long acceptedId;
    /**
     * 已接受的最大ID对应的值
     */
    private String acceptedValue;
    /**
     * 信使
     */
    private final Messenger messenger;

    public AcceptorImpl(Messenger messenger) {
        this.promiseId = -1;
        this.acceptedId = -1;
        this.acceptedValue = null;
        this.messenger = messenger;
    }

    @Override
    public void receivePrepareRequest(int proposerId, long proposalId) {
        // 如果还没有做出过承诺或者该提议ID大于等于承诺ID
        if (this.promiseId == -1 || proposalId >= this.promiseId) {
            // 更新承诺ID
            this.promiseId = proposalId;
            // 发送承诺回复
            this.messenger.sendPromise(this.id, proposerId, this.promiseId, this.acceptedId, this.acceptedValue);
        }
        // 否则，该提议ID小于承诺ID，忽略该准备请求
    }

    @Override
    public void receiveAcceptRequest(long proposalId, String proposalValue) {
        // 如果还没有做出过承诺或者该提议ID大于等于承诺ID
        if (this.promiseId == -1 || proposalId >= this.promiseId) {
            // 更新承诺ID、已接受的最大ID、已接受的最大ID对应的值
            this.promiseId = proposalId;
            this.acceptedId = proposalId;
            this.acceptedValue = proposalValue;
            // 发送接受回复（广播给所有提议者和学习者）
            this.messenger.sendAccepted(this.acceptedId, this.acceptedValue);
        }
        // 否则，该提议ID小于承诺ID，忽略该接受请求
    }

    public int getId() {
        return id;
    }

    public long getPromiseId() {
        return promiseId;
    }

    public long getAcceptedId() {
        return acceptedId;
    }

    public String getAcceptedValue() {
        return acceptedValue;
    }
}
