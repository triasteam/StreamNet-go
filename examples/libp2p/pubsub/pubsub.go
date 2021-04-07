/*
 *
 * The MIT License (MIT)
 *
 * Copyright (c) 2014 Juan Batiz-Benet
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in
 * all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
 * THE SOFTWARE.
 *
 * This program demonstrate a gossip application using p2p pubsub protocol
 *
 * this file describes how pubsub works
 */
package main

import (
	"bytes"
	"context"
	"fmt"
	"time"

	libp2p "github.com/libp2p/go-libp2p"
	host "github.com/libp2p/go-libp2p-host"
	pstore "github.com/libp2p/go-libp2p-peerstore"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
)

const gossipSubID = "/meshsub/1.0.0"

func main() {

	//golog.SetAllLoggers(gologging.DEBUG) // Change to DEBUG for extra info
	h1 := newHost(2001)
	h2 := newHost(2002)
	h3 := newHost(2003)
	fmt.Printf("host 1: \n\t-Addr:%s\n\t-ID: %s\n", h1.Addrs()[0], h1.ID().Pretty())
	fmt.Printf("host 2: \n\t-Addr:%s\n\t-ID: %s\n", h2.Addrs()[0], h2.ID().Pretty())
	fmt.Printf("host 3: \n\t-Addr:%s\n\t-ID: %s\n", h3.Addrs()[0], h3.ID().Pretty())

	time.Sleep(100 * time.Millisecond)

	// add h1 to h2's store
	h2.Peerstore().AddAddr(h1.ID(), h1.Addrs()[0], pstore.PermanentAddrTTL)
	// add h2 to h1's store
	h1.Peerstore().AddAddr(h2.ID(), h2.Addrs()[0], pstore.PermanentAddrTTL)
	// add h3 to h2's store
	h2.Peerstore().AddAddr(h3.ID(), h3.Addrs()[0], pstore.PermanentAddrTTL)
	// add h2 to h3's store
	h3.Peerstore().AddAddr(h3.ID(), h3.Addrs()[0], pstore.PermanentAddrTTL)

	// ---- gossip sub part
	topic := "random"
	opts := pubsub.WithMessageSigning(false)
	g1, err := pubsub.NewGossipSub(context.Background(), h1, opts)
	requireNil(err)
	g2, err := pubsub.NewGossipSub(context.Background(), h2, opts)
	requireNil(err)
	g3, err := pubsub.NewGossipSub(context.Background(), h3, opts)
	requireNil(err)

	s1, err := g1.Subscribe(topic)
	requireNil(err)
	s2, err := g2.Subscribe(topic)
	requireNil(err)
	s3, err := g3.Subscribe(topic)
	requireNil(err)

	// 1 connect to 2 and 2 connect to 3
	err = h1.Connect(context.Background(), h2.Peerstore().PeerInfo(h2.ID()))
	requireNil(err)
	err = h2.Connect(context.Background(), h3.Peerstore().PeerInfo(h3.ID()))
	requireNil(err)
	time.Sleep(3 * time.Second)

	// publish and read
	{
		msg := []byte("Hello Word")
		requireNil(g1.Publish(topic, msg))

		pbMsg, err := s1.Next(context.Background())
		requireNil(err)
		checkEqual(msg, pbMsg.Data)
		fmt.Println(" GOSSIPING WORKS #1")

		pbMsg, err = s2.Next(context.Background())
		requireNil(err)
		checkEqual(msg, pbMsg.Data)
		fmt.Println(" GOSSIPING WORKS #2")

		pbMsg, err = s3.Next(context.Background())
		requireNil(err)
		checkEqual(msg, pbMsg.Data)
		fmt.Println(" GOSSIPING WORKS #3")
	}

	{
		msg := []byte("Hello Word 3")
		requireNil(g3.Publish(topic, msg))

		pbMsg, err := s1.Next(context.Background())
		requireNil(err)
		checkEqual(msg, pbMsg.Data)
		fmt.Println(" GOSSIPING WORKS #1")

		pbMsg, err = s2.Next(context.Background())
		requireNil(err)
		checkEqual(msg, pbMsg.Data)
		fmt.Println(" GOSSIPING WORKS #2")

		pbMsg, err = s3.Next(context.Background())
		requireNil(err)
		checkEqual(msg, pbMsg.Data)
		fmt.Println(" GOSSIPING WORKS #3")
	}
}

func checkEqual(exp, rcvd []byte) {
	if !bytes.Equal(exp, rcvd) {
		panic("not equal")
	}
}

func requireNil(err error) {
	if err != nil {
		panic(err)
	}
}

func newHost(port int) host.Host {
	opts := []libp2p.Option{
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", port)),
		libp2p.DisableRelay(),
	}
	basicHost, err := libp2p.New(context.Background(), opts...)
	if err != nil {
		panic(err)
	}
	return basicHost
}
