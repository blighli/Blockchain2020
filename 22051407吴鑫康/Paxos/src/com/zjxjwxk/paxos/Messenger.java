package com.zjxjwxk.paxos;

/**
 * 信使
 * @author Xinkang Wu
 * @date 2021/1/18 11:12
 */
public interface Messenger {

    /**
     * 发送准备请求
     * @param proposerId 提议者ID
     * @param proposalId 提议ID
     */
    void sendPrepareRequest(int proposerId, long proposalId);

    /**
     * 发送承诺回复
     * @param acceptorId 接受者ID（from）
     * @param proposerId 提议者ID（to）
     * @param promiseId 提议ID
     * @param acceptedId 已接受ID
     * @param acceptedValue 已接受值
     */
    void sendPromise(int acceptorId, int proposerId, long promiseId, long acceptedId, String acceptedValue);

    /**
     * 发送接收请求
     * @param proposalId 提议ID
     * @param proposalValue 提议值
     */
    void sendAcceptRequest(long proposalId, String proposalValue);

    /**
     * 发送已接受回复
     * @param acceptedId 已接受ID
     * @param acceptedValue 已接受值
     */
    void sendAccepted(long acceptedId, String acceptedValue);

    /**
     * 发送已学习到的
     * @param acceptedId 已接受ID
     * @param acceptedValue 已接受值
     */
    void sendLearned(long acceptedId, String acceptedValue);
}
