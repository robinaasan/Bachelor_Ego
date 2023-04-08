#!/bin/bash
cd client
#Set environment variables in go for the eclient package
go env -w CGO_CFLAGS=-I/snap/ego-dev/current/opt/ego/include
go env -w CGO_LDFLAGS=-L/snap/ego-dev/current/opt/ego/lib

cd ../runtime
sudo rm secret.store
#set 
CGO_CFLAGS="-I$PWD/wasmer/include" CGO_LDFLAGS="$PWD/wasmer/lib/libwasmer.a -ldl -lm -static-libgcc" ego-go build -tags custom_wasmer_runtime

wait
ego sign runtime
# uniqueid=$(ego uniqueid runtime)
# export UNIQUEID=$uniqueid 
ego uniqueid runtime
cd ..
exec bash