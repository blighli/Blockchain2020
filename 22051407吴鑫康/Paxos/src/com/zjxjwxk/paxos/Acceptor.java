package com.zjxjwxk.paxos;

/**
 * 接受者接口
 * @author Xinkang Wu
 * @date 2021/1/18 10:57
 */
public interface Acceptor {

    /**
     * 收到准备请求
     * @param proposerId 提议者ID（from）
     * @param proposalId 提议ID
     */
    void receivePrepareRequest(int proposerId, long proposalId);

    /**
     * 收到接收请求
     * @param proposalId 提议ID
     * @param proposalValue 提议值
     */
    void receiveAcceptRequest(long proposalId, String proposalValue);
}
