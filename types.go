package main

import (
	"encoding/xml"
)

type mMTConnectStreams struct {
	XMLName xml.Name                `xml:"MTConnectStreams"`
	Header  mMTConnectStreamsHeader `xml:"Header"`
	Streams []mStream               `xml:"Streams>DeviceStream"`
}

type mMTConnectStreamsHeader struct {
	CreationTime  string `xml:"creationTime,attr"`
	Sender        string `xml:"sender,attr"`
	InstanceID    string `xml:"instanceId,attr"`
	Version       string `xml:"version,attr"`
	BufferSize    int    `xml:"bufferSize,attr"`
	NextSequence  int    `xml:"nextSequence,attr"`
	FirstSequence int    `xml:"firstSequence,attr"`
	LastSequence  int    `xml:"lastSequence,attr"`
}

type mStream struct {
	XMLName    xml.Name           `xml:"DeviceStream"`
	DeviceName string             `xml:"name,attr"`
	Components []mComponentStream `xml:"ComponentStream"`
}

type mComponentStream struct {
	XMLName     xml.Name `xml:"ComponentStream"`
	Component   string   `xml:"component,attr"`
	Name        string   `xml:"name,attr"`
	ComponentID string   `xml:"componentId,attr"`
	Samples     mItems   `xml:"Samples"`
	Events      mItems   `xml:"Events"`
	Condition   mItems   `xml:"Condition"`
}

type mItems struct {
	XMLName xml.Name
	Items   []mComponentSample
}

func (e *mItems) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var items []mComponentSample
	var done bool
	for !done {
		t, err := d.Token()
		if err != nil {
			return err
		}
		switch t := t.(type) {
		case xml.StartElement:
			e := mComponentSample{}
			d.DecodeElement(&e, &t)
			items = append(items, e)
		case xml.EndElement:
			done = true
		}
	}
	e.XMLName = start.Name
	e.Items = items
	return nil
}

type mComponentSample struct {
	XMLName    xml.Name
	DataItemID string     `xml:"dataItemId,attr"`
	Name       string     `xml:"name,attr"`
	Timestamp  string     `xml:"timestamp,attr"`
	Sequence   int        `xml:"sequence,attr"`
	Attrs      []xml.Attr `xml:",any,attr"`
	Value      string     `xml:",chardata"`
}

type mMTConnectDevices struct {
	XMLName xml.Name                `xml:"MTConnectDevices"`
	Header  mMTConnectDevicesHeader `xml:"Header"`
	Devices []mDevice               `xml:"Devices>Device"`
}

type mMTConnectDevicesHeader struct {
	CreationTime    string `xml:"creationTime,attr"`
	Sender          string `xml:"sender,attr"`
	InstanceID      string `xml:"instanceId,attr"`
	Version         string `xml:"version,attr"`
	AssetBufferSize int    `xml:"assetBufferSize"`
	AssetCount      int    `xml:"assetCount,attr"`
	BufferSize      int    `xml:"bufferSize"`
}

type mDevice struct {
	XMLName     xml.Name     `xml:"Device"`
	ID          string       `xml:"id,attr"`
	Name        string       `xml:"name,attr"`
	UUID        string       `xml:"uuid,attr"`
	Description mDescription `xml:"Description"`
	Components  mComponents  `xml:"Components"`
	DataItems   []mDataItem  `xml:"DataItems>DataItem"`
}

type mDataItem struct {
	XMLName xml.Name   `xml:"DataItem"`
	Attrs   []xml.Attr `xml:",any,attr"`
}

type mComponents struct {
	XMLName xml.Name
	Items   []mComponent
}

type mComponent struct {
	XMLName   xml.Name
	ID        string      `xml:"id,attr"`
	Name      string      `xml:"name,attr"`
	DataItems []mDataItem `xml:"DataItems>DataItem"`
}

func (e *mComponents) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var items []mComponent
	var done bool
	for !done {
		t, err := d.Token()
		if err != nil {
			return err
		}
		switch t := t.(type) {
		case xml.StartElement:
			e := mComponent{}
			d.DecodeElement(&e, &t)
			items = append(items, e)
		case xml.EndElement:
			done = true
		}
	}
	e.XMLName = start.Name
	e.Items = items
	return nil
}

type mDescription struct {
	Manufacturer string `xml:"manufacturer,attr"`
	Model        string `xml:"model,attr"`
	Description  string `xml:",chardata"`
}

// execution enum: READY, ACTIVE, INTERRUPTED, FEED_HOLD, STOPPED, OPTIONAL_STOP, PROGRAM_STOPPED, or PROGRAM_COMPLETED

type mMTConnectError struct {
	Header mMTConnectErrorHeader  `xml:"Header"`
	Errors []mMTConnectErrorError `xml:"Errors>Error"`
}

type mMTConnectErrorError struct {
	ErrorCode string `xml:"errorCode,attr"`
	Value     string `xml:",chardata"`
}

type mMTConnectErrorHeader struct {
	CreationTime string `xml:"creationTime,attr"`
	Sender       string `xml:"sender,attr"`
	InstanceID   string `xml:"instanceId,attr"`
	Version      string `xml:"version,attr"`
	BufferSize   int    `xml:"bufferSize,attr"`
}
