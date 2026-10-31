// Command cmr-ctrl is a pure-Go CMR controller that a real MikroTik CMR device
// pairs and connects to, with no MikroTik/Nova controller in the path. It is a
// thin wrapper over pkg/controller — the same logic is embeddable in other
// programs; see that package's docs and examples/embed.
//
// It implements the OPEN pairing flow (controller pairing-requirement=none):
// TLS 1.2 on tcp/54321 with a self-signed Ed25519 identity, the nv/m3 control
// plane bring-up, and an id10 SUCCESS that promotes the device to "paired" — no
// PKEX / password handshake.
package main

import (
	"encoding/hex"
	"flag"
	"log"
	"log/slog"
	"os"
	"strings"

	"github.com/athenanetworks/cmr-server/pkg/cmr"
	"github.com/athenanetworks/cmr-server/pkg/controller"
)

func main() {
	addr := flag.String("listen", "10.99.0.1:54321", "TLS listen address")
	keyfile := flag.String("keyfile", "", "path to persist the Ed25519 identity (stable fingerprint across restarts); empty = ephemeral")
	idle := flag.Duration("idle", 0, "per-read idle timeout before closing a connection (0 = never, hold open)")
	verbose := flag.Bool("v", false, "log every frame (Debug level) including a recursive decode")

	pushFile := flag.String("pushfile", "", "hex blob of controller root objects streamed with SUCCESS (triggers the device's state dump)")
	resyncFile := flag.String("resyncfile", "", "hex blob of id9 config objects streamed in answer to the device's getall (the controller's scope dump)")

	writeFile := flag.String("writefile", "", "hex id30 config object graph to WRITE into the device's config scope after it pairs")
	writeVLAN := flag.Int("vlan", 0, "author a CMR vlan config (vlan-ids=<n>) in-process and WRITE it; overrides -writefile when >0")
	bridgeName := flag.String("bridgename", cmr.DefaultBridgeName, "managed bridge name for -vlan")
	writeScope := flag.String("writescope", controller.DefaultWriteScope, "scope for the config write (the device's own config scope)")
	writeFlag := flag.Uint("writeflag", controller.DefaultWriteFlag, "id29 flag byte for the config-write id7")
	writeDelay := flag.Duration("writedelay", controller.DefaultWriteDelay, "delay after pairing before the config WRITE (let the device reach steady state)")

	enumScope := flag.String("enum", "", "after pairing, send an id7 getall for this scope (e.g. com.mikrotik.cmr)")
	enumFlag := flag.Uint("enumflag", 0x18, "id29 flag byte for the -enum getall")
	enumSel := flag.String("enumsel", "", "hex id30 selector for the -enum getall (empty = none)")
	subScope := flag.String("sub", "", "after pairing, send an id11 subscribe for this scope")
	subUser := flag.String("subuser", "", "optional id32 username for the -sub subscribe")
	flag.Parse()

	level := slog.LevelInfo
	if *verbose {
		level = slog.LevelDebug
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level}))

	cert, err := controller.Identity(*keyfile)
	if err != nil {
		log.Fatalf("identity: %v", err)
	}
	logger.Info("controller identity", "fingerprint", controller.Fingerprint(cert))

	cfg := controller.Config{
		Certificate: cert,
		IdleTimeout: *idle,
		Logger:      logger,
		Roots:       readHex(*pushFile),
		ConfigBurst: readHex(*resyncFile),
		WriteScope:  *writeScope,
		WriteFlag:   byte(*writeFlag),
		WriteDelay:  *writeDelay,
	}

	switch {
	case *writeVLAN > 0:
		cfg.AutoWrite = cmr.BuildVLANConfig(cmr.VLANConfig{VLANID: *writeVLAN, BridgeName: *bridgeName})
		logger.Info("authored vlan config", "vlan", *writeVLAN, "bridge", *bridgeName, "bytes", len(cfg.AutoWrite))
	case *writeFile != "":
		cfg.AutoWrite = readHex(*writeFile)
		logger.Info("loaded write payload", "bytes", len(cfg.AutoWrite))
	}

	// -enum / -sub are enumeration experiments: once the device pairs, send the
	// query so it streams its objects (reported via the default object logging).
	if *enumScope != "" || *subScope != "" {
		cfg.OnPair = func(s *controller.Session) {
			if *enumScope != "" {
				_ = s.Send(controller.GetAllQuery(*enumScope, byte(*enumFlag), mustHex(*enumSel)))
				logger.Info("sent getall", "scope", *enumScope, "flag", *enumFlag)
			}
			if *subScope != "" {
				_ = s.Send(controller.SubscribeQuery(*subScope, *subUser))
				logger.Info("sent subscribe", "scope", *subScope)
			}
		}
	}

	srv, err := controller.New(cfg)
	if err != nil {
		log.Fatalf("controller: %v", err)
	}
	if err := srv.ListenAndServe(*addr); err != nil {
		log.Fatalf("serve: %v", err)
	}
}

// readHex loads and decodes a hex blob file, or returns nil for an empty path.
func readHex(path string) []byte {
	if path == "" {
		return nil
	}
	b, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("read %s: %v", path, err)
	}
	return mustHex(strings.TrimSpace(string(b)))
}

func mustHex(s string) []byte {
	if s == "" {
		return nil
	}
	b, err := hex.DecodeString(s)
	if err != nil {
		log.Fatalf("bad hex %q: %v", s, err)
	}
	return b
}
