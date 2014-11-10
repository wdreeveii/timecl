/* 
	logic engine
	(C) 2013 Jason Hunt
	nulluser@gmail.com
*/

"use strict";

function backend_setproperties(id, PropertyCount, PropertyNames, PropertyTypes, PropertyValues) {
	var eevent = {
		Type: "edit_object",
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

function backend_hookobject(id, source) {
	var eevent = {
		Type: "edit_object",
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

function backend_unhookobject(id) {
	var eevent = {
		Type: "unhook",
		Data: id,
		Timestamp: 100000000
	}
	socket.send(JSON.stringify(eevent));
}

function backend_setoutput(id, output) {
	console.log("setting output");
	var eevent = {
		Type: "edit_object",
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

function backend_deleteobject(id) {
	console.log("delete object", id, typeof id);
	var eevent = {
		Type: "delete_object",
		Data: id,
		Timestamp: 100000000
	}
	socket.send(JSON.stringify(eevent));
}

function backend_addobject(type, x, y) {
	var event = {
		Type: "add_object",
		Data: {
			Type: type,
			X: x,
			Y: y
		},
		Timestamp: 1000000
	}
	socket.send(JSON.stringify(event));
}

function backend_moveobject(id, x_pos, y_pos) {
	var eevent = {
		Type: "edit_object",
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
var error_list = new Array();

var socket;

function backend_start() {
	socket = TameSocket({
		target: 'ws://' + window.location.host + '/engine/ws/1',
		msgMinInterval: 200,
		msgProcessor: function(bufferedEvents) {
			requestAnimationFrame(function() {
				engine.process_messages(bufferedEvents);
				engine.draw_display();
			});
		}
	});
}