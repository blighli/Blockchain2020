package com.zjxjwxk.paxos;

import java.util.HashSet;
import java.util.Set;

/**
 * 提议者实现类
 * @author Xinkang Wu
 * @date 2021/1/18 11:10
 */
public class ProposerImpl implements Proposer {
    /**
     * 提议者ID
     */
    private final int id;
    /**
     * 提议ID
     */
    private long proposalId;
    /**
     * 提议值
     */
    private String proposalValue;
    /**
     * 已接受的最大ID
     */
    private long acceptedId;
    /**
     * 大多数接受者的数量
     */
    private int majorityNum;
    /**
     * 已收到承诺回复的接受者ID集合
     */
    private final Set<Integer> promiseReceivedAcceptorIdSet;
    /**
     * 信使
     */
    private final Messenger messenger;

    public ProposerImpl(int id, int majorityNum, Messenger messenger) {
        this.id = id;
        this.proposalId = -1;
        this.proposalValue = null;
        this.acceptedId = -1;
        this.majorityNum = majorityNum;
        this.promiseReceivedAcceptorIdSet = new HashSet<>();
        this.messenger = messenger;
    }

    @Override
    public void setProposalValue(String proposalValue) {
        this.proposalValue = proposalValue;
    }

    @Override
    public void setMajorityNum(int majorityNum) {
        this.majorityNum = majorityNum;
    }

    @Override
    public void prepareRequest() {
        // 清空上一次准备请求收到的承诺回复集合
        this.promiseReceivedAcceptorIdSet.clear();
        // 使用时间戳作为提议ID
        this.proposalId = System.currentTimeMillis();
        // 发送准备请求
        this.messenger.sendPrepareRequest(this.id, this.proposalId);
    }

    @Override
    public void receivePromise(int acceptorId, long promiseId, long acceptedId, String acceptedValue) {
        // 判断是否是对最近发送提议的承诺回复，并且从未收到过来自该接受者ID的承诺回复
        if (promiseId == this.proposalId && !this.promiseReceivedAcceptorIdSet.contains(acceptorId)) {
            // 将该接受者ID加入已收到承诺回复的接受者ID集合
            this.promiseReceivedAcceptorIdSet.add(acceptorId);
            // 如果接受者曾接受过更大ID的提议
            if (acceptedId > this.acceptedId) {
                // 更新现有的已接受ID
                this.acceptedId = acceptedId;
                // 改变要提议的值为已接受值
                this.proposalValue = acceptedValue;
            }
            // 收到大多数接受者的承诺回复后，发送一个接受请求给大多数接受者
            if (this.promiseReceivedAcceptorIdSet.size() >= this.majorityNum && this.proposalId != -1 && this.proposalValue != null) {
                this.messenger.sendAcceptRequest(this.proposalId, this.proposalValue);
            }
        }
    }

    public int getId() {
        return id;
    }

    public long getProposalId() {
        return proposalId;
    }

    public String getProposalValue() {
        return proposalValue;
    }

    public long getAcceptedId() {
        return acceptedId;
    }

    public Set<Integer> getPromiseReceivedAcceptorIdSet() {
        return promiseReceivedAcceptorIdSet;
    }
}
