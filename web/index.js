

let ws;

$(function(){
    ws = new WebSocket('ws://' + window.location.host + window.location.pathname + 'ws');

    ws.onclose = function(evt) {};

    ws.onmessage = function(evt) {
        console.log('RESPONSE: ' + evt.data);
        let message = JSON.parse(evt.data);

        switch(message.type){
            case "frame":
                $('#display').attr('src', 'data:image/png;base64,' + message.data);
                break;
            default:
                console.log("unhandled message type " + message.type);
        }
    }
});
