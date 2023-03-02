# Bachelor_Ego
Bachelor thesis for Robin and Iver

## Installation

### Golang
Istall golang: [go](https://go.dev/doc/install#tarball_non_standard) 

```bash
rm -rf /usr/local/go && tar -C /usr/local -xzf go1.20.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
go version
```

### Ego
Install Ego: [Ego](https://docs.edgeless.systems/ego/getting-started/install)

```bash
sudo snap install ego-dev --classic
sudo apt install build-essential libssl-dev
```

### Wasmer
Download wasmer-go: 
```bash
wget -O- https://github.com/wasmerio/wasmer/releases/download/2.2.1/wasmer-linux-amd64.tar.gz | tar xz --one-top-level=wasmer
```
Tell go compiler to use it:
```bash
CGO_CFLAGS="-I$PWD/wasmer/include" CGO_LDFLAGS="$PWD/wasmer/lib/libwasmer.a -ldl -lm -static-libgcc" ego-go build -tags custom_wasmer_runtime
```

### enclave.json
```json
{
 "exe": "server",
 "key": "private.pem",
 "debug": true,
 "heapSize": 512,
 "executableHeap": true,
 "productID": 1,
 "securityVersion": 1,
 "mounts": [
        {
            "source": "/home/stud/robin/Bachelor_Ego/server",
            "target": "/data",
            "type": "hostfs",
            "readOnly": false
        }
           ]
}
```
## Usage
How to use the application...



