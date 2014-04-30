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

function process_messages(bufferedMsgs) {
	while (bufferedMsgs.length > 0) {
		var event_msg = JSON.parse(bufferedMsgs.shift());
		console.log(event_msg);
		if (event_msg.Type == "edit_many") {
			var event_data = event_msg["Data"];
			for (var i = 0; i < event_data.length; i++) {
				var state_change = event_data[i];
				var id = state_change["Id"];
				var changes = state_change["State"];
				for (var change in changes) {
					obj[id][change] = changes[change];
				}
			}
		} else if (event_msg.Type == "errors") {
			var event_data = event_msg["Data"];
			for (i in event_data) {
				var errkey = event_data[i]["Error"];
				if (errkey in error_list) {
					error_list[errkey]["Time"] = event_data[i]["Time"];
					error_list[errkey]["Count"] += event_data[i]["Count"];
					var eventtext = "<table><tr><td>" + errkey +
						"</td><td>" + error_list[errkey]["Count"] +
						"</td></tr></table>";
					error_list[errkey]["Noty"].setText(eventtext);
				} else {
					var eventtext = "<table><tr><td>" + errkey +
						"</td><td>" + event_data[i]["Count"] +
						"</td></tr></table>"
					error_list[errkey] = {
						Noty: noty({
							text: eventtext
						}),
						Count: event_data[i]["Count"],
						Time: event_data[i]["Time"],
						First: event_data[i]["First"]
					};
				}
			}
		} else if (event_msg.Type == "edit") {
			var event_data = event_msg["Data"];
			var id = event_data["Id"]
			if (id in obj) {
				var changes = event_data["State"]
				for (var change in changes) {
					obj[id][change] = changes[change];
				}
			}
		} else if (event_msg.Type == "add") {
			var event_data = event_msg["Data"];
			var object = load_object(event_data);
			obj[object["Id"]] = object;
		} else if (event_msg.Type == "del") {
			delete obj[event_msg.Data];
		} else if (event_msg.Type == "init") {
			var event_data = event_msg["Data"];
			obj = load_objects(event_data);
			zoom_extent();
		} else if (event_msg["Type"] == "init_ports") {
			if (property_window == null) {
				port_list = event_msg["Data"];
			} else {
				property_window.set('port_list', event_msg["Data"]);
			}
		} else {
			console.log("Unknown type:", event_msg);
		}
	}
}

function backend_start() {
	socket = TameSocket({
		target: 'ws://' + window.location.host + '/engine/ws/1',
		msgProcessor: function(bufferedEvents) {
			requestAnimationFrame(function() {
				process_messages(bufferedEvents);
				draw_display();
			});
		}
	});
}