package main

import (
	"fmt"

	"github.com/korkmazkadir/bitcoin-network-analyzer/peer"
)

var seeds = []string{
	"seed.bitcoin.sipa.be.",
	"dnsseed.emzy.de.",
	"151.20.141.9",
}

func main() {
	p := peer.NewPeer(seeds[2], 738728)
	err := p.Connect()
	if err != nil {
		panic(err)
	}

	fmt.Println("Running main loop of the peer...")
	p.MainLoop()
}
