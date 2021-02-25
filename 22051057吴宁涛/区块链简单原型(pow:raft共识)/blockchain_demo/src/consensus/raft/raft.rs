
use std::collections::HashSet;
use std::sync::Arc;

use serde::{Deserialize, Serialize};
use tokio::sync::{mpsc, oneshot, watch, Mutex};
use tokio::task::JoinHandle;

use crate::config::Config;
use crate::core::RaftCore;
use crate::error::{ChangeConfigError, ClientReadError, ClientWriteError, InitializeError, RaftError, RaftResult};
use crate::metrics::RaftMetrics;
use crate::{AppData, AppDataResponse, NodeId, RaftNetwork, RaftStorage};

struct RaftInner<D: AppData, R: AppDataResponse, N: RaftNetwork<D>, S: RaftStorage<D, R>> {
    tx_api: mpsc::UnboundedSender<RaftMsg<D, R>>,
    rx_metrics: watch::Receiver<RaftMetrics>,
    raft_handle: Mutex<Option<JoinHandle<RaftResult<()>>>>,
    tx_shutdown: Mutex<Option<oneshot::Sender<()>>>,
    marker_n: std::marker::PhantomData<N>,
    marker_s: std::marker::PhantomData<S>,
}

pub struct Raft<D: AppData, R: AppDataResponse, N: RaftNetwork<D>, S: RaftStorage<D, R>> {
    inner: Arc<RaftInner<D, R, N, S>>,
}

impl<D: AppData, R: AppDataResponse, N: RaftNetwork<D>, S: RaftStorage<D, R>> Raft<D, R, N, S> {
    pub fn new(id: NodeId, config: Arc<Config>, network: Arc<N>, storage: Arc<S>) -> Self {
        let (tx_api, rx_api) = mpsc::unbounded_channel();
        let (tx_metrics, rx_metrics) = watch::channel(RaftMetrics::new_initial(id));
        let (tx_shutdown, rx_shutdown) = oneshot::channel();
        let raft_handle = RaftCore::spawn(id, config, network, storage, rx_api, tx_metrics, rx_shutdown);
        let inner = RaftInner {
            tx_api,
            rx_metrics,
            raft_handle: Mutex::new(Some(raft_handle)),
            tx_shutdown: Mutex::new(Some(tx_shutdown)),
            marker_n: std::marker::PhantomData,
            marker_s: std::marker::PhantomData,
        };
        Self { inner: Arc::new(inner) }
    }

    #[tracing::instrument(level = "debug", skip(self, rpc))]
    pub async fn append_entries(&self, rpc: AppendEntriesRequest<D>) -> Result<AppendEntriesResponse, RaftError> {
        let (tx, rx) = oneshot::channel();
        self.inner
            .tx_api
            .send(RaftMsg::AppendEntries { rpc, tx })
            .map_err(|_| RaftError::ShuttingDown)?;
        Ok(rx.await.map_err(|_| RaftError::ShuttingDown).and_then(|res| res)?)
    }

    #[tracing::instrument(level = "debug", skip(self, rpc))]
    pub async fn vote(&self, rpc: VoteRequest) -> Result<VoteResponse, RaftError> {
        let (tx, rx) = oneshot::channel();
        self.inner
            .tx_api
            .send(RaftMsg::RequestVote { rpc, tx })
            .map_err(|_| RaftError::ShuttingDown)?;
        Ok(rx.await.map_err(|_| RaftError::ShuttingDown).and_then(|res| res)?)
    }

    #[tracing::instrument(level = "debug", skip(self, rpc))]
    pub async fn install_snapshot(&self, rpc: InstallSnapshotRequest) -> Result<InstallSnapshotResponse, RaftError> {
        let (tx, rx) = oneshot::channel();
        self.inner
            .tx_api
            .send(RaftMsg::InstallSnapshot { rpc, tx })
            .map_err(|_| RaftError::ShuttingDown)?;
        Ok(rx.await.map_err(|_| RaftError::ShuttingDown).and_then(|res| res)?)
    }

    #[tracing::instrument(level = "debug", skip(self))]
    pub async fn current_leader(&self) -> Option<NodeId> {
        self.metrics().borrow().current_leader
    }

    #[tracing::instrument(level = "debug", skip(self))]
    pub async fn client_read(&self) -> Result<(), ClientReadError> {
        let (tx, rx) = oneshot::channel();
        self.inner
            .tx_api
            .send(RaftMsg::ClientReadRequest { tx })
            .map_err(|_| ClientReadError::RaftError(RaftError::ShuttingDown))?;
        Ok(rx
            .await
            .map_err(|_| ClientReadError::RaftError(RaftError::ShuttingDown))
            .and_then(|res| res)?)
    }

    #[tracing::instrument(level = "debug", skip(self, rpc))]
    pub async fn client_write(&self, rpc: ClientWriteRequest<D>) -> Result<ClientWriteResponse<R>, ClientWriteError<D>> {
        let (tx, rx) = oneshot::channel();
        self.inner
            .tx_api
            .send(RaftMsg::ClientWriteRequest { rpc, tx })
            .map_err(|_| ClientWriteError::RaftError(RaftError::ShuttingDown))?;
        Ok(rx
            .await
            .map_err(|_| ClientWriteError::RaftError(RaftError::ShuttingDown))
            .and_then(|res| res)?)
    }

    #[tracing::instrument(level = "debug", skip(self))]
    pub async fn initialize(&self, members: HashSet<NodeId>) -> Result<(), InitializeError> {
        let (tx, rx) = oneshot::channel();
        self.inner
            .tx_api
            .send(RaftMsg::Initialize { members, tx })
            .map_err(|_| RaftError::ShuttingDown)?;
        Ok(rx
            .await
            .map_err(|_| InitializeError::RaftError(RaftError::ShuttingDown))
            .and_then(|res| res)?)
    }

    #[tracing::instrument(level = "debug", skip(self))]
    pub async fn add_non_voter(&self, id: NodeId) -> Result<(), ChangeConfigError> {
        let (tx, rx) = oneshot::channel();
        self.inner
            .tx_api
            .send(RaftMsg::AddNonVoter { id, tx })
            .map_err(|_| RaftError::ShuttingDown)?;
        Ok(rx
            .await
            .map_err(|_| ChangeConfigError::RaftError(RaftError::ShuttingDown))
            .and_then(|res| res)?)
    }

    #[tracing::instrument(level = "debug", skip(self))]
    pub async fn change_membership(&self, members: HashSet<NodeId>) -> Result<(), ChangeConfigError> {
        let (tx, rx) = oneshot::channel();
        self.inner
            .tx_api
            .send(RaftMsg::ChangeMembership { members, tx })
            .map_err(|_| RaftError::ShuttingDown)?;
        Ok(rx
            .await
            .map_err(|_| ChangeConfigError::RaftError(RaftError::ShuttingDown))
            .and_then(|res| res)?)
    }

    pub fn metrics(&self) -> watch::Receiver<RaftMetrics> {
        self.inner.rx_metrics.clone()
    }

    pub async fn shutdown(&self) -> anyhow::Result<()> {
        if let Some(tx) = self.inner.tx_shutdown.lock().await.take() {
            let _ = tx.send(());
        }
        if let Some(handle) = self.inner.raft_handle.lock().await.take() {
            let _ = handle.await?;
        }
        Ok(())
    }
}

impl<D: AppData, R: AppDataResponse, N: RaftNetwork<D>, S: RaftStorage<D, R>> Clone for Raft<D, R, N, S> {
    fn clone(&self) -> Self {
        Self { inner: self.inner.clone() }
    }
}

pub(crate) type ClientWriteResponseTx<D, R> = oneshot::Sender<Result<ClientWriteResponse<R>, ClientWriteError<D>>>;
pub(crate) type ClientReadResponseTx = oneshot::Sender<Result<(), ClientReadError>>;
pub(crate) type ChangeMembershipTx = oneshot::Sender<Result<(), ChangeConfigError>>;

pub(crate) enum RaftMsg<D: AppData, R: AppDataResponse> {
    AppendEntries {
        rpc: AppendEntriesRequest<D>,
        tx: oneshot::Sender<Result<AppendEntriesResponse, RaftError>>,
    },
    RequestVote {
        rpc: VoteRequest,
        tx: oneshot::Sender<Result<VoteResponse, RaftError>>,
    },
    InstallSnapshot {
        rpc: InstallSnapshotRequest,
        tx: oneshot::Sender<Result<InstallSnapshotResponse, RaftError>>,
    },
    ClientWriteRequest {
        rpc: ClientWriteRequest<D>,
        tx: ClientWriteResponseTx<D, R>,
    },
    ClientReadRequest {
        tx: ClientReadResponseTx,
    },
    Initialize {
        members: HashSet<NodeId>,
        tx: oneshot::Sender<Result<(), InitializeError>>,
    },
    AddNonVoter {
        id: NodeId,
        tx: ChangeMembershipTx,
    },
    ChangeMembership {
        members: HashSet<NodeId>,
        tx: ChangeMembershipTx,
    },
}

#[derive(Debug, Serialize, Deserialize)]
pub struct AppendEntriesRequest<D: AppData> {
    pub term: u64,
    pub leader_id: u64,
    pub prev_log_index: u64,
    pub prev_log_term: u64,
    #[serde(bound = "D: AppData")]
    pub entries: Vec<Entry<D>>,
    pub leader_commit: u64,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct AppendEntriesResponse {
    pub term: u64,
    pub success: bool,
    pub conflict_opt: Option<ConflictOpt>,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct ConflictOpt {
    pub term: u64,
    pub index: u64,
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
pub struct Entry<D: AppData> {
    pub term: u64,
    pub index: u64,
    #[serde(bound = "D: AppData")]
    pub payload: EntryPayload<D>,
}

impl<D: AppData> Entry<D> {
    pub fn new_snapshot_pointer(index: u64, term: u64, id: String, membership: MembershipConfig) -> Self {
        Entry {
            term,
            index,
            payload: EntryPayload::SnapshotPointer(EntrySnapshotPointer { id, membership }),
        }
    }
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
pub enum EntryPayload<D: AppData> {
    Blank,
    #[serde(bound = "D: AppData")]
    Normal(EntryNormal<D>),
    ConfigChange(EntryConfigChange),
    SnapshotPointer(EntrySnapshotPointer),
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
pub struct EntryNormal<D: AppData> {
    #[serde(bound = "D: AppData")]
    pub data: D,
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
pub struct EntryConfigChange {
    pub membership: MembershipConfig,
}

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
pub struct EntrySnapshotPointer {
    pub id: String,
    pub membership: MembershipConfig,
}

#[derive(Clone, Debug, PartialEq, Eq, Serialize, Deserialize)]
pub struct MembershipConfig {
    pub members: HashSet<NodeId>,
    pub members_after_consensus: Option<HashSet<NodeId>>,
}

impl MembershipConfig {
    pub fn all_nodes(&self) -> HashSet<u64> {
        let mut all = self.members.clone();
        if let Some(members) = &self.members_after_consensus {
            all.extend(members);
        }
        all
    }

    pub fn contains(&self, x: &NodeId) -> bool {
        self.members.contains(x)
            || if let Some(members) = &self.members_after_consensus {
            members.contains(x)
        } else {
            false
        }
    }

    pub fn is_in_joint_consensus(&self) -> bool {
        self.members_after_consensus.is_some()
    }

    pub fn new_initial(id: NodeId) -> Self {
        let mut members = HashSet::new();
        members.insert(id);
        Self {
            members,
            members_after_consensus: None,
        }
    }
}

#[derive(Debug, Serialize, Deserialize)]
pub struct VoteRequest {
    pub term: u64,
    pub candidate_id: u64,
    pub last_log_index: u64,
    pub last_log_term: u64,
}

impl VoteRequest {
    pub fn new(term: u64, candidate_id: u64, last_log_index: u64, last_log_term: u64) -> Self {
        Self {
            term,
            candidate_id,
            last_log_index,
            last_log_term,
        }
    }
}

#[derive(Debug, Serialize, Deserialize)]
pub struct VoteResponse {
    pub term: u64,
    pub vote_granted: bool,
}

#[derive(Clone, Debug, Serialize, Deserialize)]
pub struct InstallSnapshotRequest {
    pub term: u64,
    pub leader_id: u64,
    pub last_included_index: u64,
    pub last_included_term: u64,
    pub offset: u64,
    pub data: Vec<u8>,
    pub done: bool,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct InstallSnapshotResponse {
    pub term: u64,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct ClientWriteRequest<D: AppData> {
    #[serde(bound = "D: AppData")]
    pub(crate) entry: EntryPayload<D>,
}

impl<D: AppData> ClientWriteRequest<D> {
    pub fn new(entry: D) -> Self {
        Self::new_base(EntryPayload::Normal(EntryNormal { data: entry }))
    }

    pub(crate) fn new_base(entry: EntryPayload<D>) -> Self {
        Self { entry }
    }

    pub(crate) fn new_config(membership: MembershipConfig) -> Self {
        Self::new_base(EntryPayload::ConfigChange(EntryConfigChange { membership }))
    }

    pub(crate) fn new_blank_payload() -> Self {
        Self::new_base(EntryPayload::Blank)
    }
}

#[derive(Debug, Serialize, Deserialize)]
pub struct ClientWriteResponse<R: AppDataResponse> {
    pub index: u64,
    #[serde(bound = "R: AppDataResponse")]
    pub data: R,
}