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

function renderTan(model, node) {
    var id = `tan-${model.id}`
    var transform = `translate(${model.location.x}, ${model.location.y}) rotate(${model.rotation})`;
    var d = "";
    model.shape.points.forEach(function(point, i) {
        var command = i == 0 ? "M" : "L";
        d += `${command} ${point.x} ${point.y} `;
    });
    d += "Z";
    
    node.id = id;
    node.setAttribute('fill', model.shape.fill);
    node.setAttribute('fill-opacity', '1'); // fill-opacity's value will be halved when tan is possessed
    node.setAttribute('stroke', model.shape.stroke);
    node.setAttribute('transform', transform);
    node.setAttribute('d', d);

    return node
}

function attachPlayerNameOnTan(tan) {
    var owner = document.createElement("h4");
    owner.textContent = "";
    var div = document.createElement("div");
    div.className = "owner";
    div.appendChild(owner);
    tan.appendChild(div);

    return tan;
}

var socket;
var config;
var state;
document.addEventListener("DOMContentLoaded", function(e) {
    var view = document.getElementById("view");
    var dump = document.getElementById("dump");

    function getTan(model) {
        var tan = view.getElementById(`tan-${model.id}`);
        if (!tan) {
            tan = document.createElementNS(view.namespaceURI, "path");
            renderTan(model, tan);
            attachPlayerNameOnTan(tan);
            view.appendChild(tan);
            tan.addEventListener("mousedown", onMouseDown)
        }
        return tan;
    }

    function render(state) {
        for (let tan of state.tans) {
            let node = getTan(tan);
            renderTan(tan, node);
        }
    }

    // lockTan objectives
    // - set player name on tan
    // - highlight the tan to indicate someone has possession of it
    function lockTan(tanID, playerName) {
        var tan = view.getElementById(`tan-${tanID}`);
        
        if (!tan) {
            console.log("You are grabbing a tan that does not exist.");
            return;
        }

        // TODO: Check if tan is already locked
        
        // Change path's fill opacity to half
        tan.setAttribute("fill-opacity", "0.5");

        // TODO: Display player name on tan and set it to locked
        // Look for any indication of a tan being locked on the frontend

        console.log(`[Lock tan] Tan ${tanID}: I am possessed by ${playerName}.`);
        socket.send(JSON.stringify({
            type: "LockTan",
            tan: tan.id,
            lockTan: true
        }));
    }

    function unlockTan(tanID, playerID) {
        var tan = view.getElementById(`tan-${tanID}`);

        if (!tan) {
            console.log("Cannot unlock a tan that does not exist.");
            return;
        }

        // TODO: Set tan to unlocked

        // Restore path's fill opacity to 1
        tan.setAttribute("fill-opacity", "1");
        
        // TODO: Remove player name from tan

        console.log(`[Unlock tan] ${tanID}`);
        socket.send(JSON.stringify({
            type: "LockTan",
            tan: tan.id,
            lockTan: false
        }));
    }

    socket = openSocket();
    socket.addEventListener("message", function(e) {
        dump.innerHTML = e.data
        var message = JSON.parse(e.data)
        switch (message.type) {
        case "state":
            state = message.data
            render(state);
            break;
        case "config":
            config = message.data;
            view.setAttribute("width", config.Size.x)
            view.setAttribute("height", config.Size.y)
            break;
        }
    });
    socket.addEventListener("open", function(e) {
        socket.send(JSON.stringify({
            type:"GetState"
        }));
    })

    function onMouseDown(e) {
        var path = e.target;
        var id = parseInt(path.id.match(/tan-(\d+)/)[1]);
        console.log(`Holding on to tan id=${id}`);

        var tan = state.tans.find(function(tan) {
            return tan.id == id
        });
        
        var startTanPos = {
            x: tan.location.x,
            y: tan.location.y,
            r: tan.rotation
        };

        var startMousePos = {
            x: e.clientX,
            y: e.clientY
        };

        var mouseMoveListener = function (e) {
            tan.location.x = startTanPos.x + (e.clientX - startMousePos.x);
            tan.location.y = startTanPos.y + (e.clientY - startMousePos.y);
            renderTan(tan, path);
            socket.send(JSON.stringify({
                type: "MoveTan",
                tan: tan.id,
                location: tan.location,
                rotation: tan.rotation
            }));
        };
        document.addEventListener("mousemove", mouseMoveListener);

        // Rotate tan clockwise or counter-clockwise
        var rotateListener = function (e) {
            var key = e.code;
            var d = 0;
            switch (key) {
            case "KeyX":
                d = 1;
                break;
            case "KeyZ":
                d = -1
                break;
            }
            if (d) {
                console.log(`[rotate] ${key}`);
                tan.rotation = rotate(tan.rotation, d);
                renderTan(tan, path);
            }
            socket.send(JSON.stringify({
                type: "MoveTan",
                tan: tan.id,
                location: tan.location,
                rotation: tan.rotation
            }));
        };
        document.addEventListener("keypress", rotateListener);

        document.addEventListener("mouseup", function(e) {
            console.log(`Releasing tan id=${id}`);
            document.removeEventListener("mousemove", mouseMoveListener);
            document.removeEventListener("keypress", rotateListener);

            socket.send(JSON.stringify({
                type: "ObtainTan",
                tan: id,
                release: true
            }));
        }, {
            once:true
        });

        socket.send(JSON.stringify({
            type: "ObtainTan",
            tan: id,
            release: false
        }));
    }
})

function rotate(r, d) {
    return (r + d * 5 + 720) % 360;
}
