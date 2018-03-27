const EventType = {
    Subscribe: 0,
    GetState: 1,
    UpdateState: 2,
    Unsubscribe: 3
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
            path.addEventListener("mousedown", function (e) {
                console.log("Holding on to tan");
                // Parse numbers from transform attribute
                var currentTanPos = parseTransform(path.getAttribute("transform")); // returns { X, Y, R }
                var startMousePos = getMousePosition();

                var listener = function (e) {
                    var currentMousePosition = {
                        X: window.event.clientX - startMousePos.X,
                        Y: window.event.clientY - startMousePos.Y,
                    };

                    path.setAttribute("transform", "translate(" +
                        (currentTanPos.X + currentMousePosition.X) + ", " +
                        (currentTanPos.Y + currentMousePosition.Y)
                        + ") rotate(" + currentTanPos.R + ")");
                };
                path.addEventListener("mousemove", listener);

                // Rotate tan clockwise or counter-clockwise
                // keyCode: x = 88, z = 90
                var rotateListener = function (e) {
                    console.log("rotate");
                    if (e.keyCode == 88) {
                        console.log("rotate clockwise");
                        console.log("key", e.keyCode);
                        path.setAttribute("transform", "translate(" +
                            currentTanPos.X + ", " +
                            currentTanPos.Y
                            + ") rotate(" + rotateCW(currentTanPos.R) + ")");
                    } else if (e.keyCode == 90) {
                        console.log("rotate counterclockwise");
                        console.log("key", e.keyCode);
                        path.setAttribute("transform", "translate(" +
                            currentTanPos.X + ", " +
                            currentTanPos.Y
                            + ") rotate(" + rotateCCW(currentTanPos.R) + ")");
                    }
                };
                path.addEventListener("keypress", rotateListener);

                document.addEventListener("mouseup", function(e) {
                    console.log("Releasing tan");
                    path.removeEventListener("mousemove", listener);
                    path.removeEventListener("keypress", rotateListener);
                }, {
                    once:true
                });
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
    return {X: x, Y: y};
}

function parseTransform(str) {
    var coordinates = str.match(/\d+/g) ? str.match(/\d+/g) : [];
    return {X: parseInt(coordinates[0]), Y: parseInt(coordinates[1]), R: parseInt(coordinates[2])};
}

function rotateCW(n) {
    return (n + 5) % 360;
}

function rotateCCW(n) {
    return (n - 5) % 360;
}