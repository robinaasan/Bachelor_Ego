#!/bin/bash
for i in {1..100}; do
	go run ./main.go SET $i 9 robin
done
