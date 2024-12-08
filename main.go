package main

import (
	"entrypoint" //引入接口包（目前未连接只是框架）
	"log"
	"net/http"

	"github.com/casbin/casbin/v2"
)

// casbin接口
type CasbinEnforcer struct {
	enforcer *casbin.Enforcer
}

func NewCasbinEnforcer(modelPath string, policyPath string) (*CasbinEnforcer, error) {
	enforcer, err := casbin.NewEnforcer(modelPath, policyPath)
	if err != nil {
		return nil, err
	}
	return &CasbinEnforcer{enforcer: enforcer}, nil
}

func (ce *CasbinEnforcer) CheckPermission(user string, path string, method string) (bool, error) {
	return ce.enforcer.Enforce(user, path, method)
}

func main() {
	// Casbin配置
	var err error
	casbinConfig, err := NewCasbinEnforcer("model.conf", "policy.csv")
	if err != nil {
		log.Fatalf("Failed to create Casbin enforcer: %v", err)
	}

	// EntryPoint配置-还不确定只是样例
	entryPointConfig := entrypoint.Config{
		TargetURL:      "http://target-service:8000",
		AuthServiceURL: "http://auth-service:8001/verify",
	}
	ep, err := entrypoint.NewEntryPoint(&entryPointConfig)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		user := r.Header.Get("Id") // 假设用户ID在请求头中
		allowed, err := casbinConfig.CheckPermission(user, r.URL.Path, r.Method)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		if !allowed {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		ep.ProxyHandler(w, r)
	})

	log.Printf("Starting server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
