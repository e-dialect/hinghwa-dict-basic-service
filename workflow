
name: Code format check with Ruff

on:
  pull_request:
    branches:
      - master
      - dev
    types: [opened, synchronize, labeled, unlabeled]
  push:
    branches:
      - master

jobs:
  ruff:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v3
        with:
          ref: ${{ github.event.pull_request.head.ref }}
          repository: ${{ github.event.pull_request.head.repo.full_name }}
      - uses: chartboost/ruff-action@v1
        with:
          src: "./src/netspresso_trainer"
          version: 0.0.287
