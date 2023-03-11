package main

import (
    "fmt"
    "time"

    mqtt "github.com/eclipse/paho.mqtt.golang"
)

var gatewayId = "station1" 
var gatewayPassword = "123456?aD"
var broker = "localhost"
var port = 1883

var joinAcceptTopic = "frames/joinaccept"
var configsTopic = fmt.Sprintf("gateways/%s", gatewayId)

var publishTopic = "frames/joinrequest"

var receivedMsgChannel = make(chan string)

var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
    receivedMsgChannel <- fmt.Sprintf("Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
    fmt.Println("Connected")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
    fmt.Printf("Connect lost: %v", err)
}

func main() {
    opts := mqtt.NewClientOptions()
    opts.AddBroker(fmt.Sprintf("tcp://%s:%d", broker, port))
    opts.SetClientID(gatewayId)
    opts.SetUsername(gatewayId)
    opts.SetPassword(gatewayPassword)
    opts.SetDefaultPublishHandler(messagePubHandler)
    opts.OnConnect = connectHandler
    opts.OnConnectionLost = connectLostHandler
    client := mqtt.NewClient(opts)
    if token := client.Connect(); token.Wait() && token.Error() != nil {
        panic(token.Error())
    }

    topics := map[string]byte {
        joinAcceptTopic: 1, 
        configsTopic: 1,
    }
    token := client.SubscribeMultiple(topics, nil)
    token.Wait()
    
    go publish(client)
    
    for {
        msg := <- receivedMsgChannel
        fmt.Println(msg)
    }

    client.Disconnect(250)
    fmt.Println("Client disconnected")
}

func publish(client mqtt.Client) {
    for i := 0; true; i++ {
        text := fmt.Sprintf("AAEBAQEBAQEBAgICAgICAgIDAwm5ezI=")
        token := client.Publish(publishTopic, 0, false, text)
        token.Wait()
        time.Sleep(time.Second)
    }
}