package main

import (
	"flag"
	"fmt"
	"log"
	forwardingservice "mitsuko-relay/forwarding-service"
	httpserver "mitsuko-relay/http-server"
	relayservice "mitsuko-relay/relay-service"
	reportingservice "mitsuko-relay/reporting-service"
	streamingservice "mitsuko-relay/streaming-service"
	"os"
	mitsuko "mitsuko-relay/lib/payloadbuilder/src/proto/pb/mitsuko/relay"
	"strings"
)

const (
	SERVER_PORT = 42220
	RELAY_PORT  = 7777
)

func main() {
	printBuildInfo()
	serverHost := flag.String("shost", "", "server host to bind to")
	relayHost := flag.String("rhost", "127.0.0.1", "relay host")
	seedsStr := flag.String("rpSeeds", "127.0.0.1:9092", "redpanda seeds for streaming service, comma seperated")
	fwdUrlBase := flag.String("fwdIp", "http://127.0.0.1:6363", "HTTP forward url base")
	enableGSReporter := flag.Bool("gsreport", false, "Enables game server reporter for agones. NOTE: only use inside agones fleet.")
	flag.Parse()

	seeds := strings.Split(*seedsStr, ",")

	rec := make(chan []byte, 50000)
	fwc := make(chan *mitsuko.RelayPayload, 50000)
	sysc := make(chan *mitsuko.SystemMessage, 50000)
	strmc := make(chan *streamingservice.StreamPayload, 50000)

	gsReporter := &reportingservice.Reporter{
		SystemChan: sysc,
	}

	relay := &relayservice.RelayService{
		Host: *relayHost,
		Port: RELAY_PORT,
	}

	log.Println("Starting gameserver reporter for agones")
	go gsReporter.Start(*enableGSReporter)

	if err := relay.Start(); err != nil {
		log.Fatal("Unable to start relay service", err.Error())
	}
	defer relay.Stop()

	log.Println("Running relay channel on ", *relayHost, RELAY_PORT)
	go relay.Run(10, rec, fwc, strmc, sysc)

	log.Println("Running forwarding service")
	go forwardingservice.Run(*fwdUrlBase, fwc)

	log.Println("Running streaming service")
	go streamingservice.Run(seeds, strmc)

	log.Println("Starting mitsuko relay server on ", *serverHost, SERVER_PORT)
	httpserver.StartServer(fmt.Sprintf("%s:%d", *serverHost, SERVER_PORT), rec)
}

func printBuildInfo() {
	log.Println("release: ", os.Getenv("MITSUKO_RELEASE"))
}
