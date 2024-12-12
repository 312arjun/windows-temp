package eclipz

const (
	// Client to Controller
	CmdLoginReq    = "LoginReq"
	CmdClientInfo  = "ClientInfo"
	CmdPeerInfoReq = "PeerInfoReq"
	CmdLogoutReq   = "Logout"
	CmdStatsReq    = "Stats"

	// Controller to Client
	CmdLoginResp    = "LoginResp"
	CmdLogoutResp   = "LogoutResp"
	CmdClientParams = "ClientParams" // Response to ClientInfo
	CmdPeerInfo     = "PeerInfo"     // Response to PeerInfoReq or asynchronous
	CmdClientNotify = "ClientNotify" // notification from conroller to clients

	// Notifys
	NotifyServiceOnline = "ServiceOnline"
	NotifyClientReady   = "ClientReady"

	RoleUser    = "U"
	RoleService = "S"

	// Status Codes
	StatusSuccess          = 0
	StatusBadRequest       = 1
	StatusPermissionDenied = 2
	StatusLoginFailed      = 3
	StatusInternalFailure  = 4
	StatusInvalidToken     = 5
	StatusMax              = 6 // Last Status Value
)

var ErrorText [StatusMax]string = [StatusMax]string{
	/*  0 */ "Success",
	/*  1 */ "Bad Request",
	/*  2 */ "Permission Denied",
	/*  3 */ "Login Failed",
	/*  4 */ "Internal Failure",
	/*  5 */ "Invalid Token",
}

type RpcCommon struct {
	Cmd        string `json:"cmd"`
	MsgId      string `json:"msgid"`
	Token      string `json:"token,omitempty"`
	Status     int    `json:"status,omitempty"`
	StatusText string `json:"text,omitempty"`
}

type RpcLoginReq struct {
	// Common Header
	Cmd   string `json:"cmd"`
	MsgId string `json:"msgid"`
	Token string `json:"token,omitempty"`

	// Login Request
	Name         string `json:"name"`
	Domain       string `json:"domain"`
	Role         string `json:"role,omitempty"`
	DeviceId     string `json:"device_id,omitempty"`
	Password     string `json:"pass"`
	NextPassword string `json:"next_pass,omitempty"`
}

type RpcLoginResp struct {
	// Common Header
	Cmd        string `json:"cmd"`
	MsgId      string `json:"msgid"`
	Status     int    `json:"status"`
	StatusText string `json:"text,omitempty"`

	// Login Response
	Token           string `json:"token,omitempty"`
	WgControllerKey string `json:"wg_controller_key,omitempty"`
	VirtualIP       string `json:"vip,omitempty"`
	RequirePassword bool   `json:"require_password,omitempty"`
}

type RpcLogoutReq struct {
	// Common Header
	Cmd   string `json:"cmd"`
	MsgId string `json:"msgid"`
	Token string `json:"token,omitempty"`

	// Logout Request
	Suspend bool `json:"suspend,omitempty"`
}

// Send from Client (both User and Service) to Controller
type RpcClientInfo struct {
	// Common Header
	Cmd   string `json:"cmd"`
	MsgId string `json:"msgid"`
	Token string `json:"token,omitempty"`

	// ClientInfo
	WgKey       string   `json:"wg_key,omitempty"`
	WgIPAddress string   `json:"wg_ipaddress,omitempty"`
	WgPort      int      `json:"wg_port,omitempty"`
	PublicIP    string   `json:"public_ip,omitempty"`
	LocalIPs    []string `json:"local_ip,omitempty"`
}

// Application
type App struct {
	Name       string `json:"name,omitempty"`
	Service    string `json:"service,omitempty"`
	AllowedIPs string `json:"allowed_ips,omitempty"`
}

// Send from Controller to User and Service clients in response to ClientInfo request
type RpcClientParams struct {
	// Common Header
	Cmd        string `json:"cmd"`
	MsgId      string `json:"msgid"`
	Token      string `json:"token,omitempty"`
	Status     int    `json:"status"`
	StatusText string `json:"text,omitempty"`

	WgPublicAddress string `json:"wg_public_ip,omitempty"`
	VirtualIP       string `json:"vip,omitempty"`
	Apps            []*App `json:"apps,omitempty"`
}

// Send from User client to Controller when it needs to initiate a connection to App
type RpcPeerInfoReq struct {
	Cmd   string `json:"cmd"`
	MsgId string `json:"msgid"`
	Token string `json:"token,omitempty"`
	App   App    `json:"app,omitempty"`
}

// Send from Controller to both User and Service
// AllowedIPs refers to the IP addresses allowed on the Service side
type RpcPeerInfo struct {
	// Common Header
	Cmd        string `json:"cmd"`
	MsgId      string `json:"msgid"`
	Token      string `json:"token,omitempty"`
	Status     int    `json:"status"`
	StatusText string `json:"text,omitempty"`

	// PeerInfo
	PeerName       string   `json:"peer_name,omitempty"`
	EnclaveId      string   `json:"eid,omitempty"`
	Role           string   `json:"role,omitempty"`
	Initiate       bool     `json:"initiate,omitempty"`
	WgKey          string   `json:"wg_key,omitempty"`
	WgPresharedKey string   `json:"wg_presharedkey,omitempty"`
	WgIPAddress    string   `json:"wg_ipaddress,omitempty"`
	WgPort         int      `json:"wg_port,omitempty"`
	VirtualIP      string   `json:"vip,omitempty"`
	PublicIP       string   `json:"public_ip,omitempty"`
	LocalIPs       []string `json:"local_ip,omitempty"`
	AllowedIPs     []string `json:"allowed_ips,omitempty"`
}

// Send from Controller to Client when a Service comes online
type RpcClientNotify struct {
	// Common Header
	Cmd        string `json:"cmd"`
	MsgId      string `json:"msgid"`
	Token      string `json:"token,omitempty"`
	Status     int    `json:"status"`
	StatusText string `json:"text,omitempty"`

	NotifyType  string `json:"notify_type,omitempty"`
	ServiceName string `json:"service_name,omitempty"`
}

type StatEntry struct {
	Name      string `json:"name,omitempty"`
	EnclaveId string `json:"eid,omitempty"`
	RxBytes   int64  `json:"rx,omitempty"`
	TxBytes   int64  `json:"tx,omitempty"`
}

// Send from Clients and Services to Controller
type RpcStats struct {
	// Common Header
	Cmd   string `json:"cmd"`
	MsgId string `json:"msgid"`
	Token string `json:"token,omitempty"`

	Stats []*StatEntry `json:"stats,omitempty"`
}
