package main

import (
	"fmt"
	"math"
	"time"

	"agent"
)

func main() {
	ag := agent.NewAgent("agent")
	go ag.Start(false)
	for i := 0; i < 200; i++ {
		data := ag.GetData()
		if data.CPU != 0 {
			fmt.Println(
				fmt.Sprintf("%d%%", int(math.Ceil(data.CPU*100))),
				fmt.Sprintf("%d%%", int(math.Ceil((float64(data.MemUsed)/float64(data.MemTotal))*100))),
				"up "+agent.UnitConver(float64(data.NetworkTx))+"/s",
				"down "+agent.UnitConver(float64(data.NetworkRx))+"/s",
				agent.UnitConver(float64(data.NetworkIn)),
				agent.UnitConver(float64(data.NetworkOut)),
			)
		}
		//json, _ := json.Marshal(data)
		//fmt.Println(string(json))
		time.Sleep(time.Duration(1) * time.Second)
	}
	if ag.Stop() {
		fmt.Println("stop success")
	}
	fmt.Println("exit...")
}
