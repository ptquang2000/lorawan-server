package main

import (
	"crypto/aes"
	"fmt"
	"sync"

	"github.com/brocaar/lorawan"
	"github.com/pkg/errors"
)

type FakeEndDevice struct {
	nwkSKey   [16]byte
	appSKey   [16]byte
	appKey    [16]byte
	devEui    [8]byte
	devAddr   lorawan.DevAddr
	devNonce  lorawan.DevNonce
	joinNonce lorawan.JoinNonce
	FCntUp    uint16
	FCntDown  uint16
	FrameChan chan []byte
	JaWait    sync.WaitGroup
}

var joinEUI = [8]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}

func (device *FakeEndDevice) startFlow() {
	for payload := range device.FrameChan {
		var phy lorawan.PHYPayload
		if err := phy.UnmarshalBinary(payload); err != nil {
			panic(err)
		}

		var success bool
		switch phy.MHDR.MType {
		case lorawan.JoinAccept:
			success = device.joinAcceptHandler(&phy)
		case lorawan.ConfirmedDataDown:
			success = device.confirmedDownlinkHandler(&phy)
		case lorawan.UnconfirmedDataDown:
			success = device.unconfirmedDownlinkHandler(&phy)
		}

		if success {
			phyJSON, err := phy.MarshalJSON()
			if err != nil {
				panic(err)
			}
			fmt.Println(string(phyJSON))
		}
	}
}

func (device *FakeEndDevice) confirmedDownlinkHandler(phy *lorawan.PHYPayload) bool {
	return true
}

func (device *FakeEndDevice) unconfirmedDownlinkHandler(phy *lorawan.PHYPayload) bool {
	res, err := phy.ValidateDownlinkDataMIC(lorawan.LoRaWAN1_0, 0, device.nwkSKey)
	if !res || err != nil {
		return false
	}

	macPL, ok := phy.MACPayload.(*lorawan.MACPayload)
	if !ok {
		panic("Payload must be a *MACPayload")
	}

	if macPL.FPort == nil {
		return true
	}

	if *macPL.FPort == 0 {
		if err := phy.DecryptFRMPayload(device.nwkSKey); err != nil {
			panic("Decrypt FRMPayload to mac commands fail")
		}
	} else {
		if err := phy.DecryptFRMPayload(device.appSKey); err != nil {
			panic("Decrypt FRMPayload to data payload fail")
		}

		pl, ok := macPL.FRMPayload[0].(*lorawan.DataPayload)
		if !ok {
			panic("*FRMPayload must be DataPayload")
		}
		fmt.Println(pl.Bytes)
	}

	return true
}

func (device *FakeEndDevice) joinAcceptHandler(phy *lorawan.PHYPayload) bool {
	err_ := phy.DecryptJoinAcceptPayload(device.appKey)
	if err_ != nil {
		panic(err_)
	}

	res, err := phy.ValidateDownlinkJoinMIC(lorawan.JoinRequestType, joinEUI, lorawan.DevNonce(device.joinNonce+1), device.appKey)
	if !res || err != nil {
		return false
	}

	jaPL, ok := phy.MACPayload.(*lorawan.JoinAcceptPayload)
	if !ok {
		panic("*JoinAcceptPayload expected")
	}

	device.devAddr = jaPL.DevAddr

	appSKey, err := getSKey(
		false,
		0x02,
		device.appKey,
		jaPL.HomeNetID,
		joinEUI,
		jaPL.JoinNonce,
		device.devNonce)
	if err != nil {
		panic(err)
	}
	device.appSKey = appSKey

	nwkSKey, err := getSKey(
		false,
		0x01,
		device.appKey,
		jaPL.HomeNetID,
		joinEUI,
		jaPL.JoinNonce,
		device.devNonce)
	if err != nil {
		panic(err)
	}
	device.nwkSKey = nwkSKey

	device.FCntUp = 0
	device.FCntDown = 0

	device.joinNonce = jaPL.JoinNonce
	device.JaWait.Done()

	return true
}

func (dev *FakeEndDevice) sendJr(gateway *FakeGateway) {
	phy := lorawan.PHYPayload{
		MHDR: lorawan.MHDR{
			MType: lorawan.JoinRequest,
			Major: lorawan.LoRaWANR1,
		},
		MACPayload: &lorawan.JoinRequestPayload{
			JoinEUI:  joinEUI,
			DevEUI:   dev.devEui,
			DevNonce: dev.devNonce,
		},
	}

	if err := phy.SetUplinkJoinMIC(dev.appKey); err != nil {
		panic(err)
	}

	frame, err := phy.MarshalBinary()
	if err != nil {
		panic(err)
	}
	gateway.JrChan <- frame
}

func (dev *FakeEndDevice) sendUnconfirmedUl(gateway *FakeGateway) {
	nwkSKey := dev.nwkSKey
	appSKey := dev.appSKey
	fPort := uint8(10)

	dataPL := "Hello"

	phy := lorawan.PHYPayload{
		MHDR: lorawan.MHDR{
			MType: lorawan.UnconfirmedDataUp,
			Major: lorawan.Major(lorawan.LoRaWAN1_0),
		},
		MACPayload: &lorawan.MACPayload{
			FHDR: lorawan.FHDR{
				DevAddr: lorawan.DevAddr(dev.devAddr),
				FCtrl: lorawan.FCtrl{
					ADR:       false,
					ADRACKReq: false,
					ACK:       false,
				},
				FCnt: uint32(dev.FCntUp),
			},
			FPort:      &fPort,
			FRMPayload: []lorawan.Payload{&lorawan.DataPayload{Bytes: []byte(dataPL)}},
		},
	}

	if err := phy.EncryptFRMPayload(appSKey); err != nil {
		panic(err)
	}

	if err := phy.SetUplinkDataMIC(lorawan.LoRaWAN1_0, 0, 0, 0, nwkSKey, lorawan.AES128Key{}); err != nil {
		panic(err)
	}

	frame, err := phy.MarshalBinary()
	if err != nil {
		panic(err)
	}
	gateway.UlChan <- frame
}

func (dev *FakeEndDevice) sendConfirmedUl(gateway *FakeGateway) {
	nwkSKey := dev.nwkSKey
	appSKey := dev.appSKey
	fPort := uint8(10)

	dataPL := []byte{9}

	phy := lorawan.PHYPayload{
		MHDR: lorawan.MHDR{
			MType: lorawan.ConfirmedDataUp,
			Major: lorawan.Major(lorawan.LoRaWAN1_0),
		},
		MACPayload: &lorawan.MACPayload{
			FHDR: lorawan.FHDR{
				DevAddr: lorawan.DevAddr(dev.devAddr),
				FCtrl: lorawan.FCtrl{
					ADR:       false,
					ADRACKReq: false,
					ACK:       false,
				},
				FCnt: uint32(dev.FCntUp),
			},
			FPort:      &fPort,
			FRMPayload: []lorawan.Payload{&lorawan.DataPayload{Bytes: []byte(dataPL)}},
		},
	}

	if err := phy.EncryptFRMPayload(appSKey); err != nil {
		panic(err)
	}

	if err := phy.SetUplinkDataMIC(lorawan.LoRaWAN1_0, 0, 0, 0, nwkSKey, lorawan.AES128Key{}); err != nil {
		panic(err)
	}

	frame, err := phy.MarshalBinary()
	if err != nil {
		panic(err)
	}
	gateway.UlChan <- frame
}

func setUpDev() (fakeDevices []*FakeEndDevice) {
	fakeDevices = append(fakeDevices, &FakeEndDevice{
		appKey:    [16]byte{0x96, 0xf4, 0x15, 0x5e, 0xfa, 0x37, 0xbe, 0xb7, 0x60, 0x5e, 0x4d, 0x5d, 0x6d, 0x65, 0x64, 0xe8},
		devEui:    [8]byte{0xAA, 0xAA, 0x0A, 0x00, 0x00, 0xFF, 0xFF, 0xFE},
		devNonce:  0,
		FrameChan: make(chan []byte),
	})

	fakeDevices = append(fakeDevices, &FakeEndDevice{
		appKey:    [16]byte{0xb6, 0xd5, 0xc0, 0x0a, 0xd6, 0xfa, 0x96, 0x4d, 0x4b, 0xa9, 0xa0, 0x16, 0x52, 0x64, 0x7b, 0x93},
		devEui:    [8]byte{0xBB, 0xBB, 0x0B, 0x00, 0x00, 0xFF, 0xFF, 0xFE},
		devNonce:  0,
		FrameChan: make(chan []byte),
	})
	return
}

func getSKey(optNeg bool, typ byte, nwkKey lorawan.AES128Key, netID lorawan.NetID, joinEUI lorawan.EUI64, joinNonce lorawan.JoinNonce, devNonce lorawan.DevNonce) (lorawan.AES128Key, error) {
	var key lorawan.AES128Key
	b := make([]byte, 16)
	b[0] = typ

	netIDB, err := netID.MarshalBinary()
	if err != nil {
		return key, errors.Wrap(err, "marshal binary error")
	}

	joinEUIB, err := joinEUI.MarshalBinary()
	if err != nil {
		return key, errors.Wrap(err, "marshal binary error")
	}

	joinNonceB, err := joinNonce.MarshalBinary()
	if err != nil {
		return key, errors.Wrap(err, "marshal binary error")
	}

	devNonceB, err := devNonce.MarshalBinary()
	if err != nil {
		return key, errors.Wrap(err, "marshal binary error")
	}

	if optNeg {
		copy(b[1:4], joinNonceB)
		copy(b[4:12], joinEUIB)
		copy(b[12:14], devNonceB)
	} else {
		copy(b[1:4], joinNonceB)
		copy(b[4:7], netIDB)
		copy(b[7:9], devNonceB)
	}

	block, err := aes.NewCipher(nwkKey[:])
	if err != nil {
		return key, err
	}
	if block.BlockSize() != len(b) {
		return key, fmt.Errorf("block-size of %d bytes is expected", len(b))
	}
	block.Encrypt(key[:], b)

	return key, nil
}
