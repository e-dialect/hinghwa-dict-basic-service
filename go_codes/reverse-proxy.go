package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
)

// ReverseProxy 结构体，用于管理反向代理的目标服务器
type ReverseProxy struct {
	Targets   []*url.URL
	Transport http.RoundTripper
	current   int
}

// 选择目标服务器，轮询算法
func (rp *ReverseProxy) getNextTarget() *url.URL {
	target := rp.Targets[rp.current]
	rp.current = (rp.current + 1) % len(rp.Targets)
	return target
}

// ServeHTTP 方法，用于处理请求并将其转发到目标服务器
func (rp *ReverseProxy) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// 选择目标服务器
	target := rp.getNextTarget()

	// 修改请求的 URL，将请求重定向到目标服务器
	req.URL.Scheme = target.Scheme
	req.URL.Host = target.Host
	req.RequestURI = "" // 去掉原始请求的 URI

	// 使用指定的 Transport 来发送请求到后端服务器
	resp, err := rp.Transport.RoundTrip(req)
	if err != nil {
		// 错误处理
		log.Printf("Error contacting backend %s: %v", target.Host, err)
		http.Error(w, fmt.Sprintf("Error contacting backend: %v", err), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// 将后端服务器的响应复制到客户端
	for key, value := range resp.Header {
		w.Header()[key] = value
	}
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

// main 函数，设置反向代理并启动 HTTP 服务器
func main() {
	// 配置多个后端服务器
	targetURLs := []*url.URL{
		// 请替换成你实际的目标服务器地址
		&url.URL{Scheme: "http", Host: "127.0.0.1:8081"},
	}

	// 创建一个反向代理实例
	proxy := &ReverseProxy{
		Targets:   targetURLs,
		Transport: http.DefaultTransport,
	}

	// 设置 HTTP 路由，所有请求都交给反向代理处理
	http.Handle("/", proxy)

	// 启动服务器
	port := ":8080"
	fmt.Printf("Starting reverse proxy on %s...\n", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
