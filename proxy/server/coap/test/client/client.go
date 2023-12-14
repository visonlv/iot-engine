package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/plgd-dev/go-coap/v3/message"
	"github.com/plgd-dev/go-coap/v3/message/pool"
	"github.com/plgd-dev/go-coap/v3/udp"
	"github.com/visonlv/go-vkit/logger"
)

func ob() {
	sync := make(chan bool)
	co, err := udp.Dial("localhost:8082")
	if err != nil {
		log.Fatalf("Error dialing: %v", err)
	}
	num := 0
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	obs, err := co.Observe(ctx, "/productsn/devicesn/up/hhh", func(req *pool.Message) {
		b, _ := io.ReadAll(req.Body())
		log.Printf("Got %+v   %s \n", req, string(b))
		num++
		if num >= 10 {
			sync <- true
		}
	})
	if err != nil {
		log.Fatalf("Unexpected error '%v'", err)
	}
	<-sync
	ctx, cancel = context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	obs.Cancel(ctx)
}

func post() {
	logger.Infof("start connect")
	co, err := udp.Dial("10.55.132.187:5683")
	if err != nil {
		log.Fatalf("Error dialing: %v", err)
		return
	}

	logger.Infof("start success")

	logger.Infof("start send")

	// 连接
	path := "mqtt/connection"
	options1 := message.Option{ID: message.URIQuery, Value: []byte("clientid=123")}
	options2 := message.Option{ID: message.URIQuery, Value: []byte("username=admin")}
	options3 := message.Option{ID: message.URIQuery, Value: []byte("password=public1")}
	resp, err := co.Post(context.Background(), path, message.AppJSON, nil, options1, options2, options3)
	if err != nil {
		log.Fatalf("Error sending request: %v", err)
		return
	}

	logger.Infof("%v", resp)

	b, _ := io.ReadAll(resp.Body())
	log.Printf("Response payload: %v body:%s", resp.String(), string(b))
	token := string(b)

	// 发布消息
	go func() {
		for {
			time.Sleep(time.Second * 5)
			co, err := udp.Dial("10.55.132.187:5683")
			if err != nil {
				log.Fatalf("Error dialing: %v", err)
				return
			}

			path := "mqtt/connection"
			options1 := message.Option{ID: message.URIQuery, Value: []byte("clientid=123")}
			options2 := message.Option{ID: message.URIQuery, Value: []byte(fmt.Sprintf("token=%s", token))}

			logger.Infof("start send %s", token)

			resp2, err := co.Put(context.Background(), path, message.AppJSON, nil, options1, options2)
			if err != nil {
				log.Fatalf("Error sending request: %v", err)
				return
			}

			logger.Infof("%v", resp2)
		}
		// b, _ := io.ReadAll(resp2.Body())
		// log.Printf("Response payload: %v body:%s", resp2.String(), string(b))
	}()

	go func() {
		for {
			time.Sleep(time.Second * 2)
			co, err := udp.Dial("10.55.132.187:5683")
			if err != nil {
				log.Fatalf("Error dialing: %v", err)
				return
			}

			path := "ps/coap/test"
			options1 := message.Option{ID: message.URIQuery, Value: []byte("clientid=123")}
			options2 := message.Option{ID: message.URIQuery, Value: []byte(fmt.Sprintf("token=%s", token))}

			logger.Infof("start send %s", token)

			resp2, err := co.Post(context.Background(), path, message.AppJSON, bytes.NewReader([]byte("{\"msg\":\"哈哈哈是佛挡杀佛\"}")), options1, options2)
			if err != nil {
				log.Fatalf("Error sending request: %v", err)
				return
			}

			logger.Infof("%v", resp2)
		}
	}()

	time.Sleep(time.Hour)
}

func main() {
	post()
}
