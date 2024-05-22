package forwardingservice

import (
	"bytes"
	"io"
	"log"
	"net/http"
	mitsuko "mitsuko-relay/lib/payloadbuilder/src/proto/pb/mitsuko/relay"
)

func Run(fwdUrlBase string, fwc chan *mitsuko.RelayPayload) {
	for payload := range fwc {
		go callApi(fwdUrlBase, payload)
	}
}

func callApi(fwdUrlBase string, payload *mitsuko.RelayPayload) {
	log.Println("Forwarding payload to", fwdUrlBase+payload.Path)
	resp, err := http.Post(
		fwdUrlBase+payload.Path,
		"application/x-protobuf",
		bytes.NewBuffer(payload.Pkt),
	)
	if err != nil {
		log.Println(err.Error())
		return
	}
	bodyBytes, _ := io.ReadAll(resp.Body)
	log.Println(string(bodyBytes))
	resp.Body.Close()
}
