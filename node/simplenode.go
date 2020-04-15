package node

import (
	"encoding/json"
	"log"
)

const (
	Proposal  = 1
	PreVote   = 2
	PreCommit = 3
	Committed = 4
)

type Block struct {
	Blockid string
	Height  int
}

type Message struct {
	MsgType int
	Data    *Block
	Signer  string
}

type Node struct {
	Id               string
	Mailbox          chan *Message
	Blocks           []*Block
	neighbours       []*Node
	prevoteCounter   map[string]int
	precommitCounter map[string]int
	precommitLocks   map[int]*Block
	blockStatus      map[string]int
}

func (msg *Message) String() string {
	m, _ := json.Marshal(msg)
	return string(m)
}

func (msg *Block) String() string {
	m, _ := json.Marshal(msg)
	return string(m)
}

func NewNode(id string) *Node {
	me := &Node{
		Id:               id,
		Mailbox:          make(chan *Message),
		Blocks:           []*Block{},
		prevoteCounter:   map[string]int{},
		precommitCounter: map[string]int{},
		precommitLocks:   map[int]*Block{},
		blockStatus:      map[string]int{},
	}
	go me.start()
	return me
}

func (nd *Node) SetNeighbours(neighbours []*Node) {
	nd.neighbours = neighbours
}

func (nd *Node) Broadcast(msg *Message) {
	for _, n := range nd.neighbours {
		n.Mailbox <- msg
	}
}

func (nd *Node) start() {
	for msg := range nd.Mailbox {
		switch msg.MsgType {
		case Proposal:
			log.Println(nd.Id, "received proposal", msg)
			if nd.blockStatus[msg.Data.Blockid] != 0 {
				log.Println(nd.Id, "ignore proposal", msg)
				continue
			}
			if msg.Data.Height != len(nd.Blocks) {
				log.Println(nd.Id, "invalid block height", msg)
				continue
			}
			if _, exist := nd.precommitLocks[msg.Data.Height]; exist {
				msg.Data = nd.precommitLocks[msg.Data.Height]
			}
			nd.blockStatus[msg.Data.Blockid] = PreVote
			go nd.Broadcast(&Message{MsgType: PreVote, Data: msg.Data, Signer: nd.Id})
		case PreVote:
			log.Println(nd.Id, "received prevote", msg)
			//if nd.Id == "C" {
			//		continue //bad node
			//}
			//if nd.Id == "D" {
			//		continue //if we add more bad node, the block will not be confirmed.
			//}
			if nd.blockStatus[msg.Data.Blockid] > PreVote {
				log.Println(nd.Id, "ignore prevote", msg)
				continue
			}
			nd.prevoteCounter[msg.Data.Blockid]++
			collected := nd.prevoteCounter[msg.Data.Blockid]
			if collected > len(nd.neighbours)*2/3 {
				nd.precommitLocks[msg.Data.Height] = msg.Data
				nd.blockStatus[msg.Data.Blockid] = PreCommit
				go nd.Broadcast(&Message{MsgType: PreCommit, Data: msg.Data, Signer: nd.Id})
			}
		case PreCommit:
			log.Println(nd.Id, "received precommit", msg)
			if nd.blockStatus[msg.Data.Blockid] > PreCommit {
				log.Println(nd.Id, "ignore precommit", msg)
				continue
			}
			nd.precommitCounter[msg.Data.Blockid]++
			collected := nd.precommitCounter[msg.Data.Blockid]
			if collected > len(nd.neighbours)*2/3 {
				delete(nd.precommitLocks, msg.Data.Height)
				nd.Blocks = append(nd.Blocks, msg.Data)
				nd.blockStatus[msg.Data.Blockid] = Committed
			}
		}
	}
}

func (nd *Node) ProposeBlock(block *Block) {
	msg := &Message{
		MsgType: Proposal,
		Data:    block,
	}
	nd.Broadcast(msg)
}
