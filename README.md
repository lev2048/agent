![](https://img.shields.io/badge/platform-linux-green.svg) 
![](https://img.shields.io/badge/language-go-blue.svg)
# ServerStatus Agent：   

* 获取服务状态信息的探针 golang版本包
* 只支持linux,参考serverStatus项目python探针实现

## 数据格式
```
type Data struct {
	Name     string `json:"name"`
	Online4  bool   `json:"online4"`
	Online6  bool   `json:"online6"`
	IPStatus bool   `json:"ip_status"`

	CPU       float64 `json:"cpu"`
	Load1     float64 `json:"load_1"`
	Load5     float64 `json:"load_5"`
	Load15    float64 `json:"load_15"`
	Uptime    string  `json:"uptime"`
	MemUsed   int64   `json:"memory_used"`
	MemTotal  int64   `json:"memory_total"`
	SwapUsed  int64   `json:"swap_used"`
	SwapTotal int64   `json:"swap_total"`
	HddUsed   int     `json:"hdd_used"`
	HddTotal  int     `json:"hdd_total"`

	TimeCU     string  `json:"time_CU"`
	TimeCM     string  `json:"time_CM"`
	TimeCT     string  `json:"time_CT"`
	LostRateCU float64 `json:"cu_lostRate"`
	LostRateCM float64 `json:"cm_lostRate"`
	LostRateCT float64 `json:"ct_lostRate"`

	TCPCount     int `json:"tcp_count"`
	UDPCount     int `json:"udp_count"`
	ProcessCount int `json:"process_count"`
	ThreadCount  int `json:"thread_count"`

	NetworkRx  int64 `json:"network_rx"`
	NetworkTx  int64 `json:"network_tx"`
	NetworkIn  int64 `json:"network_in"`
	NetworkOut int64 `json:"network_out"`
}

  {
    "name":"",
    "online4":true,
    "online6":false,
    "ip_status":true,
    "cpu":0.015113350125944613,
    "load_1":0.09000000357627869,
    "load_5":0.1899999976158142,
    "load_15":0.18000000715255737,
    "uptime":"177961.79",
    "memory_used":2243552,
    "memory_total":17511292,
    "swap_used":0,
    "swap_total":5715964,
    "hdd_used":11213,
    "hdd_total":931653,
    "time_CU":"7.485662ms",
    "time_CM":"6.691432ms",
    "time_CT":"6.092826ms",
    "cu_lostRate":0,
    "cm_lostRate":0,
    "ct_lostRate":0,
    "tcp_count":138,
    "udp_count":1,
    "process_count":311,
    "thread_count":715,
    "network_rx":1980,
    "network_tx":2022,
    "network_in":1616900885,
    "network_out":560739931
  }

```

## 参考项目

* ServerStatus：https://github.com/cppla/ServerStatus