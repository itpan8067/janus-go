package janus

import (
	"testing"
	"time"
	"fmt"
	"github.com/pion/webrtc/v4"
)

func Test_Connect(t *testing.T) {

	client, err := Connect("ws://159.138.5.73:8188")
	if err != nil {
		t.Fail()
		return
	}

	session, err := client.Create()
	if err != nil {
		t.Fail()
		return
	}
	t.Log("\n\n")
	handle, err := session.Attach("janus.plugin.videoroom")
	if err != nil {
		t.Fail()
		return
	}
	t.Log("\n\n")
	
	go func() {
		for {
			if _, keepAliveErr := session.KeepAlive(); keepAliveErr != nil {
				//panic(keepAliveErr)
			}

			time.Sleep(5 * time.Second)
		}
	}()
	
	_, err = handle.Message(map[string]interface{}{
		"request": "join",
		"ptype":   "publisher",
		"room":    1234,
		"display": "2222",
	}, nil)
	if err != nil {
		t.Fail()
		return
	}
	t.Log("\n\n")
	
	
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	}
	peerConnection, err := webrtc.NewPeerConnection(config)
	if err != nil {
		t.Fail()
		return
	}
	peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		fmt.Printf("ICE Connection State has changed: %s\n", connectionState.String())
	})
	vp8Track, err := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: "video/h264"}, "video", "pion")
	if err != nil {
		t.Fail()
		return
	}
	_, err = peerConnection.AddTrack(vp8Track)
	if err != nil {
		t.Fail()
		return
	}
	offer, err := peerConnection.CreateOffer(nil)
	if err != nil {
		t.Fail()
		return
	}
	gatherComplete := webrtc.GatheringCompletePromise(peerConnection)
	if err = peerConnection.SetLocalDescription(offer); err != nil {
		t.Fail()
		return
	}
	<-gatherComplete
	

	msg, err := handle.Message(map[string]interface{}{
		"request": "publish",
		"audio":   false,
		"video":   true,
		"data":    false,
	}, map[string]interface{}{
		"type":    "offer",
		"sdp":     peerConnection.LocalDescription().SDP,
		"trickle": false,
	})
	if err != nil {
		t.Fail()
		return
	}
	
	if msg.Jsep != nil {
		err = peerConnection.SetRemoteDescription(webrtc.SessionDescription{
			Type: webrtc.SDPTypeAnswer,
			SDP:  msg.Jsep["sdp"].(string),
		})
		if err != nil {
			t.Fail()
			return
		}
	}
	t.Log("\n\n")
	time.Sleep(10 * time.Second) // 等待 10 秒钟
}
