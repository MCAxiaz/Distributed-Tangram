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

const NO_PLAYER = -1;

function renderTan(model, path, txtPath) {
    var transform = `translate(${model.location.x}, ${model.location.y}) rotate(${model.rotation})`;
    var d = "";
    model.shape.points.forEach(function (point, i) {
        var command = i == 0 ? "M" : "L";
        d += `${command} ${point.x} ${point.y} `;
    });
    d += "Z";

    path.id = `tan-${model.id}`;
    path.setAttribute('fill', model.shape.fill);
    if (model.Matched) {
        path.setAttribute('stroke', 'green');
    } else {
        path.setAttribute('stroke', model.shape.stroke);
    }
    path.setAttribute('transform', transform);
    path.setAttribute('d', d);

    if (model.player !== NO_PLAYER) {
        // Render player ID to tan
        path.classList.add("locked")
        txtPath.innerHTML = model.player;
    } else {
        path.classList.remove("locked")
        txtPath.innerHTML = "";
    }
    return path;
}

function renderTargetTan(model, offset, node) {
    var transform = `translate(${model.location.x + offset.x}, ${model.location.y + offset.y}) rotate(${model.rotation})`;
    var d = "";
    model.shape.points.forEach(function (point, i) {
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
document.addEventListener("DOMContentLoaded", function (e) {
    var view = document.getElementById("view");
    var gPath = document.getElementById("g-path");
    var gText = document.getElementById("g-text");
    var timer = document.getElementById("timer");
    var dump = document.getElementById("dump");
    
    function getTan(id) {
        var model = state.tans.find(function (tan) {
            return tan.id == id
        });

        if (!model) {
            // The tan we are trying to find does not exist
            return null
        }

        var path = view.getElementById(`tan-${id}`);
        var text = view.getElementById(`txtPath-${id}`);
        if (!path) {
            var result = initializeTan(id);
            path = result.path;
            text = result.text;
        }

        return {model, path, text};
    }

    function render(state) {
        for (let tan of state.tans) {
            let {model, path, text} = getTan(tan.id);
            renderTan(model, path, text);
        }
    }

    // lockTan objectives
    // - set player name on tan
    // - highlight the tan to indicate someone has possession of it
    // returns true if tan is successfully locked, false if not
    function lockTan(tanID) {
        var {model, path, text} = getTan(tanID)

        if (model.player !== NO_PLAYER && model.player !== player.ID) {
            console.log(`Another player ${model.player} is already holding onto the tan.`);
            return false;
        }

        model.player = player.ID;
        renderTan(model, path, text);

        console.log(`[Lock tan] Tan ${tanID}: I am possessed by ${player.ID}.`);
        return true;
    }

    function unlockTan(tanID) {
        var {model, path, text} = getTan(tanID)

        if (model.player !== player.ID) {
            console.log(`Another player ${model.player} is already holding onto the tan.`);
            return false;
        }

        model.player = NO_PLAYER;
        renderTan(model, path, text);

        console.log(`[Unlock tan] ${tanID}`);
        return true;

    }

    function renderTarget(config) {
        for (let ttan of config.targets) {
            let node = document.createElementNS(view.namespaceURI, "path");
            var gTarget = document.getElementById("g-target");
            renderTargetTan(ttan, config.Offset, node)
            gTarget.appendChild(node);
        }
    }

    function renderGroups() {
        var view = document.getElementById("view");
    }

    socket = openSocket();
    socket.addEventListener("message", function (e) {
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
                renderGroups();
                renderTarget(config);
                break;
            case "player":
                player = message.data;
                var currentPlayerID = document.getElementById("current-player-id");
                currentPlayerID.innerHTML = player.ID;
                break;
        }
    });
    socket.addEventListener("open", function (e) {
        socket.send(JSON.stringify({
            type: "GetState"
        }));
    })

    setInterval(function () {
        if (state) {
            var d = Date.now() - new Date(state.Timer).getTime()
            timer.innerHTML = Math.round(d / 1000)
        }
    }, 100)

    function mouseMoveListener(tan, startTanPos, startMousePos) {
        return (e) => {
            tan.location.x = Math.round(clamp(startTanPos.x + (e.clientX - startMousePos.x), 0, config.Size.x));
            tan.location.y = Math.round(clamp(startTanPos.y + (e.clientY - startMousePos.y), 0, config.Size.y));
            var {path, text} = getTan(tan.id)
            renderTan(tan, path, text);
            socket.send(JSON.stringify({
                type: "MoveTan",
                tan: tan.id,
                location: tan.location,
                rotation: tan.rotation
            }));
        }
    };

    // Rotate tan clockwise or counter-clockwise
    function rotateListener (tan) {
        return (e) => {
            const key = e.which;
            let d = 0;
            switch (key) {
                case 88:
                d = 1;
                break;
                case 90:
                d = -1
                break;
            }

            if (d) {
                console.log(`[rotate] ${key}`);
                tan.rotation = rotate(tan.rotation, d);
                var {path, text} = getTan(tan.id);
                renderTan(tan, path, text);

                socket.send(JSON.stringify({
                    type: "MoveTan",
                    tan: tan.id,
                    location: tan.location,
                    rotation: tan.rotation
                }));
            }
        }
    }

    // Creates DOM nodes necessary to display a tan
    function initializeTan(tanID) {
        var path = document.createElementNS(view.namespaceURI, "path");
        path.addEventListener("click", onMouseDown)
        
        var txt = document.createElementNS(view.namespaceURI, "text");
        txt.setAttribute("font-family", "Verdana");
        txt.setAttribute("font-size", "12");

        var txtPath = document.createElementNS(view.namespaceURI, "textPath");
        txtPath.setAttribute("href", `#${tanID}`);

        txtPath.id = `txtPath-${tanID}`;
        txtPath.innerHTML = "";

        txt.appendChild(txtPath);
        gPath.appendChild(path);
        gText.appendChild(txt);

        return {path, text: txtPath}
    }

    const handlers = {};
    function onMouseDown(e) {
        if (e.ctrlKey) {
            var path = e.target;
            var id = parseInt(path.id.match(/tan-(\d+)/)[1]);

            var tan = state.tans.find(function (tan) {
                return tan.id == id
            });

            if (tan.player === player.ID) {
                var unlock = unlockTan(id);
                if (!unlock) {
                    console.log(`Error encountered while unlocking tan ${id}`);
                }
                console.log(`Releasing tan id=${id}`);

                if (handlers[id] && handlers[id].move && handlers[id].rotate) {
                    document.removeEventListener("pointermove", handlers[id].move);
                    document.removeEventListener("keydown", handlers[id].rotate);
                }

                socket.send(JSON.stringify({
                    type: "ObtainTan",
                    tan: id,
                    release: true
                }));
            } else {
                var locked = lockTan(tan.id);
                if (!locked) {
                    return;
                }

                console.log(`Holding on to tan id=${id}`);
                const startTanPos = {
                    x: tan.location.x,
                    y: tan.location.y,
                    r: tan.rotation
                };

                const startMousePos = {
                    x: e.clientX,
                    y: e.clientY
                };

                handlers[id] = {
                    move: mouseMoveListener(tan, startTanPos, startMousePos),
                    rotate: rotateListener(tan)
                };

                document.addEventListener("pointermove", handlers[id].move);
                document.addEventListener("keydown", handlers[id].rotate);

                socket.send(JSON.stringify({
                    type: "ObtainTan",
                    tan: id,
                    release: false
                }));
            }
        }
    }
});

function rotate(r, d) {
    return (r + d * 15 + 720) % 360;
}

function clamp(x, min, max) {
    return Math.max(min, Math.min(x, max))
}