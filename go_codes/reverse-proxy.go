package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"
)

// ReverseProxy 结构体，用于管理反向代理的目标服务器
type ReverseProxy struct {
	Targets   []*url.URL
	Transport http.RoundTripper
	current   int
	//新增部分
	AuthCheckClient *http.Client // 权限管理模块的客户端
	AuthCheckURL    string       // 权限管理模块的检查URL
}

// 选择目标服务器，轮询算法
func (rp *ReverseProxy) getNextTarget() *url.URL {
	target := rp.Targets[rp.current]
	rp.current = (rp.current + 1) % len(rp.Targets)
	return target
}

// 创建了一个新的方法checkAuthorization，它会构造一个请求发送给权限管理模块，并根据其响应决定是否继续处理原始请求。
// 如果权限管理模块拒绝了请求或返回非成功的状态码，则返回错误。
// 检查权限
func (rp *ReverseProxy) checkAuthorization(req *http.Request) error {
	userID := req.Header.Get("X-User-ID") // 获取用户ID
	if userID == "" {                     //并且如果用户ID缺失，则直接返回错误。这确保了权限管理模块有足够的信息来进行授权决策。
		return fmt.Errorf("missing user ID")
	}
	// 创建一个新的请求发送到权限管理模块
	authReq, err := http.NewRequest("POST", rp.AuthCheckURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create auth request: %v", err)
	}

	// 将原始请求的相关信息添加到权限管理模块请求的头部或body中，例如用户ID、请求路径等
	// 这里假设权限管理模块期望在header中找到这些信息
	authReq.Header.Set("X-Original-User", userID) //将用户ID添加到发送给权限管理模块的请求头部，以便权限管理模块可以根据用户信息进行权限检查。//新增部分，方便联动
	authReq.Header.Set("X-Original-Path", req.URL.Path)
	authReq.Header.Set("X-Original-Method", req.Method)
	// 可以根据实际情况添加更多需要传递的信息

	// 发送请求到权限管理模块
	resp, err := rp.AuthCheckClient.Do(authReq)
	if err != nil {
		return fmt.Errorf("error contacting auth service: %v", err)
	}
	defer resp.Body.Close()

	// 检查权限管理模块的响应状态码
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)                                                                                       // 读取响应体以便记录更多错误信息
		return fmt.Errorf("auth service denied the request with status code: %d, response: %s", resp.StatusCode, string(body)) //在权限检查失败时，不仅检查状态码，还读取权限管理模块的响应体，并将其包含在错误信息中。这有助于更详细地了解权限拒绝的原因，便于调试和问题排查。
	}

	return nil
}

// ServeHTTP 方法，用于处理请求并将其转发到目标服务器
func (rp *ReverseProxy) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// 检查权限
	if err := rp.checkAuthorization(req); err != nil {
		log.Printf("Authorization check failed: %v", err)
		http.Error(w, "Forbidden.", http.StatusForbidden)
		return
	}

	// 如果权限检查通过，则继续处理请求
	target := rp.getNextTarget()
	req.URL.Scheme = target.Scheme
	req.URL.Host = target.Host
	req.RequestURI = "" // 去掉原始请求的 URI

	resp, err := rp.Transport.RoundTrip(req)
	if err != nil {
		log.Printf("Error contacting backend %s: %v", target.Host, err)
		http.Error(w, fmt.Sprintf("Error contacting backend: %v", err), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	for key, value := range resp.Header {
		w.Header()[key] = value
	}
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func main() {
	// 配置多个后端服务器
	targetURLs := []*url.URL{
		&url.URL{Scheme: "http", Host: "127.0.0.1:8081"},
	}

	// 权限管理模块的URL和HTTP客户端配置
	//根据github库里面负责权限管理同学写的代码 权限检查的API端点是在http://auth-service:8001/verify
	authCheckURL := "http://auth-service:8001/verify" // 更新为权限管理模块的实际地址
	authCheckClient := &http.Client{
		Timeout: 5 * time.Second, // 设置5秒的超时时间
	}
	// 创建带有超时设置的HTTP客户端

	// 创建一个反向代理实例
	proxy := &ReverseProxy{
		Targets:         targetURLs,
		Transport:       http.DefaultTransport,
		AuthCheckClient: authCheckClient,
		AuthCheckURL:    authCheckURL,
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
