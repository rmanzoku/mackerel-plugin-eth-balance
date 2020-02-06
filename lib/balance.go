package mpethbalance

import (
	"context"
	"errors"
	"flag"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	mp "github.com/mackerelio/go-mackerel-plugin"
)

// EthBalancePlugin mackerel plugin for Ethereum balance
type EthBalancePlugin struct {
	Prefix    string
	RPC       string
	Addresses []LabeledAddress
	EthClient *ethclient.Client
}

type LabeledAddress struct {
	Name    string
	Address common.Address
	Label   string
}

// FetchMetrics interface for mackerelplugin
func (p EthBalancePlugin) FetchMetrics() (map[string]float64, error) {
	ctx := context.TODO()
	ret := make(map[string]float64)
	for _, a := range p.Addresses {
		key := a.Name
		balance, err := p.EthClient.BalanceAt(ctx, a.Address, nil)
		if err != nil {
			return nil, err
		}
		ret[key] = weiToEther(balance)
	}
	return ret, nil
}

// GraphDefinition interface for mackerelplugin
func (p EthBalancePlugin) GraphDefinition() map[string]mp.Graphs {
	metrics := make([]mp.Metrics, len(p.Addresses))
	for i, a := range p.Addresses {
		metrics[i] = mp.Metrics{
			Name:  a.Name,
			Label: a.Label,
		}
	}
	return map[string]mp.Graphs{
		"balance": {
			Label:   "Ether",
			Unit:    "float",
			Metrics: metrics,
		},
	}
}

// MetricKeyPrefix interface for PluginWithPrefix
func (p EthBalancePlugin) MetricKeyPrefix() string {
	return p.Prefix
}

// Do the plugin
func Do() {

	var err error
	var eb EthBalancePlugin
	var rpc string
	var addresses string
	var tempfile string

	flag.StringVar(&eb.Prefix, "metric-key-prefix", "ethereum", "Metric key prefix")
	flag.StringVar(&rpc, "rpc", "", "Ethereum rpc")
	flag.StringVar(&addresses, "addresses", "0x0", "Metric key prefix")
	flag.StringVar(&tempfile, "tempfile", "", "Temp file name")
	flag.Parse()

	eb.Addresses = parseAddresses(addresses)
	eb.EthClient, err = ethclient.Dial(rpc)
	if err != nil {
		panic(err)
	}

	helper := mp.NewMackerelPlugin(eb)
	helper.Tempfile = tempfile
	helper.Run()
}

func parseAddresses(addresses string) []LabeledAddress {
	s := strings.Split(addresses, ",")
	ret := make([]LabeledAddress, len(s))
	for i, la := range strings.Split(addresses, ",") {
		a := strings.Split(la, ":")
		addr := strings.ToLower(a[0])
		if !common.IsHexAddress(a[0]) {
			panic(errors.New("Invalid address: " + a[0]))
		}
		var l string
		if len(a) == 2 {
			l = a[1]
		} else {
			l = addr
		}
		ret[i] = LabeledAddress{
			Name:    addr,
			Address: common.HexToAddress(addr),
			Label:   l,
		}
	}
	return ret
}

func weiToEther(wei *big.Int) float64 {
	ether, _ := new(big.Float).SetString("1000000000000000000")
	w := new(big.Float).SetInt(wei)
	ret, _ := new(big.Float).Quo(w, ether).Float64()
	return ret
}
