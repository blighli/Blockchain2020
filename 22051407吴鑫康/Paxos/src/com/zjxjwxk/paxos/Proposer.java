package com.zjxjwxk.paxos;

/**
 * 提议者接口
 * @author Xinkang Wu
 * @date 2021/1/18 10:57
 */
public interface Proposer {
    /**
     * 设置提议值
     * @param proposalValue 提议值
     */
    void setProposalValue(String proposalValue);

    /**
     * 设置大多数接受者的数量
     * @param majorityNum 大多数接受者的数量
     */
    void setMajorityNum(int majorityNum);

    /**
     * 发送准备请求
     */
    void prepareRequest();

    /**
     * 收到承诺回复，若收到大多数接受者的承诺回复，则发送接受请求
     * @param acceptorId 接受者ID（from）
     * @param promiseId 承诺ID
     * @param acceptedId 已接受ID
     * @param acceptedValue 已接受值
     */
    void receivePromise(int acceptorId, long promiseId, long acceptedId, String acceptedValue);
}
