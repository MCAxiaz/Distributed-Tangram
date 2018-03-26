const EventType = {
    Subscribe: 0,
    GetState: 1,
    UpdateState: 2,
    Unsubscribe: 3
};

var gameState = {
    tans: new Map()
};

function openSocket() {
    var socket = new WebSocket(`ws://${location.host}/ws`);
    socket.addEventListener("open", function (e) {
        console.log(`[Socket] Connected to ${socket.url}`);
    });
    socket.addEventListener("message", function (e) {
        console.log("[Socket] Message\n", e.data);
    });
    socket.addEventListener("error", function (e) {
        console.error(e);
    });
    return socket;
}

document.addEventListener("DOMContentLoaded", function (e) {
    var socket = openSocket();
    socket.addEventListener("message", function (e) {
        var view = document.getElementById("view");
        view.innerHTML = e.data;
        var paths = document.getElementsByTagName("path");
        for (var path of paths) {
            // Parse numbers from transform attribute
            gameState.tans.set(path.id, parseTransform(path.getAttribute("transform")));

            path.addEventListener("mousedown", function (e) {
                console.log("Holding on to tan");
                var tan = gameState.tans.get(path.id);

                path.addEventListener("mousemove", function (e) {
                    path.setAttribute("transform", "translate(" + tan.X + ", " + tan.Y + ") rotate(" + tan.R + ")");
                });
            });

            path.addEventListener("mouseup", function (e) {
                console.log("Releasing tan");
            });

            path.addEventListener("x", function (e) {

            });
            
            path.addEventListener("z", function (e) {

            });
        }
    });
    socket.addEventListener("open", function (e) {
        socket.send("");
    });
});

function getMousePosition() {
    var x = window.event.clientX;
    var y = window.event.clientY;
    return {x, y};
}

function parseTransform(str) {
    var coordinates = str.match(/\d+/g) ? str.match(/\d+/g) : [];
    return {X: coordinates[0], Y: coordinates[1], R: coordinates[2]};
}