# Decentralized Tangram
A go client for running a game of Tangram

## Running on Azure
Run the following commands after ssh-ing into the server of choice.
```
$ cd tan
$ kill $(cat pid) 
$ go build
$ nohup ./tan > web/application.log 2>&1 </dev/null &
$ echo $! > pid
```

If the application is connecting to a server, run `nohup ./tan -c <server> > web/application.log 2>&1 </dev/null &` instead.

## Usage
1. Run the program: `go run client.go [-c remoteAddr] [-p rpcPort] [clientAddr]`
1. Navigate to `[clientAddr]` to see the browser client
## Arguments
clientAddr  
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;*required*&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;*default: :8080*&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;The address to access the local browser game  
-c remoteAddr  
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;*optional*&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;Remote peer to connect to  
-p rpcAddress  
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;*optional*&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;*default: :9000*&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;Address to use for RPC.  
-i identifier  
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;*optional*&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;*default: 0*&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;Id to use for this client. 0 will randomize.  
