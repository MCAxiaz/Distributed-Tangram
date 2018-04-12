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

function renderTan(model, node) {
    var id = `tan-${model.id}`
    var transform = `translate(${model.location.x}, ${model.location.y}) rotate(${model.rotation})`;
    var d = "";
    model.shape.points.forEach(function (point, i) {
        var command = i == 0 ? "M" : "L";
        d += `${command} ${point.x} ${point.y} `;
    });
    d += "Z";

    node.id = id;
    node.setAttribute('fill', model.shape.fill);
    // Check if tan is being held by a player, and if it is, show that player's ID on it
    if (!state.tans[model.id - 1]) {
        console.log("No such tan.");
    }
    var txtPath = document.getElementById(`txtPath-tan-${model.id}`);
    if (state.tans[model.id - 1].player !== NO_PLAYER) {
        // Render player ID to tan
        if (!txtPath) {
            console.log(`textPath for tan ${model.id} does not exist.`);
        } else {
            node.setAttribute("fill-opacity", "0.5");
            txtPath.innerHTML = state.tans[model.id - 1].player;
        }
    } else {
        node.setAttribute("fill-opacity", "1");
        if (!txtPath) {
            console.log(`textPath for tan ${model.id} does not exist.`);
        } else {
            txtPath.innerHTML = "";
        }
    }
    node.setAttribute('stroke', model.shape.stroke);
    if (model.Matched) {
        node.setAttribute('stroke', 'green');
    }

    node.setAttribute('transform', transform);
    node.setAttribute('d', d);
    node.classList.add("tan");
    return node
}

// Attaches textPath to SVG for player's name
function attachPlayerNameTextToSVG(tanID) {
    var svg = document.getElementById("g-text");
    var use = document.createElementNS(view.namespaceURI, "use");

    use.setAttribute("href", `#${tanID}`);
    var txt = document.createElementNS(view.namespaceURI, "text");
    txt.setAttribute("font-family", "Verdana");
    txt.setAttribute("font-size", "12");

    var txtPath = document.createElementNS(view.namespaceURI, "textPath");
    txtPath.setAttribute("href", `#${tanID}`);

    txtPath.id = `txtPath-${tanID}`;
    txtPath.innerHTML = "";

    txt.appendChild(txtPath);
    svg.appendChild(use);
    svg.appendChild(txt);
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
    var timer = document.getElementById("timer");
    var dump = document.getElementById("dump");

    function getTan(model) {
        var tan = view.getElementById(`tan-${model.id}`);

        if (!tan) {
            tan = document.createElementNS(view.namespaceURI, "path");
            renderTan(model, tan);
            attachPlayerNameTextToSVG(tan.id);
            var gPath = document.getElementById("g-paths");
            gPath.appendChild(tan);
            tan.addEventListener("click", onMouseDown)
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

        // Check if tan is already held by someone else
        var tanObj = state.tans[tanID - 1];
        if (!tanObj) {
            console.log("Tan does not exist.");
            return false;
        }

        if (tanObj.player !== NO_PLAYER && tanObj.player !== player.ID) {
            console.log("Another player is already holding onto the tan.");
            return false;
        } else {
            tanObj.player = player.ID;
        }

        // Change path's fill opacity to half
        tan.setAttribute("fill-opacity", "0.5");

        // Display player name on tan
        var txtPath = document.getElementById(`txtPath-tan-${tanID}`);
        if (!txtPath) {
            console.log(`No such txtPath with tan ${tanID}.`);
            return false;
        }

        txtPath.innerHTML = player.ID;
        tan.classList.add("locked");
        tan.classList.add("draggable");

        console.log(`[Lock tan] Tan ${tanID}: I am possessed by ${player.ID}.`);

        return true;
    }

    function unlockTan(tanID) {
        var tan = view.getElementById(`tan-${tanID}`);

        if (!tan) {
            console.log("Cannot unlock a tan that does not exist.");
            return false;
        }

        // Set tan to unlocked
        var tanObj = state.tans[tanID - 1];
        if (!tanObj) {
            console.log("There exists no such tan.");
            return false;
        }

        tanObj.player = NO_PLAYER;

        // Restore path's fill opacity to 1
        tan.setAttribute("fill-opacity", "1");

        // Remove player name from tan
        var txtPath = document.getElementById(`txtPath-tan-${tanID}`);
        if (!txtPath) {
            console.log(`No such txtPath with tan ${tanID}`);
            return false;
        }

        txtPath.innerHTML = "";
        tan.classList.remove("locked");
        tan.classList.remove("draggable");

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
        var gTarget = document.createElementNS(view.namespaceURI, "g");
        gTarget.id = "g-target";
        view.appendChild(gTarget);
        var gPaths = document.createElementNS(view.namespaceURI, "g");
        gPaths.id = "g-paths";
        view.appendChild(gPaths);
        var gText = document.createElementNS(view.namespaceURI, "g");
        gText.id = "g-text";
        view.appendChild(gText);
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

    function mouseMoveListener(tan, path, startTanPos, startMousePos) {
        return (e) => {
            tan.location.x = Math.round(Math.max(0, Math.min(startTanPos.x + (e.clientX - startMousePos.x), config.Size.x)));
            tan.location.y = Math.round(Math.max(0, Math.min(startTanPos.y + (e.clientY - startMousePos.y), config.Size.y)));
            renderTan(tan, path);
            socket.send(JSON.stringify({
                type: "MoveTan",
                tan: tan.id,
                location: tan.location,
                rotation: tan.rotation
            }));
        }
    };

    // Rotate tan clockwise or counter-clockwise
    function rotateListener (tan, path) {
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
                renderTan(tan, path);
            }

            socket.send(JSON.stringify({
                type: "MoveTan",
                tan: tan.id,
                location: tan.location,
                rotation: tan.rotation
            }));
        }
    }

    const handlers = {};

    function onMouseDown(e) {
        if (e.ctrlKey) {
            var path = e.target;
            var id = parseInt(path.id.match(/tan-(\d+)/)[1]);

            var tan = state.tans.find(function (tan) {
                return tan.id == id
            });

            if (state.tans[id - 1].player === player.ID) {
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
                    move: mouseMoveListener(tan, path, startTanPos, startMousePos),
                    rotate: rotateListener(tan, path)
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
