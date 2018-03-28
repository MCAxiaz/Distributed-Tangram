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
    node.setAttribute('stroke', model.shape.stroke);
    node.setAttribute('transform', transform);
    node.setAttribute('d', d);
    return node
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
            view.appendChild(tan);
        }
        return tan;
    }

    function render(state) {
        for (let tan of state.tans) {
            let node = getTan(tan);
            renderTan(tan, node);
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
        case "config":
            config = message.data;
            view.setAttribute("width", config.Size.x)
            view.setAttribute("height", config.Size.y)
        }
    });
    socket.addEventListener("open", function(e) {
        socket.send(JSON.stringify({
            type:"GetState"
        }));
    })
})
