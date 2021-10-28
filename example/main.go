package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/lev2048/agent"
)

func main() {
	agent := agent.NewAgent("xxx")
	go agent.Start()
	for i := 0; i < 20; i++ {
		data := agent.GetData()
		json, _ := json.Marshal(data)
		fmt.Println(string(json))
		time.Sleep(time.Duration(1) * time.Second)
	}
	if agent.Stop() {
		fmt.Println("stop success")
	}
	fmt.Println("exit...")
}
