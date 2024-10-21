on:
  push:
    branches: [ workflow ]
  pull_request:
    branches: [ workflow ]
steps:
  - name: Access secret
    env:
      API_KEY: ${{ secrets.API_KEY }}
    run: |
      # Use the API_KEY environment variable
      echo "API Key: $API_KEY"
steps:
  - name: Configure AWS credentials
    uses: aws-actions/configure-aws-credentials@v1
    with:
      aws-access-key-id: ${{ secrets.AWS_ACCESS_KEY_ID }}
      aws-secret-access-key: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
      aws-region: us-east-1
env:
  API_URL: https://api.example.com

jobs:
  build:
    env:
      DEBUG: true
    steps:
      - name: Build
        run: |
          # Access environment variables
          echo "API URL: $API_URL"
          echo "Debug mode: $DEBUG"
steps:
  - name: Deploy
    env:
      API_KEY: ${{ secrets.PROD_API_KEY }}
    run: |
      # Use the environment-specific API key
      echo "API Key: $API_KEY"
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v2
    - name: Set up Go
      uses: actions/setup-go@v2
      with:
        go-version: 1.17
    - name: Build
      run: go build -v ./...
    - name: Test
      run: go test -v ./...
steps:
  - uses: actions/checkout@v2
  - uses: actions/setup-go@v2
    with:
      go-version: 1.17
  - uses: actions/cache@v2
    with:
      path: |
        ~/.cache/go-build
        ~/go/pkg/mod
      key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
      restore-keys: |
        ${{ runner.os }}-go-
  - name: Build
    run: go build -v ./...
  - name: Test
    run: go test -v ./...
steps:
  - uses: actions/checkout@v2
  - name: Set up Go
    uses: actions/setup-go@v2
    with:
      go-version: 1.17
  - name: Run golint
    run: golint ./...
  - name: Run gofmt
    run: gofmt -d ./
  - name: Run go vet
    run: go vet ./...
steps:
  - uses: actions/checkout@v2
  - name: Set up Go
    uses: actions/setup-go@v2
    with:
      go-version: 1.17
  - name: Run tests with coverage
    run: go test -coverprofile=coverage.out ./...
  - name: Upload coverage to Codecov
    uses: codecov/codecov-action@v2
    with:
      file: ./coverage.out
