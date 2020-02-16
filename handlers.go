package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/miekg/dns"
)

func HandlerTCP(w dns.ResponseWriter, req *dns.Msg) {
	Handler(w, req)
}

func HandlerUDP(w dns.ResponseWriter, req *dns.Msg) {
	Handler(w, req)
}

func HandlerHTTP(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "Redirected from %s\n", req.Host)
}

func Handler(w dns.ResponseWriter, req *dns.Msg) {
	defer w.Close()

	question := req.Question[0]

	cachedReq := cache.Get(question.Qtype, question.Name)
	if cachedReq != nil {
		response := &dns.Msg{}
		response.SetReply(req)
		response.Answer = append(response.Answer, cachedReq)

		w.WriteMsg(response)
		return
	}

	if question.Qtype == dns.TypeA || question.Qtype == dns.TypeAAAA {

		response := &dns.Msg{}
		response.SetReply(req)

		//Block
		if blackList.Contains(question.Name) {
			head := dns.RR_Header{
				Name:   question.Name,
				Rrtype: question.Qtype,
				Class:  dns.ClassINET,
				Ttl:    uint32(config.UpdateInterval.Seconds()),
			}

			var line dns.RR
			if question.Qtype == dns.TypeA {
				line = &dns.A{
					Hdr: head,
					A:   net.ParseIP(config.BlockAddress4),
				}
			} else {
				line = &dns.AAAA{
					Hdr:  head,
					AAAA: net.ParseIP(config.BlockAddress6),
				}
			}
			response.Answer = append(response.Answer, line)
			log.Println("blocked", question.Name)
			w.WriteMsg(response)
			return
		}

		//Redirect to localhost
		if localAliasesList.Contains(question.Name) {
			response.Authoritative = true

			response.Answer = append(response.Answer, &dns.A{
				Hdr: dns.RR_Header{Name: question.Name, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 60},
				A:   net.ParseIP(config.LocalDNSProxy),
			})
			log.Printf("%s redirected to %s\n", question.Name, config.LocalDNSProxy)
			w.WriteMsg(response)
			return
		}
	}

	resp, err := Lookup(req)
	if err != nil {
		resp = &dns.Msg{}
		resp.SetRcode(req, dns.RcodeServerFailure)
		log.Println("fail", question.Name)
	} else {
		if len(resp.Answer) > 0 {
			cache.Set(question.Qtype, question.Name, resp.Answer[0])
		}
	}

	w.WriteMsg(resp)
}

func Lookup(req *dns.Msg) (*dns.Msg, error) {
	c := &dns.Client{
		Net:          "tcp",
		ReadTimeout:  time.Second * 5,
		WriteTimeout: time.Second * 5,
	}

	qName := req.Question[0].Name

	res := make(chan *dns.Msg, 1)
	var wg sync.WaitGroup
	L := func(nameserver string) {
		defer wg.Done()
		r, _, err := c.Exchange(req, nameserver)
		if err != nil {
			log.Printf("%s socket error on %s", qName, nameserver)
			log.Printf("error:%s", err.Error())
			return
		}
		if r != nil && r.Rcode != dns.RcodeSuccess {
			if r.Rcode == dns.RcodeServerFailure {
				return
			}
		}
		select {
		case res <- r:
		default:
		}
	}

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	// Start lookup on each nameserver top-down, in every second
	for _, nameserver := range config.Nameservers {
		wg.Add(1)
		go L(nameserver)
		// but exit early, if we have an answer
		select {
		case r := <-res:
			return r, nil
		case <-ticker.C:
			continue
		}
	}

	// wait for all the namservers to finish
	wg.Wait()
	select {
	case r := <-res:
		return r, nil
	default:
		return nil, errors.New("can't resolve ip for" + qName)
	}
}
