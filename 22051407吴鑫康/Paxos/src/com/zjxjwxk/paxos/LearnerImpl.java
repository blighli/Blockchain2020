package com.zjxjwxk.paxos;

import java.util.HashMap;
import java.util.Map;

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
    /**
     * 大多数接受者的数量
     */
    private final int majorityNum;
    /**
     * 收到的接受者ID到提议ID的映射
     */
    private final Map<Integer, Long> acceptorIdToProposalIdMap;
    /**
     * 收到的提议ID到提议的映射
     */
    private final Map<Long, Proposal> proposalMap;

    private final Messenger messenger;

    public LearnerImpl(long acceptedId, String acceptedValue, int majorityNum, Messenger messenger) {
        this.acceptedId = -1;
        this.acceptedValue = null;
        this.majorityNum = majorityNum;
        this.acceptorIdToProposalIdMap = new HashMap<>();
        this.proposalMap = new HashMap<>();
        this.messenger = messenger;
    }

    @Override
    public void receiveAccepted(int acceptorId, long acceptedId, String acceptedValue) {
        // 如果还没有完成学习
        if (this.acceptedValue != null) {
            // 获取最近从该接受者收到的提议ID
            Long preProposalId = acceptorIdToProposalIdMap.get(acceptorId);
            // 如果之前没有收到过任何提议，或者没有收到过来自该接受者的更大ID的提议
            if (preProposalId == null || acceptedId > preProposalId) {
                // 更新该接受者的提议ID
                acceptorIdToProposalIdMap.put(acceptorId, acceptedId);
                // 如果曾经收到过提议，则该提议的保留计数减一（为零时则删除）
                if (preProposalId != null) {
                    Proposal preProposal = proposalMap.get(preProposalId);
                    if (--preProposal.retentionCount == 0) {
                        proposalMap.remove(preProposalId);
                    }
                }
                // 将收到的提议加入映射
                proposalMap.putIfAbsent(acceptedId, new Proposal(0, 0, acceptedValue));
                Proposal proposal = proposalMap.get(acceptedId);
                // 将相应提议的被接受数和保留数加一
                ++proposal.acceptedCount;
                ++proposal.retentionCount;
                // 当被接受数达到一个大多数接受者的数量时，学习完成
                if (proposal.acceptedCount == this.majorityNum) {
                    this.acceptedId = acceptedId;
                    this.acceptedValue = acceptedValue;
                    proposalMap.clear();
                    acceptorIdToProposalIdMap.clear();
                    messenger.sendLearned(acceptedId, acceptedValue);
                }
            }
        }
    }

    public int getMajorityNum() {
        return majorityNum;
    }

    public long getAcceptedId() {
        return acceptedId;
    }

    public String getAcceptedValue() {
        return acceptedValue;
    }

    private class Proposal {
        public int retentionCount;
        public int acceptedCount;
        public String acceptedValue;

        public Proposal(int retentionCount, int acceptedCount, String acceptedValue) {
            this.retentionCount = retentionCount;
            this.acceptedCount = acceptedCount;
            this.acceptedValue = acceptedValue;
        }
    }

}
