package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/labstack/gommon/log"
	"github.com/valyala/fasthttp"
)

type ReasonCode int

const (
	reasonInternalError ReasonCode = -1
	reasonOk            ReasonCode = 0
	reasonForceEnabled  ReasonCode = 1
	reasonNotOk         ReasonCode = 2
)

type Response struct {
	*NodeStatus
	ReasonText string
	ReasonCode ReasonCode
}

func checkerHandler(ctx *fasthttp.RequestCtx) {
	response := Response{NodeStatus: status}
	ctx.SetContentType("application/json")

	if config.CheckForceEnabled {
		ctx.SetStatusCode(fasthttp.StatusOK)
		response.ReasonText = "Force enabled"
		response.ReasonCode = reasonForceEnabled
	} else if !status.NodeAvailable {
		ctx.SetStatusCode(fasthttp.StatusServiceUnavailable)
		response.ReasonText = "Node isn't available"
		response.ReasonCode = reasonNotOk
	} else if status.NodeAvailable {
		ctx.SetStatusCode(fasthttp.StatusOK)
		response.ReasonText = "OK"
		response.ReasonCode = reasonOk
	}

	if ctx.IsGet() {
		if respJson, err := json.Marshal(response); err != nil {
			errStr := fmt.Sprintf(`{"ReasonText":"Internal checker error","ReasonCode":%d,"err":"%s"}`, reasonInternalError, err)
			ctx.SetBody([]byte(errStr))
			ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		} else {
			ctx.SetBody(respJson)
		}
	}
	return
}

func checker(status *NodeStatus) {
	dsn := fmt.Sprintf("tcp(%s:%d)/api/aliveness-test/%%2F", config.RabbitMQHost, config.RabbitMQPort)
	log.Printf("Connecting to RabbitMQ with dsn: %s", dsn)

	for {
		time.Sleep(time.Duration(config.CheckInterval) * time.Millisecond)
		curStatus := &NodeStatus{}
		curStatus.Timestamp = time.Now().UnixNano() / (int64(time.Millisecond) / int64(time.Nanosecond))

		req, err := http.NewRequest("GET", fmt.Sprintf("http://%s:%d/api/aliveness-test/%%2F", config.RabbitMQHost, config.RabbitMQPort), nil)

		if err != nil {
			log.Printf("Can't do http.NewRequest", err.Error())
			*status = *curStatus
		} else {

			req.Header.Set("Healthcheck", "rabbitmq-checker")
			req.Header.Set("Authorization", fmt.Sprintf("Basic %s", config.RabbitMQBasicAuth))

			resp, err := netClient.Do(req)

			if err != nil {
				log.Printf("Can't do netClient.Do(): %s", err.Error())
				*status = *curStatus
			} else {

				b, _ := ioutil.ReadAll(resp.Body)
				respText := string(b)

				if config.Debug {
					log.Printf("HTTP code: %s, HTTP response: %s", resp.Status, respText)
				}

				curStatus.HTTPResponseCode = resp.StatusCode

				if resp.StatusCode != http.StatusOK {
					curStatus.HTTPResponseText = respText
					curStatus.NodeAvailable = false
				} else {
					curStatus.HTTPResponseText = respText
					curStatus.NodeAvailable = true
				}
				*status = *curStatus
				defer resp.Body.Close()

			}
		}
	}
}
