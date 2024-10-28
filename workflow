name: Go Lint

# 触发条件
on:
  # 当推送到 main 分支时触发
  push:
    branches: [ main ]
  # 当对任何分支发起 pull request 时触发
  pull_request:
    branches: [ '**' ]

# 定义工作流程
jobs:
  golangci:
    name: lint
    runs-on: ubuntu-latest
    
    steps:

      - uses: actions/checkout@v3


      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21' 
          cache: true
      

      - name: Install dependencies
        run: go mod download
      

      - name: Run golangci-lint
        uses: golangci/golangci-lint-action@v3
        with:
          version: latest
          args: --timeout=5m
          

      - name: Check code formatting
        run: |
          if [ "$(gofmt -s -l . | wc -l)" -gt 0 ]; then
            echo "Code is not properly formatted. Please run 'go fmt'"
            gofmt -s -l .
            exit 1
          fi

      - name: Run go vet
        run: go vet ./...
