#!/bin/bash
cd orderingservice

sudo rm blockFiles/*
ego-go build
ego sign orderingservice
cd ..
exec bash