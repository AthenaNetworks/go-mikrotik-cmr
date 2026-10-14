// Package model defines the CMR controller's managed-object types.
//
// These mirror the Nova bus "stores" that the MikroTik `cmr` binary registers
// in main() (reverse-engineered with Ghidra via REA):
//
//	/nova/store/cmr/dev        -> Device
//	/nova/store/cmr/wifiradio  -> WifiRadio
//	/nova/store/cmr/wifinet    -> WifiNet
//	/nova/store/cmr/vlan       -> VLAN
//	/nova/store/cmr/layout     -> Layout
//	/nova/store/cmr/laynode    -> LayNode
//	/nova/store/cmr/laylink    -> LayLink
//	/nova/store/cmr/upgrade    -> Upgrade (+ /upgrade/job -> UpgradeJob)
//	/nova/store/cmr/alert      -> Alert
//	/nova/store/cmr/inst       -> Instance (self/controller config)
//
// Field sets are taken from the binary's embedded column lists, e.g.
// "ID,NODE_ID,MAC,IP,NAME,NET,IS_WIFI,IS_GATEWAY,WIFI_BAND" and
// "ID,IDENTITY,IS_CMR,IS_MIKROTIK", plus the console menu property tree.
// Where a field's exact on-wire type was not recovered it is modeled with a
// pragmatic Go type and noted; see docs/PROTOCOL.md for provenance and the
// list of items still pending on-device validation.
package model

import (
	"net"
	"time"
)

// ID is the controller-assigned identifier used as the key in every store.
type ID uint32

// Band enumerates the Wi-Fi band reported per managed device (WIFI_BAND).
type Band string

const (
	Band2G   Band = "2ghz"
	Band5G   Band = "5ghz"
	Band6G   Band = "6ghz"
	BandNone Band = ""
)

// Device is one node managed by the controller (/nova/store/cmr/dev).
//
// Columns observed in the binary:
//
//	ID,NODE_ID,MAC,IP,NAME,NET,IS_WIFI,IS_GATEWAY,WIFI_BAND
//	ID,IDENTITY,IS_CMR,IS_MIKROTIK
type Device struct {
	ID         ID               `json:"id"`
	NodeID     ID               `json:"node_id"` // position in the layout graph
	MAC        net.HardwareAddr `json:"-"`       // enrollment identity (see MACString)
	IP         net.IP           `json:"ip,omitempty"`
	Name       string           `json:"name,omitempty"`
	Identity   string           `json:"identity,omitempty"` // RouterOS /system identity
	Net        string           `json:"net,omitempty"`      // CMR network name it belongs to
	IsWifi     bool             `json:"is_wifi"`
	IsGateway  bool             `json:"is_gateway"`
	IsCMR      bool             `json:"is_cmr"` // runs a cmr instance itself
	IsMikrotik bool             `json:"is_mikrotik"`
	Band       Band             `json:"wifi_band,omitempty"`

	// Firmware state, driving the upgrade subsystem.
	Version        string `json:"version,omitempty"`
	UpgradeVersion string `json:"upgrade_version,omitempty"`
	UpgradeError   string `json:"upgrade_error,omitempty"`

	// Enrollment bookkeeping (controller-local, not a RouterOS column).
	PeerID    string    `json:"peer_id,omitempty"` // PKEX-exchanged peer identity
	Enrolled  bool      `json:"enrolled"`
	FirstSeen time.Time `json:"first_seen"`
	LastSeen  time.Time `json:"last_seen"`
}

// MACString renders the MAC in the lower-case colon form RouterOS uses.
func (d *Device) MACString() string {
	if d.MAC == nil {
		return ""
	}
	return d.MAC.String()
}

// WifiRadio is a managed radio on a device (/nova/store/cmr/wifiradio).
// The channel.* property family comes from the console menu tree.
type WifiRadio struct {
	ID       ID     `json:"id"`
	Device   ID     `json:"device"`
	Band     Band   `json:"band,omitempty"`
	Channel  string `json:"channel,omitempty"` // configuration.channel / frequency
	TxChains int    `json:"tx_chains,omitempty"`
	TxPower  int    `json:"tx_power,omitempty"`
	Country  string `json:"country,omitempty"`
	Disabled bool   `json:"disabled"`
}

// WifiNet is a managed SSID/network (/nova/store/cmr/wifinet). The security.*
// family (WPA3 SAE, 802.11r ft-*, EAP, etc.) comes from the console menu tree.
type WifiNet struct {
	ID               ID     `json:"id"`
	Name             string `json:"name,omitempty"`
	SSID             string `json:"ssid,omitempty"`
	Band             Band   `json:"band,omitempty"`
	VLAN             ID     `json:"vlan,omitempty"`
	SecurityProfile  string `json:"security_profile,omitempty"`
	Passphrase       string `json:"-"` // never serialized
	ManagementProtec bool   `json:"management_protection,omitempty"`
	Disabled         bool   `json:"disabled"`
}

// VLAN is a managed VLAN (/nova/store/cmr/vlan).
type VLAN struct {
	ID   ID     `json:"id"`
	VID  int    `json:"vid"`
	Name string `json:"name,omitempty"`
}

// Layout is the network map container (/nova/store/cmr/layout).
type Layout struct {
	ID   ID     `json:"id"`
	Name string `json:"name,omitempty"`
}

// LayNode is a node in the topology graph (/nova/store/cmr/laynode).
type LayNode struct {
	ID     ID  `json:"id"`
	Layout ID  `json:"layout"`
	Device ID  `json:"device"`
	X      int `json:"x"`
	Y      int `json:"y"`
}

// LayLink is an edge between two LayNodes (/nova/store/cmr/laylink).
type LayLink struct {
	ID     ID   `json:"id"`
	Layout ID   `json:"layout"`
	From   ID   `json:"from"`
	To     ID   `json:"to"`
	Wifi   bool `json:"wifi"`
}

// DevState mirrors the cmrids::JOB_DEVSTATE_* values referenced in the binary.
type DevState string

const (
	JobPending       DevState = "pending"
	JobPendReboot    DevState = "pending-reboot"    // JOB_DEVSTATE_PENDREBOOT
	JobReboot        DevState = "reboot"            // JOB_DEVSTATE_REBOOT
	JobRebootDisconn DevState = "reboot-disconnect" // JOB_DEVSTATE_REBOOTDISCONN
	JobDone          DevState = "done"
	JobFailed        DevState = "failed"
)

// Upgrade is an upgrade rule/campaign (/nova/store/cmr/upgrade).
type Upgrade struct {
	ID            ID     `json:"id"`
	TargetVersion string `json:"target_version,omitempty"`
	CheckInterval string `json:"upgrade_check_interval,omitempty"`
	Disabled      bool   `json:"disabled"`
}

// UpgradeJob is a per-device upgrade task (/nova/store/cmr/upgrade/job).
type UpgradeJob struct {
	ID      ID       `json:"id"`
	Device  ID       `json:"device"`
	State   DevState `json:"state"`
	Version string   `json:"version,omitempty"`
	Target  string   `json:"upgrade_version,omitempty"`
	Error   string   `json:"upgrade_error,omitempty"`
}

// Alert is a controller alert (/nova/store/cmr/alert).
type Alert struct {
	ID      ID        `json:"id"`
	Kind    string    `json:"kind"`
	Device  ID        `json:"device,omitempty"`
	Message string    `json:"message"`
	Time    time.Time `json:"time"`
}

// Instance is the controller's own configuration (/nova/store/cmr/inst).
type Instance struct {
	Identity            string   `json:"identity,omitempty"`
	Enabled             bool     `json:"enabled"`
	ControllerAddresses []string `json:"controller_addresses,omitempty"`
}
