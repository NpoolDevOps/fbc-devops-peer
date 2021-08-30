package types

const (
	ParentSpecAPI = "/api/v0/peer/parentspec"
	HeartbeatAPI  = "/api/v0/peer/heartbeat"
	OperationAPI  = "/api/v0/peer/operation"
)

const (
	FullNode        = "fullnode"
	MinerNode       = "miner"
	FullMinerNode   = "fullminer"
	WorkerNode      = "worker"
	StorageNode     = "storage"
	GatewayNode     = "gateway"
	ChiaMinerNode   = "chiaminer"
	ChiaPlotterNode = "chiaplotter"
)

const (
	StorageRoleAPI = "api"
	StorageRoleMgr = "mgr"
	StorageRoleOsd = "osd"
)

const (
	ExporterPort = 52379
)
