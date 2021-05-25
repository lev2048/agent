package agent

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-ping/ping"
)

type Agent struct {
	data Data

	cu    string
	ct    string
	cm    string
	run   chan bool
	exit  chan int
	token string
}

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

func NewAgent(token string) *Agent {
	return &Agent{
		data:  Data{},
		cu:    "cu.tz.shield.asia",
		ct:    "ct.tz.shield.asia",
		cm:    "cm.tz.shield.asia",
		token: token,
	}
}

func (at *Agent) getUptime() {
	//第一列 运行时间 第二列 空闲时间
	file, err := os.OpenFile("/proc/uptime", os.O_RDONLY, os.ModeSymlink)
	if err != nil {
		fmt.Println("open uptime file err", err)
		return
	}
	defer file.Close()
	buf := bufio.NewReader(file)
	line, _ := buf.ReadString('\n')
	upTimeInfo := strings.Split(line, " ")
	at.data.Uptime = upTimeInfo[0]
}

func (at *Agent) getMemory() {
	re := regexp.MustCompile(`^(?P<key>\S*):\s*(?P<value>\d*)\s*kB`)
	file, _ := os.OpenFile("/proc/meminfo", os.O_RDONLY, os.ModeSymlink)
	buf := bufio.NewReader(file)
	data := make(map[string]int64)
	for {
		line, err := buf.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				//File read ok!
				break
			} else {
				fmt.Println("Read file error!", err)
				return
			}
		}
		k, v := at.getRxParams(re, line)
		if k == "" {
			continue
		}
		data[k] = v
	}
	at.data.MemTotal = data["MemTotal"]
	at.data.MemUsed = data["MemTotal"] - data["MemFree"] - data["Buffers"] - data["Cached"] - data["SReclaimable"]
	at.data.SwapTotal = data["SwapTotal"]
	at.data.SwapUsed = data["SwapTotal"] - data["SwapFree"]
}

func (at *Agent) getRxParams(rx *regexp.Regexp, str string) (key string, value int64) {
	if !rx.MatchString(str) {
		return
	}
	p := rx.FindStringSubmatch(str)
	n := rx.SubexpNames()
	for i := range n {
		if i == 0 {
			continue
		}
		if n[i] != "" && p[i] != "" {
			if n[i] == "key" {
				key = p[i]
			} else {
				v, err := strconv.ParseInt(p[i], 10, 64)
				if err != nil {
					value = 0
				} else {
					value = v
				}

			}
		}
	}
	return
}

func (at *Agent) getDisk() {
	cmd := exec.Command("df", "-Tlm", "--total", "-t", "ext4", "-t", "ext3", "-t", "ext2", "-t", "reiserfs", "-t", "jfs", "-t", "ntfs", "-t", "fat32", "-t", "btrfs", "-t", "fuseblk", "-t", "zfs", "-t", "simfs", "-t", "xfs")
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(err)
	}
	diskInfo := strings.Split(string(out), "\n")
	totalDiskInfo := strings.Fields(diskInfo[len(diskInfo)-2])
	at.data.HddTotal, _ = strconv.Atoi(totalDiskInfo[2])
	at.data.HddUsed, _ = strconv.Atoi(totalDiskInfo[3])
}

func (at *Agent) getCpuUseInfo() {
	strToInt64 := func(data string) int64 {
		v, _ := strconv.ParseInt(data, 10, 0)
		return v
	}
	getData := func() (data []int64) {
		file, _ := os.OpenFile("/proc/stat", os.O_RDONLY, os.ModeSymlink)
		buf := bufio.NewReader(file)
		line, _ := buf.ReadString('\n')
		for _, v := range strings.Fields(line)[1:] {
			data = append(data, strToInt64(v))
		}
		return data
	}
	sumData := func(data []int64) (result int64) {
		for _, v := range data {
			result += v
		}
		return
	}
	d1 := getData()
	time.Sleep(time.Duration(1) * time.Second)
	d2 := getData()
	totalD1 := sumData(d1)
	totalD2 := sumData(d2)
	at.data.CPU = 1 - float64(d2[3]+d2[4]-d1[3]-d1[4])/float64(totalD2-totalD1)
}

func (at *Agent) getTrafficInfo() (int64, int64) {
	isDevSkip := func(name, b1, b2 string) bool {
		devs := []string{"lo", "tun", "docker", "veth", "br-", "vmbr", "vnet", "kube"}
		if b1 == "0" || b2 == "0" {
			return false
		}
		for _, v := range devs {
			if v == name[:len(name)-1] || v == name[:2] {
				return true
			} else {
				return false
			}
		}
		return false
	}
	re := regexp.MustCompile(`([^\s]+):[\s]{0,}(\d+)\s+(\d+)\s+(\d+)\s+(\d+)\s+(\d+)\s+(\d+)\s+(\d+)\s+(\d+)\s+(\d+)\s+(\d+)\s+(\d+)`)
	file, _ := os.OpenFile("/proc/net/dev", os.O_RDONLY, os.ModeSymlink)
	buf := bufio.NewReader(file)
	var NET_IN, NET_OUT int
	for {
		line, err := buf.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				//File read ok!
				break
			} else {
				fmt.Println("Read file error!", err)
				return 0, 0
			}
		}
		res := re.FindAllStringSubmatch(line, -1)
		if len(res) > 0 {
			devInfo := strings.Fields(res[0][0])
			if !isDevSkip(devInfo[0], devInfo[1], devInfo[9]) {
				NET_IN += at.numberCover(devInfo[1], "int").(int)
				NET_OUT += at.numberCover(devInfo[9], "int").(int)
			} else {
				continue
			}
		}

	}
	return int64(NET_IN), int64(NET_OUT)
}

//获取流量统计信息
func (at *Agent) getTrafficStats() {
	at.data.NetworkIn, at.data.NetworkOut = at.getTrafficInfo()
}

//获取tcp/udp/process/thread 计数
func (at *Agent) getTUPDCount() {
	getOutput := func(command string) string {
		cmd := exec.Command("bash", "-c", command)
		out, _ := cmd.CombinedOutput()
		return strings.Replace(string(out), "\n", "", -1)
	}
	var tupd []int
	for _, command := range []string{"ss -t|wc -l", "ss -u|wc -l", "ps -ef|wc -l", "ps -eLf|wc -l"} {
		tupd = append(tupd, at.numberCover(getOutput(command), "int").(int))
	}
	at.data.TCPCount = tupd[0]
	at.data.UDPCount = tupd[1]
	at.data.ProcessCount = tupd[2]
	at.data.ThreadCount = tupd[3]
}

//通过cu/ct/cm 判断国内网络
func (at *Agent) getChinaNetStatus() {
	netErrCount := 0
	for _, dist := range []string{at.cu, at.ct, at.cm} {
		conn, err := net.DialTimeout("tcp", dist+":80", time.Duration(1)*time.Second)
		if err != nil {
			netErrCount++
		} else {
			conn.Close()
		}
	}
	if netErrCount > 2 {
		at.data.IPStatus = false
	} else {
		at.data.IPStatus = true
	}
}

//通过谷歌ip /v4/v6 判断网络状态
func (at *Agent) getWordNetStatus() {
	getNetWork := func(domain string) bool {
		conn, err := net.DialTimeout("tcp", domain+":80", time.Duration(1)*time.Second)
		if err != nil {
			return false
		} else {
			conn.Close()
			return true
		}
	}
	at.data.Online4 = getNetWork("ipv4.google.com")
	at.data.Online6 = getNetWork("ipv6.google.com")
}

//获取CU/CT/CM 丢包率 5s 检测一次 一次发送5个icmp请求
func (at *Agent) getLostRate() {
	wg := &sync.WaitGroup{}
	getRate := func(dest, mask string) {
		pinger, err := ping.NewPinger(dest)
		if err != nil {
			fmt.Println("ping err", err)
		}
		pinger.Count = 5
		pinger.Interval = time.Duration(200) * time.Millisecond
		pinger.Timeout = time.Duration(1) * time.Second
		pinger.SetPrivileged(true)
		pinger.Run() // Blocks until finished.
		stats := pinger.Statistics()
		switch mask {
		case "cu":
			at.data.LostRateCU = stats.PacketLoss
			at.data.TimeCU = stats.AvgRtt.String()
		case "ct":
			at.data.LostRateCT = stats.PacketLoss
			at.data.TimeCT = stats.AvgRtt.String()
		case "cm":
			at.data.LostRateCM = stats.PacketLoss
			at.data.TimeCM = stats.AvgRtt.String()
		}
		wg.Done()
	}
	wg.Add(3)
	go getRate(at.cu, "cu")
	go getRate(at.ct, "ct")
	go getRate(at.cm, "cm")
	wg.Wait()
}

//获取实时网络流量数据
func (at *Agent) getNetSpeed() {
	t1 := time.Now()
	rx, tx := at.getTrafficInfo()
	time.Sleep(time.Duration(1) * time.Second)
	rx2, tx2 := at.getTrafficInfo()
	el := time.Since(t1)
	at.data.NetworkRx = (rx2 - rx) / (el.Milliseconds() / 1000)
	at.data.NetworkTx = (tx2 - tx) / (el.Milliseconds() / 1000)
}

//获取cpu 1 5 15 负载
func (at *Agent) getLoadAvg() {
	file, err := os.OpenFile("/proc/loadavg", os.O_RDONLY, os.ModeSymlink)
	if err != nil {
		fmt.Println("open uptime file err", err)
		return
	}
	defer file.Close()
	buf := bufio.NewReader(file)
	line, _ := buf.ReadString('\n')
	loadInfo := strings.Fields(string(line))
	at.data.Load1 = at.numberCover(loadInfo[0], "float64").(float64)
	at.data.Load5 = at.numberCover(loadInfo[1], "float64").(float64)
	at.data.Load15 = at.numberCover(loadInfo[2], "float64").(float64)
}

func (at *Agent) numberCover(data, nType string) (value interface{}) {
	switch nType {
	case "int":
		value, _ = strconv.Atoi(data)
		return
	case "float64":
		value, _ = strconv.ParseFloat(data, 32)
		return
	default:
		return
	}
}

//GetData 获取采集数据汇总
func (at *Agent) GetData() Data {
	return at.data
}

func (at *Agent) Start() {
	at.run = make(chan bool, 1)
	at.exit = make(chan int, 2)
	go func(run <-chan bool, exit chan int) {
		defer func() {
			exit <- 0
			fmt.Println("stop ping")
		}()
		for {
			select {
			case <-at.run:
				return
			default:
				at.getLostRate()
			}
			time.Sleep(time.Duration(5) * time.Second)
		}
	}(at.run, at.exit)
	go func(run <-chan bool, exit chan int) {
		defer func() {
			exit <- 0
			fmt.Println("stop query")
		}()
		for {
			select {
			case <-run:
				return
			default:
				at.getUptime()
				at.getMemory()
				at.getDisk()
				at.getCpuUseInfo()
				at.getTrafficStats()
				at.getTUPDCount()
				at.getChinaNetStatus()
				at.getWordNetStatus()
				at.getLostRate()
				at.getNetSpeed()
				at.getLoadAvg()
			}
			time.Sleep(time.Duration(1) * time.Second)
		}
	}(at.run, at.exit)
}
func (at *Agent) Stop() bool {
	close(at.run)
	timeOut := time.After(time.Second * 5)
	for {
		select {
		case <-timeOut:
			fmt.Println("stop timeOut")
			return false
		default:
			if len(at.exit) == 2 {
				close(at.exit)
				at.exit = nil
				return true
			}
		}
	}
}
