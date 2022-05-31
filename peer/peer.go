package peer

import (
	"fmt"
	"net"

	"github.com/btcsuite/btcd/wire"
)

const (
	defaultPort = 8333
	localIP     = "147.210.129.35"
)

type Peer struct {
	address         string
	lastBlockHeight int32
	conn            net.Conn

	pver   uint32
	btcnet wire.BitcoinNet
}

func NewPeer(address string, startHeight int32) *Peer {
	p := &Peer{address: address, lastBlockHeight: startHeight}
	p.pver = wire.ProtocolVersion
	p.btcnet = wire.MainNet
	return p
}

func (p *Peer) Connect() error {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", p.address, defaultPort))
	if err != nil {
		return err
	}
	p.conn = conn
	return nil
}

func (p *Peer) MainLoop() {

	err := p.sendVersion()
	if err != nil {
		panic(err)
	}

	for {
		msg, rawPayload, err := wire.ReadMessage(p.conn, p.pver, p.btcnet)
		if err != nil {
			panic(err)
		}

		//fmt.Printf("Command of read message: %s\n", msg.Command())
		fmt.Printf("Raw payload length: %d\n", len(rawPayload))

		p.handleMessage(msg)

	}
}

func (p *Peer) handleMessage(msg wire.Message) error {
	switch msg := msg.(type) {
	case *wire.MsgVersion:
		fmt.Printf("===> Peer Protocol version: %v\n", msg.ProtocolVersion)
		fmt.Printf("===> Peer Services: %s\n", msg.Services.String())
	case *wire.MsgVerAck:
		fmt.Printf("===> VerAck message received...\n")
		err := wire.WriteMessage(p.conn, wire.NewMsgVerAck(), p.pver, p.btcnet)
		if err != nil {
			return err
		}
		return p.requestPeerList()

	case *wire.MsgInv:
		fmt.Println("===> INV message received:")
		return p.requestFirstMessage(msg)

	case *wire.MsgTx:
		fmt.Println("TX received:")
		fmt.Printf("---> TX Hash: %x\n", msg.TxHash())
		fmt.Printf("---> Witness Hash: %x\n", msg.WitnessHash())
	case *wire.MsgBlock:
		fmt.Println("Block received:")
		fmt.Printf("---> Block Hash: %x\n", msg.BlockHash())
		fmt.Printf("---> Merkle Root: %x\n", msg.Header.MerkleRoot)

	case *wire.MsgAddr:
		fmt.Println("Address message received:")

		for i, addr := range msg.AddrList {
			fmt.Printf("[%d] %s:%d\n", i, addr.IP, addr.Port)
		}

	default:
		fmt.Errorf("###### Undefined command. Message command: %s ######\n", msg.Command())
	}

	return nil
}

func (p *Peer) sendVersion() error {
	me := &wire.NetAddress{IP: net.IP(localIP), Port: defaultPort}
	you := &wire.NetAddress{IP: net.IP(p.address), Port: defaultPort}

	msgVersion := wire.NewMsgVersion(me, you, 546118, p.lastBlockHeight)
	msgVersion.AddService(wire.SFNodeWitness)
	msgVersion.AddService(wire.SFNodeNetwork)

	return wire.WriteMessage(p.conn, msgVersion, p.pver, p.btcnet)
}

func (p *Peer) requestFirstMessage(inv *wire.MsgInv) error {

	msgGetData := wire.NewMsgGetData()
	msgGetData.AddInvVect(inv.InvList[0])

	return wire.WriteMessage(p.conn, msgGetData, p.pver, p.btcnet)
}

func (p *Peer) requestPeerList() error {

	msgGetAddress := wire.NewMsgGetAddr()
	return wire.WriteMessage(p.conn, msgGetAddress, p.pver, p.btcnet)
}
