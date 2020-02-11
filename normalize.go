package main

import (
	"encoding/json"
	"time"
)

type rawDataTemp struct {
	PeerIP          string   `json:"peerIP"`
	Command         []string `json:"commands"`
	Loggedin        []string `json:"loggedin"`
	Protocol        string   `json:"protocol"`
	StartTime       string   `json:"startTime"`
	Ttylog          string   `json:"ttylog"`
	HostIP          string   `json:"hostIP"`
	PeerPort        int      `json:"peerPort"`
	Session         string   `json:"session"`
	Urls            []string `json:"url"`
	HostPort        int      `json:"hostPort"`
	Credentials     []string `json:"credentials"`
	Hashes          []string `json:"hashes"`
	EndTime         string   `json:"endTime"`
	Version         string   `json:"version"`
	UnknownCommands []string `json:"unknownCommands"`
}

type reDataTemp struct {
	Ident           string   `json:"ident"`
	Name            string   `json:"name"`
	PeerIP          string   `json:"peer_ip"`
	Command         []string `json:"commands"`
	Loggedin        []string `json:"loggedin"`
	Protocol        string   `json:"protocol"`
	StartTime       string   `json:"start_time"`
	Ttylog          string   `json:"ttylog"`
	HostIP          string   `json:"host_ip"`
	PeerPort        int      `json:"peer_port"`
	Session         string   `json:"session"`
	Urls            []string `json:"url"`
	HostPort        int      `json:"host_port"`
	Credentials     []string `json:"credentials"`
	Hashes          []string `json:"hashes"`
	EndTime         string   `json:"end_time"`
	Version         string   `json:"version"`
	UnknownCommands []string `json:"unknownCommands"`
	Timestamp       int64      `json:"timestamp"`
}

func normalize(rawData []byte) ([]byte, error) {
	data := rawDataTemp{}
	if err := json.Unmarshal(rawData, &data); err != nil {
		return nil, err
	}
	redata := reDataTemp{
		Ident:           Ident,
		Name:            NodeName,
		PeerIP:          data.PeerIP,
		Command:         data.Command,
		Loggedin:        data.Loggedin,
		Protocol:        data.Protocol,
		StartTime:       data.StartTime,
		Ttylog:          data.Ttylog,
		HostIP:          data.HostIP,
		PeerPort:        data.PeerPort,
		Session:         data.Session,
		Urls:            data.Urls,
		HostPort:        data.HostPort,
		Credentials:     data.Credentials,
		Hashes:          data.Hashes,
		EndTime:         data.EndTime,
		Version:         data.Version,
		UnknownCommands: data.UnknownCommands,
		Timestamp:       time.Now().Unix(),
	}

	byteData, _ := json.Marshal(redata)
	return byteData, nil
}
