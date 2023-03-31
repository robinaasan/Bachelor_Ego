#!/bin/bash
cd orderingservice

rm blockFiles/*
ego-go build
ego sign orderingservice
cd ..
exec bash