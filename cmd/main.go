package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/process"
)

type Metrics struct {
	TID         string        `json:"tid"`
	CID         string        `json:"cid"`
	Key         string        `json:"key"`
	Timestamp   string        `json:"timestamp"`
	CPUUsage    float64       `json:"cpuUsagePercent"`
	MemoryUsed  uint64        `json:"memoryUsedMb"`
	MemoryTotal uint64        `json:"memoryTotalMb"`
	DiskUsed    uint64        `json:"diskUsedGb"`
	DiskTotal   uint64        `json:"diskTotalGb"`
	Processes   []ProcessInfo `json:"processList"`
}

type ProcessInfo struct {
	Pid         int32   `json:"pid"`
	Name        string  `json:"name"`
	CPU         float64 `json:"cpu"`
	Memory      float32 `json:"memory"`
	CommandLine string  `json:"commandLine"`
}

type SlackPayload struct {
	Text string `json:"text"`
}

var cid string
var max float64
var webhookUrl string

func main() {

	err := godotenv.Load()
	if err != nil {
		fmt.Println(".env 파일 로드를 실패하였습니다. :", err)
		return
	}

	cid = os.Getenv("CID")
	webhookUrl = os.Getenv("WEBHOOK_URL")
	max, _ = strconv.ParseFloat(os.Getenv("MAX"), 64)

	for {
		metrics, err := collectMetrics()

		if err != nil {
			panic(err)
		}
		if metrics.CPUUsage > max {
			SendSlackAlert(metrics)
		}
		time.Sleep(5 * time.Second)
	}
}

func collectMetrics() (Metrics, error) {
	// CPU 사용률
	// Percent 메서드의 두 번째 인자는 전체 코어를 기준으로 사용률을 계살한 건지 or not
	cpuPercent, err := cpu.Percent(1*time.Second, false)
	if err != nil {
		panic(err)
	}

	cpuPercent[0] = math.Round(cpuPercent[0]*10) / 10

	// Memory 사용량
	memStats, err := mem.VirtualMemory()
	if err != nil {
		panic(err)
	}

	// 루트 기준 Disk 사용량
	diskStats, err := disk.Usage("/")
	if err != nil {
		panic(err)
	}

	processList := getProcess()

	metrics := Metrics{
		Timestamp:   time.Now().Format(time.RFC3339),
		CPUUsage:    cpuPercent[0],
		MemoryUsed:  memStats.Used / 1024 / 1024,          // MB 단위
		MemoryTotal: memStats.Total / 1024 / 1024,         // MB 단위
		DiskUsed:    diskStats.Used / 1024 / 1024 / 1024,  // GB 단위
		DiskTotal:   diskStats.Total / 1024 / 1024 / 1024, // GB 단위
		Processes:   processList,
	}

	return metrics, nil
}

func getProcess() []ProcessInfo {
	processes, err := process.Processes()
	if err != nil {
		return nil
	}

	var processList []ProcessInfo

	for _, p := range processes {
		name, _ := p.Name()

		cpuPercent, err := p.CPUPercent()
		if err != nil {
			continue
		}

		memory, err := p.MemoryPercent()

		if err != nil {
			continue
		}

		processList = append(processList, ProcessInfo{
			Pid:    p.Pid,
			Name:   name,
			CPU:    cpuPercent,
			Memory: memory,
		})
	}

	// CPU 사용률 기준으로 내림차순 정렬
	sort.Slice(processList, func(i, j int) bool {
		return processList[i].CPU > processList[j].CPU
	})

	for i := 0; i < len(processList); i++ {
		p, _ := process.NewProcess(processList[i].Pid)
		processList[i].CommandLine, _ = p.Cmdline()
	}

	return processList[:5]
}

func SendSlackAlert(metrics Metrics) error {
	var blocks []map[string]interface{}

	blocks = append(blocks, map[string]interface{}{
		"type": "section",
		"text": map[string]string{
			"type": "mrkdwn",
			"text": fmt.Sprintf("*🚨 [%s] Top CPU Consuming Processes 🚨*", cid),
		},
	})

	blocks = append(blocks, map[string]interface{}{
		"type": "divider",
	})

	blocks = append(blocks, map[string]interface{}{
		"type": "section",
		"text": map[string]string{
			"type": "mrkdwn",
			"text": "Please consider cleaning up or reviewing the following processes.",
		},
	})

	blocks = append(blocks, map[string]interface{}{
		"type": "divider",
	})

	for i, p := range metrics.Processes {
		block := map[string]interface{}{
			"type": "section",
			"text": map[string]string{
				"type": "mrkdwn",
				"text": fmt.Sprintf("*%d.* `%s` CPU - *%.2f%%* | Memory - *%.2f%%*", i+1, p.Name, p.CPU, p.Memory),
			},
		}
		blocks = append(blocks, block)
	}

	// JSON payload 구성
	payload := map[string]interface{}{
		"blocks": blocks,
	}

	jsonBody, err := json.Marshal(payload)

	if err != nil {
		return err
	}

	resp, err := http.Post(webhookUrl, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}
