/* 
	logic engine
	(C) 2013 Jason Hunt
	nulluser@gmail.com
*/

"use strict";

function backend_setproperties(id, PropertyCount, PropertyNames, PropertyTypes, PropertyValues)
{
	var eevent = {
		Type: "edit",
		Data: {
			Id: id,
			State: {
				"PropertyCount": PropertyCount,
				"PropertyNames": PropertyNames,
				"PropertyTypes": PropertyTypes,
				"PropertyValues": PropertyValues
			}
		},
		Timestamp: 100000000
	}
	socket.send(JSON.stringify(eevent));
}

function backend_new_engine(name)
{
	$.ajax({
		url: "/engine/new",
		context: document.body,
		type: "POST",
	});
}

function backend_save_engine(name)
{
	var cmd = "name="+name;

	$.ajax({
		url: "/engine/save",
		context: document.body,
		type: "POST",
		data: cmd
	});
}
function backend_hookobject(id, source)
{
	var eevent = {
		Type: "edit",
		Data: {
			Id: id,
			State: {
				"Source": source
			}
		},
		Timestamp: 100000000
	}
	socket.send(JSON.stringify(eevent));
}

function backend_unhookobject(id)
{
	var eevent = {
		Type: "edit",
		Data: {
			Id: id,
			State: {
				"Source": -1
			}
		},
		Timestamp: 100000000
	}
	socket.send(JSON.stringify(eevent));
}

function backend_setoutput(id, output)
{
	console.log("setting output");
	var eevent = {
		Type: "edit",
		Data: {
			Id: id,
			State: {
				"Output": output
			}
		},
		Timestamp: 100000000
	}
	socket.send(JSON.stringify(eevent));
}

function backend_deleteobject(id)
{
	var eevent = {
		Type: "del",
		Data: {
			Id: id
		},
		Timestamp: 100000000
	}
	socket.send(JSON.stringify(eevent));
}

function backend_addobject(obj)
{
	console.log("adding", obj)
	var event = {
		Type: "add",
		Data: obj,
		Timestamp: 1000000,
	}
	socket.send(JSON.stringify(event));
}

function backend_moveobject(id, x_pos, y_pos)
{
	var eevent = {
		Type: "edit",
		Data: {
			Id: id,
			State: {
				"Xpos": x_pos,
				"Ypos": y_pos
			}
		},
		Timestamp: 100000000
	}
	socket.send(JSON.stringify(eevent));
}

var socket;
var socket_events = new Array();
function process_messages() {
	//console.log("events len", socket_events.length);
	while (socket_events.length > 0) {
		var event_msg = JSON.parse(socket_events.shift());
		if (event_msg["Type"] == "add") {
			var event_data = event_msg["Data"];
			var object = load_object(event_data);
			obj[object["Id"]] = object;
		} else if (event_msg["Type"] == "edit_many") {
			var event_data = event_msg["Data"];
			for (var i = 0; i < event_data.length; i++ ) {
				var state_change = event_data[i];
				var id = state_change["Id"];
				var changes = state_change["State"];
				for (var change in changes) {
					obj[id][change] = changes[change];
				}
			}
		} else if (event_msg["Type"] == "edit") {
			var event_data = event_msg["Data"];
			var id = event_data["Id"]
			if (id in obj) {
				var changes = event_data["State"]
				for (var change in changes) {
					obj[id][change] = changes[change];
				}
			}
		} else if (event_msg["Type"] == "init") {
			var event_data = event_msg["Data"];
			/*for (x in event_data) {
				console.log(x);
			}*/
			console.log(event_data);
			obj = load_objects(event_data);
			resize_canvas();
		} else if (event_msg["Type"] == "init_ports") {
			property_window.set('port_list', event_msg["Data"]);
		}
	}
}
var msg_rate_limit = null;
function handle_message(event) {
	//console.log(event.data);
	socket_events.push(event.data);
	if (msg_rate_limit == null) {
		msg_rate_limit = setTimeout(function() {
			requestAnimationFrame(function() {
				process_messages();
				draw_display();
			});
			msg_rate_limit = null;
		}, 1000);
	}
}

var backend_reset_timeout = false;
function setup_socket() {
	backend_reset_timeout = false;
	socket = new WebSocket('ws://'+window.location.host+'/engine/ws');
	socket.onmessage = handle_message;
	socket.onerror = function(event) {
		console.log("socket error", event);
	}
	socket.onclose = function(event) {
		console.log("on close");
		reset_socket();
	}
}

function reset_socket() {
	if (backend_reset_timeout == false) {
		socket.close();
		setTimeout(setup_socket, 5000);
		backend_reset_timeout = true;
	}
}
function backend_start() {
	var sleep_detect_interval_start = new Date();
	var sleep_detect_interval = setInterval(function() {
		if (sleep_detect_interval_start.getTime() + 60000 < new Date().getTime()) {
			console.log("Detected sleep: restarting...")
			reset_socket();
		}
		sleep_detect_interval_start = new Date();
	}, 1000);
	setup_socket();
	$(document).on('online', function (event) {
		console.log("online");
    	reset_socket();
	});
}

