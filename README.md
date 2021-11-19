# fabric-sdk-tutorial

## About this branch

**This branch is for key-switch process performance test only.**

This branch is a fork from the `develop` branch, but different in that:  
1. More timestamp logs in `internal/background/key_switch_server.go` and the logs are written to several `*.out` files.
2. A simple tool (in `cmd/timestampstat`) that analyzes the `*.out` files to get info about  
  2.1 The overall execution time consumptions of each sub-process.  
  2.2 The average execution time consumptions of each sub-process.
3. Some helper functions in `internal/utils/timingutils/timer.go`.

## Usage

A common usage is to checkout several files in this branch from the branch you're working on.  
For example, say you're planning to test the key-switch performance based on a running environment of an 8-user setup. Here's the steps you can follow:
1. Suppose you're in branch `8-users-setup` and that `git status` is clean.
2. Run `git checkout perf-ppks-overall cmd/timestampstat internal/background/key_switch_server.go internal/utils/timingutils/timer.go`.
3. Perform the tests and DON'T apply the changes.
