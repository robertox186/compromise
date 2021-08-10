package mando

import (
	"flag"
	"fmt"

	"github.com/0xPolygon/minimal/command/helper"

	//hbbft

	"encoding/binary"
	"encoding/gob"
	"log"
	"math/rand"
	"time"

	"github.com/anthdm/hbbft"
)

// estructura para mandar un mensaje simple
type Mando struct {
	helper.Meta
}

// GetHelperText returns a simple description of the command
func (p *Mando) GetHelperText() string {
	return "Vamos a ver si en algun momento aparece este mensaje"
}

func (i *Mando) DefineFlags() {
	if i.FlagMap == nil {
		// Flag map not initialized
		i.FlagMap = make(map[string]helper.FlagDescriptor)
	}

	i.FlagMap["mando"] = helper.FlagDescriptor{
		Description: "Este es un mando de prueba",
		Arguments: []string{
			"Prueba_Mando",
		},
		ArgumentsOptional: false,
		FlagOptional:      false,
	}
}

// Help implements the cli.IbftInit interface
func (p *Mando) Help() string {
	p.DefineFlags()

	return helper.GenerateHelp(p.Synopsis(), helper.GenerateUsage(p.GetBaseCommand(), p.FlagMap), p.FlagMap)
}

// Synopsis implements the cli.IbftInit interface
func (p *Mando) Synopsis() string {
	return p.GetHelperText()
}

func (p *Mando) GetBaseCommand() string {
	return "mando"
}

// Run implements the cli.mando interface
func (p *Mando) Run(args []string) int {
	flags := flag.NewFlagSet(p.GetBaseCommand(), flag.ContinueOnError)
	var mando string
	flags.StringVar(&mando, "inicia", "", "")
	var adios string

	flags.StringVar(&adios, "adios", "", "")

	if err := flags.Parse(args); err != nil {
		p.UI.Error(err.Error())
		return 1
	}
	if adios == "" {
		p.UI.Error("Te falto colocar el comando adios(Ej: go run main.go mando mando)")
		return 1
	}
	if mando == "" {
		p.UI.Error("Te falto colocar el comando mando(Ej: go run main.go mando mando)")
		return 1
	}

	output := "\nTodo ha corrido correctamente\n"

	output += helper.FormatKV([]string{
		fmt.Sprintln("me alegra que haya funcionado"),
		fmt.Sprintln("variable mando:" + mando),
	})

	output += "\n"

	p.UI.Output(output)
	benchmark(4, 128, 100)
	benchmark(6, 128, 200)
	benchmark(8, 128, 400)
	benchmark(12, 128, 1000)

	return 0
}

///////////////////////////////////////////////ingreso de hbbft a ver si funciona

type message struct {
	from    uint64
	payload hbbft.MessageTuple
}

func benchmark(n, txsize, batchSize int) {
	log.Printf("Starting benchmark %d nodes %d tx size %d batch size over 5 seconds...", n, txsize, batchSize)
	var (
		nodes    = makeNodes(n, 10000, txsize, batchSize)
		messages = make(chan message, 1024*1024)
	)
	for _, node := range nodes {
		if err := node.Start(); err != nil {
			log.Fatal(err)
		}
		for _, msg := range node.Messages() {
			messages <- message{node.ID, msg}
		}
	}

	timer := time.After(5 * time.Second)
running:
	for {
		select {
		case messag := <-messages:
			node := nodes[messag.payload.To]
			hbmsg := messag.payload.Payload.(hbbft.HBMessage)
			if err := node.HandleMessage(messag.from, hbmsg.Epoch, hbmsg.Payload.(*hbbft.ACSMessage)); err != nil {
				log.Fatal(err)
			}
			for _, msg := range node.Messages() {
				messages <- message{node.ID, msg}
			}
		case <-timer:
			for _, node := range nodes {
				total := 0
				for _, txx := range node.Outputs() {
					total += len(txx)
				}
				log.Printf("node (%d) processed a total of (%d) transactions in 5 seconds [ %d tx/s ]",
					node.ID, total, total/5)
			}
			break running
		default:
		}
	}
}

func makeNodes(n, ntx, txsize, batchSize int) []*hbbft.HoneyBadger {
	nodes := make([]*hbbft.HoneyBadger, n)
	for i := 0; i < n; i++ {
		cfg := hbbft.Config{
			N:         n,
			ID:        uint64(i),
			Nodes:     makeids(n),
			BatchSize: batchSize,
		}
		nodes[i] = hbbft.NewHoneyBadger(cfg)
		for ii := 0; ii < ntx; ii++ {
			nodes[i].AddTransaction(newTx(txsize))
		}
	}
	return nodes
}

func makeids(n int) []uint64 {
	ids := make([]uint64, n)
	for i := 0; i < n; i++ {
		ids[i] = uint64(i)
	}
	return ids
}

type tx struct {
	Nonce uint64
	Data  []byte
}

// Size can be used to simulate large transactions in the network.
func newTx(size int) *tx {
	return &tx{
		Nonce: rand.Uint64(),
		Data:  make([]byte, size),
	}
}

// Hash implements the hbbft.Transaction interface.
func (t *tx) Hash() []byte {
	buf := make([]byte, 8) // sizeOf(uint64) + len(data)
	binary.LittleEndian.PutUint64(buf, t.Nonce)
	return buf
}

func init() {
	rand.Seed(time.Now().UnixNano())
	gob.Register(&tx{})
}
