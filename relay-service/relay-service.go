package relayservice

import (
	"log"
	streamingservice "mitsuko-relay/streaming-service"
	mitsuko "mitsuko-relay/lib/payloadbuilder/src/proto/pb/mitsuko/relay"
	"time"

	"github.com/TubbyStubby/go-enet-sharp"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/protobuf/proto"
)

type RelayService struct {
	Host            string
	Port            uint16
	client          enet.Host
	peer            enet.Peer
	backoffCount    uint8
	MaxBackoffCount uint8
}

func init() {
	registerMetrics()
}

func (rs *RelayService) Start() error {
	enet.Initialize()
	var err error
	rs.client, err = enet.NewHost(nil, 10, 10, 0, 0, 0)
	if err != nil {
		return err
	}
	rs.peer, err = rs.Connect(enet.NewAddress(rs.Host, rs.Port))
	if err != nil {
		return err
	}
	return nil
}

func (rs *RelayService) Connect(addr enet.Address) (enet.Peer, error) {
	peer, err := rs.client.Connect(addr, 10, 1)
	if err != nil {
		return nil, err
	}
	peer.SetTimeout(32, 500, 2000)
	peer.PingInterval(250)
	return peer, nil
}

func (rs *RelayService) Stop() {
	rs.client.Destroy()
	enet.Deinitialize()
}

func (rs *RelayService) Run(
	timeout uint32,
	rec chan []byte,
	fwc chan *mitsuko.RelayPayload,
	strmc chan *streamingservice.StreamPayload,
	sysc chan *mitsuko.SystemMessage) {
	for {
		start := time.Now()
		select {
		case b := <-rec:
			if rs.isConnected() {
				if err := rs.peer.SendBytes(b, 0, enet.PacketFlagReliable); err != nil {
					log.Println("Error while relaying bytes", err.Error())
				}
			}
		default:
			event := rs.client.Service(timeout)
			switch event.GetType() {
			case enet.EventNone:
				break
			case enet.EventConnect:
				mConnectEvent.Inc()
				log.Println("Connected to", rs.peer.GetAddress())
			case enet.EventDisconnect:
				mDisconnectEvent.With(prometheus.Labels{"type": "normal"}).Inc()
				log.Println("Disconnected, will reconnect in a moment...")
				rs.peer = nil
				rs.backoffReconnect()
			case enet.EventDisconnectTimeout:
				mDisconnectEvent.With(prometheus.Labels{"type": "timeout"}).Inc()
				log.Println("Disconnect timeout, will reconnect in a moment...")
				rs.peer = nil
				rs.backoffReconnect()
			case enet.EventReceive:
				var payload = &mitsuko.RelayPayload{}
				if err := proto.Unmarshal(event.GetPacket().GetData(), payload); err != nil {
					log.Println("Error while unmarshling relay payload", err.Error())
				} else {
					mRelayPayload.With(prometheus.Labels{"type": payload.Type.String()}).Inc()
					switch payload.Type {
					case mitsuko.RelayType_HTTP_FORWARD:
						fwc <- payload
					case mitsuko.RelayType_STREAM_FORWARD:
						strmc <- &streamingservice.StreamPayload{
							Payload: payload,
							Topic:   "raw-events",
						}
					case mitsuko.RelayType_EVENT_FORWARD:
						strmc <- &streamingservice.StreamPayload{
							Payload: payload,
							Topic:   "raw-event-history",
						}
					case mitsuko.RelayType_SYSTEM:
						sysc <- payload.GetSysMsg()
					default:
						log.Printf("Data received on unsupported channel (%d)\n", event.GetChannelID())
					}
				}
				event.GetPacket().Destroy()
			}
		}
		trackEnetHostMetrics(start, rs.client)
	}
}

func (rs *RelayService) backoffReconnect() error {
	if rs.peer != nil {
		return nil
	}
	if rs.MaxBackoffCount > 0 && rs.backoffCount > rs.MaxBackoffCount {
		panic("Unable to reconnect, max backoff reached.")
	}
	var err error
	rs.peer, err = rs.Connect(enet.NewAddress(rs.Host, rs.Port))
	if err != nil {
		rs.backoffCount++
		return err
	} else {
		return nil
	}
}

func (rs *RelayService) isConnected() bool {
	return rs.peer != nil
}

func trackEnetHostMetrics(loopStart time.Time, host enet.Host) {
	mLoopLatency.Observe(time.Since(loopStart).Seconds())
	mTotalBytesSent.Add(float64(host.GetBytesSent()))
	host.ResetBytesSent()
	mTotalBytesReceived.Add(float64(host.GetBytesReceived()))
	host.ResetBytesReceived()
	mTotalPacketsSent.Add(float64(host.GetPacketsReceived()))
	host.ResetPacketsSent()
	mTotalPacketsReceived.Add(float64(host.GetPacketsReceived()))
	host.ResetPacketsReceived()
}
