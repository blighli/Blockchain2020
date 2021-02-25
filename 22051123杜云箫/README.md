### 1. RAFT算法实现和在次基础上的扩展

本次作业的raft实现部分主要参考了MIT的6.824:Distributed Systems，其中raft算法我首先是按照其lab的要求完成的，然后在改用了grpc，实现了一个可用的kv分布式存储系统系统。由于这学期我同时在实验室做了一篇raft结合spot instance的论文，所以在raft的基础上我又增加了一些新的角色Secretary和Observer，形成了一个新的协议。具体的实现文档在 http://jackdu.cn/distributed%20system/2020/12/20/Geo-Raft/
项目地址在 https://github.com/yunxiao3/MyRaft
（在本文件下也可以直接查看）其中raft目录是mit lab的实现，myraft是将其改为grpc通信，geo-raft则是新的协议。Notes中是lab实现的一些记录。

### 2.读书报告

由于这学期主要做的是一致性方面的工作，读书文档我主要是写的我对《Bitcoin: A Peer-to-Peer Electronic Cash System》中为什么使用Pow一致性协议的一些体会，主要包括：1. Bitcoin为什么使用proof of work的机制来确认大多数以达成一致，2. Bitcoin如何通过巧妙的设计在一个开放的系统中保持一致性并且防止forgery double spending。3. bitcoin的安全性并非牢不可破。