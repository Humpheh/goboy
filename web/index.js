

let ws;

$(function(){
    ws = new WebSocket('ws://' + window.location.host + window.location.pathname + 'ws');

    ws.onclose = function(evt) {};

    ws.onmessage = function(evt) {
        let message = JSON.parse(evt.data);

        switch(message.type){
            case "frame":
                $('#display').attr('src', 'data:image/png;base64,' + message.data);
                break;
            default:
                console.log("unhandled message type " + message.type);
        }
    }

    window.addEventListener("keyup", send_input);
    window.addEventListener("keydown", send_input);

    function send_input(event){
        if(event.repeat){
            return
        }
        let message = {
            "type": "input",
            "key": event.key,
            "pressed": event.type === "keydown"
        }
        ws.send(JSON.stringify(message));
    }
});
