package relayservice

import "github.com/prometheus/client_golang/prometheus"

var mLoopLatency = prometheus.NewHistogram(
	prometheus.HistogramOpts{
		Name: "enet_loop_latency",
		Help: "Latency of the loop calling enet host service. Note: this represents application loop and not the service call itself.",
	},
)

var mTotalBytesSent = prometheus.NewCounter(
	prometheus.CounterOpts{
		Name: "enet_total_bytes_sent",
		Help: "Total bytes sent as reported by enet.",
	},
)

var mTotalBytesReceived = prometheus.NewCounter(
	prometheus.CounterOpts{
		Name: "enet_total_bytes_received",
		Help: "Total bytes received as reported by enet.",
	},
)

var mTotalPacketsSent = prometheus.NewCounter(
	prometheus.CounterOpts{
		Name: "enet_total_packets_sent",
		Help: "Total packets sent as reported by enet.",
	},
)

var mTotalPacketsReceived = prometheus.NewCounter(
	prometheus.CounterOpts{
		Name: "enet_total_packets_received",
		Help: "Total packets received as reported by enet.",
	},
)

var mConnectEvent = prometheus.NewCounter(
	prometheus.CounterOpts{
		Name: "enet_connect",
		Help: "Enet connect events received.",
	},
)

var mDisconnectEvent = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "enet_disconnect",
		Help: "Enet disconnect events received.",
	},
	[]string{"type"},
)

var mRelayPayload = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "relay_payload",
		Help: "Payloads received by mitsuko relay.",
	},
	[]string{"type"},
)

func registerMetrics() {
	prometheus.MustRegister(
		mLoopLatency,
		mTotalBytesSent,
		mTotalBytesReceived,
		mTotalPacketsSent,
		mTotalPacketsReceived,
		mConnectEvent,
		mDisconnectEvent,
		mRelayPayload,
	)
}
