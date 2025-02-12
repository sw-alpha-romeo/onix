package core

/*
  Onix Config Manager - Pilot
  Copyright (c) 2018-2021 by www.gatblau.org
  Licensed under the Apache License, Version 2.0 at http://www.apache.org/licenses/LICENSE-2.0
  Contributors to this project, hereby assign copyright in this code to the project,
  to be licensed under the same terms as the rest of the code.
*/
import (
	"fmt"
	ctl "github.com/gatblau/onix/pilotctl/types"
	"gopkg.in/mcuadros/go-syslog.v2"
)

// SyslogCollector syslog log collection service that wraps a syslog server
type SyslogCollector struct {
	server *syslog.Server
	port   string
}

// NewCollector creates an instance of a syslog collection service
func NewCollector(bindIP, port string) (*SyslogCollector, error) {
	defer TRA(CE())
	channel := make(syslog.LogPartsChannel)
	sysServ := syslog.NewServer()
	sysServ.SetHandler(syslog.NewChannelHandler(channel))
	// uses RFC3164 because it is default for rsyslog
	sysServ.SetFormat(syslog.RFC3164)
	err := sysServ.ListenUDP(fmt.Sprintf("%s:%s", bindIP, port))
	if err != nil {
		return nil, err
	}
	go func(channel syslog.LogPartsChannel) {
		for logEntry := range channel {
			info, err := ctl.NewHostInfo()
			if err != nil {
				info = &ctl.HostInfo{}
			}
			event, err := ctl.NewEvent(logEntry, *info)
			if err != nil {
				ErrorLogger.Printf("cannot format syslog entry: %s\n", err)
			}
			err = event.Save(SubmitPath())
			if err != nil {
				ErrorLogger.Printf("cannot save syslog entry to file: %s\n", err)
			}
		}
	}(channel)
	return &SyslogCollector{
		server: sysServ,
		port:   port,
	}, nil
}

// Start the server
func (s *SyslogCollector) Start() error {
	defer TRA(CE())
	InfoLogger.Printf("starting syslog collector on port %s\n", s.port)
	return s.server.Boot()
}

// Wait the server
func (s *SyslogCollector) Wait() {
	defer TRA(CE())
	s.server.Wait()
}

// Stop the server
func (s *SyslogCollector) Stop() error {
	defer TRA(CE())
	return s.server.Kill()
}
