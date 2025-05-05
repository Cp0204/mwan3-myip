package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type IPResponse struct {
	IP   string `json:"ip"`
	Port int    `json:"port"`
}

type PortsResponse struct {
	Ports []int `json:"ports"`
}

func loadEnv() {
	// 读取环境变量文件
	envContent, err := os.ReadFile(".env")
	if err != nil {
		fmt.Println("Warning: .env not found, using default values")
		return
	}

	// 解析环境变量
	lines := strings.SplitSeq(string(envContent), "\n")
	for line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			os.Setenv(parts[0], parts[1])
		}
	}
}

func getCheckPorts() []int {
	portsStr := os.Getenv("CHECK_PORTS")
	if portsStr == "" {
		return []int{80}
	}

	portStrs := strings.Split(portsStr, ",")
	ports := make([]int, 0, len(portStrs))
	for _, portStr := range portStrs {
		port, err := strconv.Atoi(strings.TrimSpace(portStr))
		if err == nil {
			ports = append(ports, port)
		}
	}
	return ports
}

func getIP(r *http.Request) string {
	// 尝试从 X-Real-IP 获取
	ip := r.Header.Get("X-Real-IP")
	if ip != "" {
		return ip
	}

	// 尝试从 X-Forwarded-For 获取
	ip = r.Header.Get("X-Forwarded-For")
	if ip != "" {
		// 取第一个IP
		ips := strings.Split(ip, ",")
		return strings.TrimSpace(ips[0])
	}

	// 从 RemoteAddr 获取
	ip = r.RemoteAddr
	// 移除端口号
	if strings.Contains(ip, ":") {
		ip = strings.Split(ip, ":")[0]
	}
	return ip
}

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	(*w).Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func main() {
	// 加载环境变量
	loadEnv()

	// 获取配置的端口列表
	checkPorts := getCheckPorts()

	// 设置静态文件服务
	fs := http.FileServer(http.Dir("public"))
	http.Handle("/", fs)

	// 设置API端点
	http.HandleFunc("/api/ip", func(w http.ResponseWriter, r *http.Request) {
		enableCors(&w)
		if r.Method == "OPTIONS" {
			return
		}
		response := IPResponse{
			IP:   getIP(r),
			Port: 80, // 由于Docker端口映射，这里固定为80
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	// 添加获取端口列表的API端点
	http.HandleFunc("/api/ports", func(w http.ResponseWriter, r *http.Request) {
		enableCors(&w)
		if r.Method == "OPTIONS" {
			return
		}
		response := PortsResponse{
			Ports: checkPorts,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	// 启动服务器
	port := os.Getenv("PORT")
	if port == "" {
		port = "800"
	}
	fmt.Printf("Server starting on port %s\n", port)
	http.ListenAndServe(":"+port, nil)
}
