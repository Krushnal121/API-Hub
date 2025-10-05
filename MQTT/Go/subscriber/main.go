package main

import (
    "fmt"
    "time"
    mqtt "github.com/eclipse/paho.mqtt.golang"
)
// Message handler callback
var messageHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
    fmt.Printf("Received: %s from topic: %s\n", msg.Payload(), msg.Topic())
}
func main() {
    opts := mqtt.NewClientOptions()
    opts.AddBroker("tcp://localhost:1883")
    opts.SetClientID("dashboard")
    opts.SetDefaultPublishHandler(messageHandler)
    
    client := mqtt.NewClient(opts)
    if token := client.Connect(); token.Wait() && token.Error() != nil {
        panic(token.Error())
    }
    fmt.Println("Subscriber connected")
    
    // Subscribe to temperature topic with QoS 1
    if token := client.Subscribe("home/livingroom/temperature", 1, nil); token.Wait() && token.Error() != nil {
        panic(token.Error())
    }
    fmt.Println("Subscribed to topic")
    
    // Keep listening for 30 seconds
    time.Sleep(30 * time.Second)
    
    client.Disconnect(250)
}