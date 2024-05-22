package reportingservice

import (
	"errors"
	"log"
	mitsuko "mitsuko-relay/lib/payloadbuilder/src/proto/pb/mitsuko/relay"
	"time"

	"agones.dev/agones/pkg/sdk"
	agones "agones.dev/agones/sdks/go"
)

type Reporter struct {
	SystemChan      chan *mitsuko.SystemMessage
	agonesSdk       *agones.SDK
	lastPlayerCount int
	lastMatchCount  int
	startedAt       time.Time
	isOld           bool
}

func init() {
	registerMetrics()
}

func (r *Reporter) Start(agonesEnabled bool) {
	if agonesEnabled {
		r.initAgones()
	}
	for sysmsg := range r.SystemChan {
		switch sysmsg.Event {
		case mitsuko.SystemEvent_HEALTH:
			if agonesEnabled {
				go r.doHealth()
			}
		case mitsuko.SystemEvent_METRIC:
			r.lastPlayerCount = int(sysmsg.Gameserver.PlayerCount)
			r.lastMatchCount = int(sysmsg.Gameserver.MatchCount)
			mPlayerCount.Set(float64(r.lastPlayerCount))
			mMatchCount.Set(float64(r.lastMatchCount))
		}
	}
}

func (r *Reporter) initAgones() {
	r.startedAt = time.Now()
	r.isOld = false
	s, err := agones.NewSDK()
	if err != nil {
		log.Fatal(err)
	}
	r.agonesSdk = s
	if gs, err := r.agonesSdk.GameServer(); err != nil {
		log.Print(err)
	} else {
		log.Printf("%v", gs)
	}
	if err := r.agonesSdk.Ready(); err != nil {
		log.Fatal(err)
	}
	r.agonesSdk.WatchGameServer(func(gs *sdk.GameServer) {
		r.watchCb(gs)
	})
	go r.startCounterReports()
	go r.watchAgeAndShutdown(6 * 60 * 60)
	err = s.SetLabel("gs-session-ready", "true")
	if err != nil {
		log.Print(err)
	}
}

func (r *Reporter) UpdateCounter(update int32) error {
	var under_cap bool
	var err error
	if update > 0 {
		under_cap, err = r.agonesSdk.Alpha().IncrementCounter("player_count", int64(update))
	} else if update < 0 {
		under_cap, err = r.agonesSdk.Alpha().DecrementCounter("player_count", int64(update*-1))
	}
	if !under_cap {
		return errors.New("reached max capacity")
	}
	if err != nil {
		return err
	}
	return nil
}

func (r *Reporter) startHealth() {
	for range time.Tick(time.Second) {
		go r.doHealth()
	}
}

func (r *Reporter) watchCb(gs *sdk.GameServer) {
	if gs.ObjectMeta.Labels["hws/mitsuko-deploy-generation"] == "old" {
		r.agonesSdk.Alpha().SetCounterCapacity("player_count", 0)
	}
}

func (r *Reporter) watchAgeAndShutdown(seconds float64) {
	continousEmpty := 0
	for range time.Tick(time.Second * 10) {
		if r.lastPlayerCount > 0 {
			continousEmpty = 0
		} else {
			continousEmpty++
			if continousEmpty >= 5 {
				if err := r.agonesSdk.Ready(); err != nil {
					log.Println("error while marking ready in timeout", err)
				}
			}
		}
		if time.Since(r.startedAt).Seconds() < seconds {
			continue
		} else {
			if err := r.agonesSdk.SetLabel("gs-session-ready", "false"); err != nil {
				log.Println("error while setting g session in time out", err)
			}
			if _, err := r.agonesSdk.Alpha().SetCounterCapacity("player_count", 0); err != nil {
				log.Println("error while removing capacity")
			}
		}
		log.Println("gs at timeout")
		if continousEmpty >= 5 {
			log.Println("gs shutdown timeout")
			r.agonesSdk.Shutdown()
		}
	}
}

func (r *Reporter) doHealth() {
	if err := r.agonesSdk.Health(); err != nil {
		log.Println("Agones health failed", err)
	}
}

func (r *Reporter) startCounterReports() {
	for range time.Tick(time.Second * 5) {
		var count int64 = 0
		if r.lastPlayerCount >= 0 {
			count = int64(r.lastPlayerCount)
		} else {
			log.Println("Invalid player count:", r.lastPlayerCount)
		}
		r.agonesSdk.Alpha().SetCounterCount("player_count", count)
	}
}
