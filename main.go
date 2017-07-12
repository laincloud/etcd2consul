package main

import (
	"context"
	"flag"
	"github.com/coreos/etcd/client"
	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/go-cleanhttp"
	"log"
	"time"
)

var (
	etcdAddr, consulAddr string
)

func init() {
	flag.StringVar(&etcdAddr, "etcd", "", "etcd addr")
	flag.StringVar(&consulAddr, "consul", "", "consul addr")
	flag.Parse()
}

func main() {
	cfg := client.Config{
		Endpoints:               []string{"http://" + etcdAddr},
		Transport:               client.DefaultTransport,
		HeaderTimeoutPerRequest: time.Second,
	}
	c, err := client.New(cfg)
	if err != nil {
		log.Fatal(err)
	}
	kapi := client.NewKeysAPI(c)
	keys, values := getKVByKey(kapi, "/docker")

	config := &api.Config{
		Address:   consulAddr,
		Scheme:    "http",
		Transport: cleanhttp.DefaultPooledTransport(),
	}

	consulClient, err := api.NewClient(config)
	if err != nil {
		log.Fatal(err)
	}

	kv := consulClient.KV()

	for i := 0; i < len(keys); i++ {
		p := &api.KVPair{Key: keys[i][1:] + "/", Value: []byte(values[i])}
		_, err = kv.Put(p, nil)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func getKVByKey(kapi client.KeysAPI, key string) (keys, values []string) {
	resp, err := kapi.Get(context.Background(), key, nil)
	if err != nil {
		log.Fatal(err)
	} else {
		if resp.Node.Dir {
			if len(resp.Node.Nodes) > 0 {
				for i := 0; i < len(resp.Node.Nodes); i++ {
					k, v := getKVByKey(kapi, resp.Node.Nodes[i].Key)
					keys = append(keys, k...)
					values = append(values, v...)
				}
			}
		} else {
			keys = append(keys, resp.Node.Key)
			values = append(values, resp.Node.Value)
			log.Printf("%q key has %q value\n", resp.Node.Key, resp.Node.Value)
		}
	}
	return keys, values
}
