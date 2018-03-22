const EventType = {
    Subscribe: 0,
    GetState: 1,
    UpdateState: 2,
    Unsubscribe: 3,
}

function openSocket() {
    var socket = new WebSocket(`ws://${location.host}/ws`);
    socket.addEventListener("open", function(e) {
        console.log(`[Socket] Connected to ${this.url}`);
    });
    socket.addEventListener("message", function(e) {
        console.log(`[Socket] Message\n${e.data}`);
    });
    socket.addEventListener("error", function(e) {
        console.error(e);
    });
    return socket;
}

document.addEventListener("DOMContentLoaded", function(e) {
    var socket = openSocket();
    socket.addEventListener("message", function(e) {
        var view = document.getElementById("view");
        view.innerHTML = e.data
    });
    socket.addEventListener("open", function(e) {
        socket.send("");
    })
    
})
