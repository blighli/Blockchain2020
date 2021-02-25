package com.zjxjwxk.paxos;

/**
 * 学习者接口
 * @author Xinkang Wu
 * @date 2021/1/18 10:57
 */
public interface Learner {

    /**
     * 收到接受回复
     * @param acceptorId 接受者ID（from）
     * @param acceptedId 接收ID
     * @param acceptedValue 接收值
     */
    void receiveAccepted(int acceptorId, long acceptedId, String acceptedValue);
}
