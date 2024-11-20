// Package main ...
package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/Jennifer-chen-rh/ptp-events-consumer/internal/common"
	"github.com/google/uuid"
	"github.com/redhat-cne/sdk-go/pkg/pubsub"
)

var (
	resourcePrefix     = "/cluster/node/%s%s"
	subs               []pubsub.PubSub
	nodeName           string
	nodeNameFull       string
	namespace          string
	port               string
	clientID           uuid.UUID
	stopHTTPServerChan chan bool
	wg                 sync.WaitGroup
)

func initResources() {
	for _, resource := range common.GetResources() {
		subs = append(subs, pubsub.PubSub{
			ID:       getUUID(fmt.Sprintf(resourcePrefix, nodeNameFull, resource)).String(),
			Resource: fmt.Sprintf(resourcePrefix, nodeNameFull, resource),
		})
	}
	log.Println("initResources ", subs)
}

func main() {
	cancelChan := make(chan os.Signal, 1)
	stopHTTPServerChan = make(chan bool)
	signal.Notify(cancelChan, syscall.SIGTERM, syscall.SIGINT)
	wg = sync.WaitGroup{}
	nodeNameFull = os.Getenv("NODE_NAME")
	namespace = os.Getenv("NAME_SPACE")
	nodeName = strings.Split(nodeNameFull, ".")[0]
	port = os.Getenv("CONSUMER_PORT")
	if namespace == "" {
		log.Printf("please set env variable, export NAME_SPACE=k8s namespace")
		os.Exit(1)
	}
	if port == "" {
		log.Printf("please set env variable, export CONSUMER_PORT=consumer app listening port")
		os.Exit(1)
	}

	publisherServiceName := fmt.Sprintf("http://ptp-event-publisher-service-%s.openshift-ptp.svc.cluster.local:9043", nodeName)

	//publisherServiceNameSub := fmt.Sprintf("http://ptp-event-publisher-service-%s.openshift-ptp.svc.cluster.local:9043", nodeNameFull)
	clientAddress := fmt.Sprintf("0.0.0.0:%s", port)
	clientExternalEndPoint := fmt.Sprintf("http://ptp-event-consumer-service.%s.svc.cluster.local:%s", namespace, port)

	clientID = func(serviceName string) uuid.UUID {
		var namespace = uuid.NameSpaceURL
		var url = []byte(serviceName)
		return uuid.NewMD5(namespace, url)
	}(clientExternalEndPoint)

	log.Println("clientID ", clientID)
	//set up log file daemon stand alone
	//not needed when running from the container
	log.Println("+++++++++ START ", publisherServiceName, " ptp events log +++++++++ ")
	//log.Println("+++++++++ START publisherServiceNameSub ", publisherServiceNameSub, " ptp events log +++++++++ ")
	log.Println("+++++++++ START ", clientAddress, " ptp events log +++++++++ ")
	log.Println("+++++++++ START ", clientExternalEndPoint, " ptp events log +++++++++ ")

	//1. health check on publisher endpoint-check for availability
	healthCheckPublisher(publisherServiceName)

	//2. consumer app - spin web server to receive events
	wg.Add(1)
	go common.StartServer(&wg, clientAddress, stopHTTPServerChan)

	// EVENT subscription and consuming
	initResources()
	// 1.first subscribe to all resources
	if e := common.Subscribe(clientID, subs, nodeName, fmt.Sprintf("%s/subscription", publisherServiceName),
		clientExternalEndPoint); e != nil {
		log.Printf("error processing subscription %s", e)
		stopHTTPServerChan <- true
		os.Exit(1)
	}

	// 2. event will be received at /event as event happens
	// Polling for events
	//2.call get current state once for all three resource
	common.PrintHeader()
	// get current state
	callGetCurrentState(publisherServiceName)
	fmt.Println("\n---------------------------------------------------------------------------------")
	//wg.Add(1)
	//check the pub heath
	go func() {
		defer wg.Done()
		for {
			time.Sleep(300 * time.Second)
			healthCheckPublisher(publisherServiceName)
			callGetCurrentState(publisherServiceName)
		}
	}()

	<-cancelChan
	log.Println("handling exit")
	callGetCurrentState(publisherServiceName)
	if err := common.DeleteSubscription(fmt.Sprintf("%s/subscription", publisherServiceName), clientID); err != nil {
		log.Println(err.Error())
	}
	stopHTTPServerChan <- true
	wg.Wait()
}

func callGetCurrentState(publisherServiceName string) {
	for _, r := range subs {
		if event, data, err := common.GetCurrentState(clientID, publisherServiceName, r.Resource); err == nil {
			log.Println("Succeeded callGetCurrentStat ", event.Time(), event.Type(), data.Values)

		} else {
			log.Printf("Failed callGetCurrentStat %s callGetCurrentState error: %s", r, err.Error())
		}
	}
}

func getUUID(s string) uuid.UUID {
	var namespace = uuid.NameSpaceURL
	var url = []byte(s)
	return uuid.NewMD5(namespace, url)
}

func healthCheckPublisher(publisherServiceName string) {
	resp, err := http.Get(fmt.Sprintf("%s/health", publisherServiceName))
	if err != nil {
		if errors.Is(err, syscall.ECONNREFUSED) {
			log.Printf("port forwarding is not set to access k8s service")
			//help()
			log.Fatalln(err)
		}
		log.Println("healthCheckPublisher failed")
		log.Fatalln(err)

	}
	log.Println("healthCheckPublisher Succeeds ")
	defer resp.Body.Close()
}
