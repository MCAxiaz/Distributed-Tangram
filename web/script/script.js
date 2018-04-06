function openSocket() {
    var socket = new WebSocket(`ws://${location.host}/ws`);
    socket.addEventListener("open", function (e) {
        console.log(`[Socket] Connected to ${socket.url}`);
    });
    socket.addEventListener("message", function (e) {
        //console.log("[Socket] Message\n", e.data);
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
    // TODO: Check if tan is locked first, then decide on the fill opacity value
    node.setAttribute('stroke', model.shape.stroke);
    if (model.Matched) {
      node.setAttribute('stroke', 'green');
    }

    node.setAttribute('transform', transform);
    node.setAttribute('d', d);
    node.setAttribute('class', 'draggable');
    return node
}

// Attaches textPath to SVG for player's name
function attachPlayerNameTextToSVG(tanID, playerName) {
    var svg = document.getElementById("view");
    var use = document.createElementNS(view.namespaceURI, "use");

    use.setAttribute("href", `#${tanID}`);
    var txt = document.createElementNS(view.namespaceURI, "text");
    txt.setAttribute("font-family", "Verdana");
    txt.setAttribute("font-size", "42.5");
    var txtPath = document.createElementNS(view.namespaceURI, "textPath");
    txtPath.setAttribute("href", `#${tanID}`);

    if (playerName) {
        txtPath.id = `txtPath-${tanID}-${playerName}`;
        txtPath.innerText(playerName);
    } else {
        txtPath.id = `txtPath-${tanID}`;
        txtPath.innerText = "";
    }

    txt.appendChild(txtPath);
    svg.appendChild(use);
    svg.appendChild(txt);
}

function renderTargetTan(model, offset, node) {
    var transform = `translate(${model.location.x + offset.x}, ${model.location.y + offset.y}) rotate(${model.rotation})`;
    var d = "";
    model.shape.points.forEach(function(point, i) {
        var command = i == 0 ? "M" : "L";
        d += `${command} ${point.x} ${point.y} `;
    });
    d += "Z";

    node.setAttribute('fill', 'grey');
    node.setAttribute('stroke', 'grey');
    node.setAttribute('stroke-width', 2);
    node.setAttribute('stroke-linejoin', 'round');
    node.setAttribute('transform', transform);
    node.setAttribute('d', d);
    return node
}

var socket;
var config;
var state;
var player;
document.addEventListener("DOMContentLoaded", function(e) {
    var view = document.getElementById("view");
    var defs = document.createElementNS(view.namespaceURI, "defs");
    defs.id = "defs";
    //view.appendChild(defs);
    var timer = document.getElementById("timer");
    var dump = document.getElementById("dump");

    function getTan(model) {
        var tan = view.getElementById(`tan-${model.id}`);
        //var defs = view.getElementById("defs");
        if (!tan) {
            tan = document.createElementNS(view.namespaceURI, "path");
            renderTan(model, tan);
            attachPlayerNameTextToSVG(tan.id, tan.playerName);
            view.appendChild(tan);
            tan.addEventListener("pointerdown", onMouseDown)
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
    // returns true if tan is successfully locked, false if not
    function lockTan(tanID) {
        var tan = view.getElementById(`tan-${tanID}`);

        if (!tan) {
            console.log("You are grabbing a tan that does not exist.");
            return false;
        }

        // TODO: Check if tan is already locked
        
        // Change path's fill opacity to half
        tan.setAttribute("fill-opacity", "0.5");

        // Display player name on tan
        var txtPath = document.getElementById(`txtPath-tan-${tanID}`);
        if (!txtPath) {
            console.log(`No such txtPath with tan ${tanID}`);
            return false;
        }
        
        txtPath.innerText = player.ID;
        txtPath.id = `txtPath-tan-${tanID}-${player.ID}`;
        
        console.log(`[Lock tan] Tan ${tanID}: I am possessed by ${player.ID}.`);

        return true;
    }

    function unlockTan(tanID) {
        var tan = view.getElementById(`tan-${tanID}`);

        if (!tan) {
            console.log("Cannot unlock a tan that does not exist.");
            return false;
        }

        // TODO: Set tan to unlocked

        // Restore path's fill opacity to 1
        tan.setAttribute("fill-opacity", "1");
        
        // Remove player name from tan
        var txtPath = document.getElementById(`txtPath-tan-${tanID}-${player.ID}`);
        if (!txtPath) {
            console.log(`No such txtPath with tan ${tanID} and player ${player.ID}`);
            return false;
        }

        txtPath.innerText = "";
        txtPath.id = `txtPath-tan-${tanID}`;

        console.log(`[Unlock tan] ${tanID}`);

        return true;

    }

    function renderTarget(config) {
      for (let ttan of config.targets) {
        let node = document.createElementNS(view.namespaceURI, "path");
        renderTargetTan(ttan, config.Offset, node)
        view.appendChild(node);
      }
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
            renderTarget(config);
            break;
        case "player":
            player = message.data;
            break;
        }
    });
    socket.addEventListener("open", function(e) {
        socket.send(JSON.stringify({
            type:"GetState"
        }));
    })

    setInterval(function() {
        if (state) {
            var d = Date.now() - new Date(state.Timer).getTime()
            timer.innerHTML = Math.round(d / 1000)
        }
    }, 100)

    function onMouseDown(e) {
        var path = e.target;
        var id = parseInt(path.id.match(/tan-(\d+)/)[1]);

        var tan = state.tans.find(function(tan) {
            return tan.id == id
        });

        // TODO: The player name used here is incorrect.
        // Is there a data structure on the frontend that stores the player who is playing the game on that node?
        var locked = lockTan(tan.id);
        if (!locked) {
            return;
        }

        console.log(`Holding on to tan id=${id}`);
        
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
            tan.location.x = Math.round(Math.max(0, Math.min(startTanPos.x + (e.clientX - startMousePos.x), config.Size.x)));
            tan.location.y = Math.round(Math.max(0, Math.min(startTanPos.y + (e.clientY - startMousePos.y), config.Size.y)));
            renderTan(tan, path);
            socket.send(JSON.stringify({
                type: "MoveTan",
                tan: tan.id,
                location: tan.location,
                rotation: tan.rotation
            }));
        };
        document.addEventListener("pointermove", mouseMoveListener);

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

        document.addEventListener("pointerup", function(e) {
            var unlock = unlockTan(id);
            if (!unlock) {
                console.log(`Error encountered while unlocking tan ${id}`);
            }
            console.log(`Releasing tan id=${id}`);
            document.removeEventListener("pointermove", mouseMoveListener);
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
    return (r + d * 15 + 720) % 360;
}
