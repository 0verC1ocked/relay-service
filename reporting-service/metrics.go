package reportingservice

import "github.com/prometheus/client_golang/prometheus"

var mPlayerCount = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Name: "mitsuko_player_count",
		Help: "Player count as reported by mitsuko gameserver.",
	},
)

var mMatchCount = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Name: "mitsuko_match_count",
		Help: "Match count as reported by mitsuko gameserver.",
	},
)

func registerMetrics() {
	prometheus.MustRegister(
		mMatchCount,
		mPlayerCount,
	)
}
