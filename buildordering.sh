#!/bin/bash
cd orderingservice

#sudo rm files/blockFiles/*
ego-go build
ego sign orderingservice
sudo ego run orderingservice
exec bash