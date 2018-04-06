# Decentralized Tangram
A go client for running a game of Tangram

## Usage
1. Run the program: `go run client.go [-c remoteAddr] [-p rpcPort] [clientAddr]`
1. Navigate to `[clientAddr]` to see the browser client
## Arguments
clientAddr  
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;*required*&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;*default: :8080*&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;The address to access the local browser game  
-c remoteAddr  
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;*optional*&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;Remote peer to connect to  
-p rpcPort  
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;*optional*&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;*default: 9000*&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;Port to use for RPC  
