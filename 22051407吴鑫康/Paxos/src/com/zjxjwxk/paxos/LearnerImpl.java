package com.zjxjwxk.paxos;

/**
 * 学习者实现类
 * @author Xinkang Wu
 * @date 2021/1/18 17:29
 */
public class LearnerImpl implements Learner {
    /**
     * 已接受ID
     */
    private long acceptedId;
    /**
     * 已接受值
     */
    private String acceptedValue;

    public LearnerImpl() {
        this.acceptedId = -1;
        this.acceptedValue = null;
    }

    @Override
    public void receiveAccepted(int acceptorId, long acceptedId, String acceptedValue) {
        if (acceptedId > this.acceptedId) {
            this.acceptedId = acceptedId;
            this.acceptedValue = acceptedValue;
        }
    }

    public long getAcceptedId() {
        return acceptedId;
    }

    public String getAcceptedValue() {
        return acceptedValue;
    }
}
