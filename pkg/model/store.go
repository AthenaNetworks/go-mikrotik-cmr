package model

import (
	"sync"
	"sync/atomic"
)

// EventType classifies a store mutation.
type EventType string

const (
	Created EventType = "created"
	Updated EventType = "updated"
	Deleted EventType = "deleted"
)

// Collection names each managed store (matches the /nova/store/cmr/* paths).
type Collection string

const (
	ColDevice     Collection = "dev"
	ColWifiRadio  Collection = "wifiradio"
	ColWifiNet    Collection = "wifinet"
	ColVLAN       Collection = "vlan"
	ColLayout     Collection = "layout"
	ColLayNode    Collection = "laynode"
	ColLayLink    Collection = "laylink"
	ColUpgrade    Collection = "upgrade"
	ColUpgradeJob Collection = "upgrade/job"
	ColAlert      Collection = "alert"
)

// Event is emitted on every store mutation. Obj is the current object value
// (a copy); for Deleted it is the last-known value.
type Event struct {
	Type       EventType
	Collection Collection
	ID         ID
	Obj        any
}

// Observer receives store events. Register with Store.Watch. Observers are
// called synchronously under no lock; do not block or re-enter the Store from
// within OnEvent — copy what you need and hand off to your own goroutine.
type Observer func(Event)

// Store is the in-memory home of all managed objects and the extensibility
// seam of the server: external projects react to changes via Watch and drive
// changes via the typed Put/Delete methods. It is safe for concurrent use.
//
// This replaces the Nova message bus the real cmr binary uses: the binary's
// main() registers each of these collections as a bus handler and other
// RouterOS processes read/write them. Here, the same objects live in-process
// and the Observer hooks take the place of bus subscriptions.
type Store struct {
	mu        sync.RWMutex
	nextID    uint32
	devices   map[ID]*Device
	radios    map[ID]*WifiRadio
	nets      map[ID]*WifiNet
	vlans     map[ID]*VLAN
	layouts   map[ID]*Layout
	laynodes  map[ID]*LayNode
	laylinks  map[ID]*LayLink
	upgrades  map[ID]*Upgrade
	jobs      map[ID]*UpgradeJob
	alerts    map[ID]*Alert
	instance  Instance
	observers []Observer
}

// NewStore returns an empty store.
func NewStore() *Store {
	return &Store{
		devices:  map[ID]*Device{},
		radios:   map[ID]*WifiRadio{},
		nets:     map[ID]*WifiNet{},
		vlans:    map[ID]*VLAN{},
		layouts:  map[ID]*Layout{},
		laynodes: map[ID]*LayNode{},
		laylinks: map[ID]*LayLink{},
		upgrades: map[ID]*Upgrade{},
		jobs:     map[ID]*UpgradeJob{},
		alerts:   map[ID]*Alert{},
	}
}

// Watch registers an Observer for all store events.
func (s *Store) Watch(o Observer) {
	s.mu.Lock()
	s.observers = append(s.observers, o)
	s.mu.Unlock()
}

// NewID allocates a fresh, monotonically increasing object ID.
func (s *Store) NewID() ID { return ID(atomic.AddUint32(&s.nextID, 1)) }

func (s *Store) emit(t EventType, c Collection, id ID, obj any) {
	// observers snapshot taken under lock by callers; emit is called after the
	// map mutation while still holding the write lock is avoided — see Put*.
	for _, o := range s.observers {
		o(Event{Type: t, Collection: c, ID: id, Obj: obj})
	}
}

// --- Devices ---------------------------------------------------------------

// PutDevice inserts or replaces a device and emits the matching event.
func (s *Store) PutDevice(d *Device) {
	s.mu.Lock()
	_, existed := s.devices[d.ID]
	cp := *d
	s.devices[d.ID] = &cp
	obs := append([]Observer(nil), s.observers...)
	s.mu.Unlock()
	t := Created
	if existed {
		t = Updated
	}
	for _, o := range obs {
		o(Event{Type: t, Collection: ColDevice, ID: d.ID, Obj: cp})
	}
}

// Device returns a copy of the device with id, if present.
func (s *Store) Device(id ID) (Device, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	d, ok := s.devices[id]
	if !ok {
		return Device{}, false
	}
	return *d, true
}

// DeviceByMAC finds a device by its enrollment MAC.
func (s *Store) DeviceByMAC(mac string) (Device, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, d := range s.devices {
		if d.MACString() == mac {
			return *d, true
		}
	}
	return Device{}, false
}

// Devices returns a snapshot slice of all devices.
func (s *Store) Devices() []Device {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Device, 0, len(s.devices))
	for _, d := range s.devices {
		out = append(out, *d)
	}
	return out
}

// DeleteDevice removes a device and emits a Deleted event.
func (s *Store) DeleteDevice(id ID) {
	s.mu.Lock()
	d, ok := s.devices[id]
	if ok {
		delete(s.devices, id)
	}
	obs := append([]Observer(nil), s.observers...)
	var last Device
	if ok {
		last = *d
	}
	s.mu.Unlock()
	if ok {
		for _, o := range obs {
			o(Event{Type: Deleted, Collection: ColDevice, ID: id, Obj: last})
		}
	}
}

// --- Generic helpers for the remaining collections -------------------------
// (kept simple and uniform; extend as the integrating project needs.)

// PutWifiNet inserts or replaces a Wi-Fi network.
func (s *Store) PutWifiNet(n *WifiNet) {
	s.put(ColWifiNet, n.ID, func() bool { _, e := s.nets[n.ID]; cp := *n; s.nets[n.ID] = &cp; return e }, func() any { return *n })
}

// PutWifiRadio inserts or replaces a radio.
func (s *Store) PutWifiRadio(r *WifiRadio) {
	s.put(ColWifiRadio, r.ID, func() bool { _, e := s.radios[r.ID]; cp := *r; s.radios[r.ID] = &cp; return e }, func() any { return *r })
}

// PutVLAN inserts or replaces a VLAN.
func (s *Store) PutVLAN(v *VLAN) {
	s.put(ColVLAN, v.ID, func() bool { _, e := s.vlans[v.ID]; cp := *v; s.vlans[v.ID] = &cp; return e }, func() any { return *v })
}

// PutUpgradeJob inserts or replaces an upgrade job.
func (s *Store) PutUpgradeJob(j *UpgradeJob) {
	s.put(ColUpgradeJob, j.ID, func() bool { _, e := s.jobs[j.ID]; cp := *j; s.jobs[j.ID] = &cp; return e }, func() any { return *j })
}

// PutLayNode inserts or replaces a topology node.
func (s *Store) PutLayNode(n *LayNode) {
	s.put(ColLayNode, n.ID, func() bool { _, e := s.laynodes[n.ID]; cp := *n; s.laynodes[n.ID] = &cp; return e }, func() any { return *n })
}

// PutLayLink inserts or replaces a topology link.
func (s *Store) PutLayLink(l *LayLink) {
	s.put(ColLayLink, l.ID, func() bool { _, e := s.laylinks[l.ID]; cp := *l; s.laylinks[l.ID] = &cp; return e }, func() any { return *l })
}

// AddAlert records an alert and emits a Created event.
func (s *Store) AddAlert(a *Alert) {
	s.put(ColAlert, a.ID, func() bool { cp := *a; s.alerts[a.ID] = &cp; return false }, func() any { return *a })
}

// put centralizes the lock/mutate/emit dance for the uniform collections.
func (s *Store) put(c Collection, id ID, mutate func() bool, snap func() any) {
	s.mu.Lock()
	existed := mutate()
	obj := snap()
	obs := append([]Observer(nil), s.observers...)
	s.mu.Unlock()
	t := Created
	if existed {
		t = Updated
	}
	for _, o := range obs {
		o(Event{Type: t, Collection: c, ID: id, Obj: obj})
	}
}

// Instance returns the controller's own configuration.
func (s *Store) Instance() Instance {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.instance
}

// SetInstance updates the controller configuration.
func (s *Store) SetInstance(i Instance) {
	s.mu.Lock()
	s.instance = i
	s.mu.Unlock()
}
