package main

import (
    "fmt"
    "time"
    mqtt "github.com/eclipse/paho.mqtt.golang"
)
func main() {
    // Configure connection options
    opts := mqtt.NewClientOptions()
    opts.AddBroker("tcp://localhost:1883")
    opts.SetClientID("temperature_sensor")
    
    // Create and connect client
    client := mqtt.NewClient(opts)
    if token := client.Connect(); token.Wait() && token.Error() != nil {
        panic(token.Error())
    }
    fmt.Println("Connected to broker")
    
    // Publish temperature readings every 2 seconds
    for i := 0; i < 10; i++ {
        temp := 20.0 + float64(i)*0.5
        payload := fmt.Sprintf("%.1f", temp)
        
        // Publish with QoS 1 (at least once delivery)
        token := client.Publish("home/livingroom/temperature", 1, false, payload)
        token.Wait()
        fmt.Printf("Published: %s°C\n", payload)
        time.Sleep(2 * time.Second)
    }
    
    // Always disconnect gracefully to send pending messages
    client.Disconnect(250)
    fmt.Println("Publisher disconnected")
}