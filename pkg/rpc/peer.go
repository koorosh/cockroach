// Copyright 2023 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package rpc

import (
	"context"
	"time"

	"github.com/cockroachdb/cockroach/pkg/util/circuit"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/redact"
)

// A peer is a remote node that we are trying to maintain a healthy RPC
// connection (for a given connection class not known to the peer itself)
// to. If no healthy connection exists, the peer tracks the time of
// disconnect and maintains a circuit breaker that resets once the connection
// can be re-established.
type peer struct {
	// b maintains connection health. This breaker's async probe is always
	// active - it is the heartbeat loop and manages `mu.c.` (including
	// recreating it after the connection fails and has to be redialed).
	//
	// NB: at the time of writing, we don't use the breaking capabilities,
	// i.e. we don't check the circuit breaker in `Connect`. We will do that
	// once the circuit breaker is mature, and then retire the breakers
	// returned by Context.getBreaker.
	//
	// Currently what will happen when a peer is down is that `c` will be
	// recreated (blocking new callers to `Connect()`), a connection attempt
	// will be made, and callers will see the failure to this attempt.
	//
	// With the breaker, callers would be turned away eagerly until there
	// is a known-healthy connection.
	//
	// mu must *NOT* be held while operating on `b`. This is because the async
	// probe will sometimes have to synchronously acquire mu before spawning off.
	b  *circuit.Breaker
	nm NodeMetrics
	mu struct {
		syncutil.Mutex
		// Copies of PeerSnap may be leaked outside of lock, since the memory within
		// is never mutated in place.
		PeerSnap
	}
}

// PeerSnap is the state of a peer.
type PeerSnap struct {
	c *Connection // never nil, only mutated in the breaker probe
	// disconnected is zero initially, reset on successful heartbeat, set on
	// heartbeat teardown if zero. In other words, does not move forward across
	// subsequent connection failures - it tracks the first disconnect since
	// having been healthy.
	//
	// NB: this field has no bearing on whether connections are returned to
	// callers.
	disconnected time.Time
	// INVARIANT: decommissionedOrSuperseded never transitions from true to false.
	decommissionedOrSuperseded bool
}

func (p *peer) snap() PeerSnap {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.mu.PeerSnap
}

// newPeer returns circuit breaker that trips when connection (associated
// with provided ConnKey) is failed. The breaker's probe *is* the heartbeat loop
// and is thus running at all times. The exception is a decommissioned node, for
// which the probe simply exits (any future connection attempts to the same peer
// will trigger the probe but the probe will exit again), and a superseded peer,
// i.e. one for which a node restarted with a different IP address and we're the
// "old", unhealth, peer.
func (rpcCtx *Context) newPeer(k ConnKey) *peer {
	// Initialization here is a bit circular. The peer holds the breaker. The
	// breaker probe references the peer because it needs to replace the one-shot
	// Connection when it makes a new connection in the probe. And (all but the
	// first incarnation of) the Connection also holds on to the breaker since the
	// Connect method needs to do the short-circuiting (if a Connection is created
	// while the breaker is tripped, we want to block in Connect only once we've
	// seen the first heartbeat succeed).
	p := &peer{
		// NB: we currently don't refcount the node metrics and instead leak them.
		//
		// Multiple peers to a given node can temporarily exist at any given point
		// in time (if the node restarts under a different IP). We assume that
		// ultimately one of those will become unhealthy and repeatedly fail its
		// probe. On probe failure, we check the node map for duplicates and if a
		// healthy duplicate exists, remove ourselves. If we ever add refcounting,
		// we need to update this mechanism to decrease the refcount.
		nm: rpcCtx.metrics.loadNodeMetrics(k.NodeID),
	}
	var b *circuit.Breaker
	probe := breakerProbe{
		peer:              p,
		k:                 k,
		heartbeatInterval: rpcCtx.heartbeatInterval,
		stopper:           rpcCtx.Stopper,
		runHeartbeatUntilFailure: func(ctx context.Context, conn *Connection, k ConnKey, healBreaker func(), nm NodeMetrics) error {
			return rpcCtx.runHeartbeatUntilFailure(ctx, p, healBreaker)
		},
	}

	ctx := rpcCtx.makeDialCtx(k.TargetAddr, k.NodeID, k.Class)
	b = circuit.NewBreaker(circuit.Options{
		Name: "breaker", // log tags already represent `k`
		AsyncProbe: func(report func(error), done func()) {
			probe.launch(ctx, report, done)
		},
		EventHandler: &circuitBreakerLogger{wrapped: &circuit.EventLogger{
			Log: func(buf redact.StringBuilder) {
				log.Health.InfofDepth(ctx, 6, "%s", buf)
			},
		}},
	})
	p.b = b
	// The first connection attempt is special since the breaker will be the
	// breaker will be tripped initially (as the probe is the heartbeat loop) but
	// we still want callers to block on the attempt. So we set up sigFn to return
	// a dummy signal until we've succeeded (or failed) the first heartbeat.
	c := newConnectionToNodeID(k, nil /* sigFn */)
	c.sigFn = func() circuit.Signal {
		select {
		case <-c.initialHeartbeatDone:
			return b.Signal()
		default:
			return neverTripSignal
		}
	}
	p.mu.PeerSnap = PeerSnap{c: c}

	return p
}
