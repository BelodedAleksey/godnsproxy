package main

import (
	"log"
	"net/http"
	"time"

	"github.com/miekg/dns"
)

func startServer(address string) {

	tcpHandler := dns.NewServeMux()
	tcpHandler.HandleFunc(".", HandlerTCP)

	udpHandler := dns.NewServeMux()
	udpHandler.HandleFunc(".", HandlerUDP)

	httpHandler := http.NewServeMux()
	httpHandler.HandleFunc("/", HandlerHTTP)

	tcpServer := &dns.Server{Addr: address + ":53",
		Net:          "tcp",
		Handler:      tcpHandler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	//dns server
	udpServer := &dns.Server{Addr: address + ":53",
		Net:          "udp",
		Handler:      udpHandler,
		UDPSize:      65535,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	//Test Http server
	httpServer := &http.Server{
		Addr:    address + ":80",
		Handler: httpHandler,
	}

	go func() {
		if err := tcpServer.ListenAndServe(); err != nil {
			log.Fatal("TCP-server start failed", err.Error())
		}
	}()

	go func() {
		if err := udpServer.ListenAndServe(); err != nil {
			log.Fatal("UDP-server start failed", err.Error())
		}
	}()

	go func() {
		if err := httpServer.ListenAndServe(); err != nil {
			log.Fatal("HTTP-server start failed", err.Error())
		}
	}()
}
