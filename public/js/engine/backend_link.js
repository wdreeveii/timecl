/* 
	logic engine
	(C) 2013 Jason Hunt
	nulluser@gmail.com
*/

"use strict";

function backend_setproperties(index, id, PropertyCount, PropertyNames, PropertyTypes, PropertyValues)
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
function backend_hookobject(index, id, source)
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

function backend_unhookobject(index, id)
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

function backend_setoutput(index, id, output)
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

function backend_deleteobject(index, id)
{
	var cmd = "id=" + id ;
	
	$.ajax({
		url: "/engine/delete",
		context: document.body,
		type: "POST",
		data: cmd
	});
}

function backend_addobject(obj)
{
	var event = {
		Type: "add",
		Data: obj,
		Timestamp: 1000000,
	}
	socket.send(JSON.stringify(event));
}

function backend_moveobject(index, id, x_pos, y_pos)
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

function backend_setguides(index, id, Terminals)
{
	var eevent = {
		Type: "edit",
		Data: {
			Id: id,
			State: {
				"Terminals": Terminals
			}
		},
		Timestamp: 100000000
	}
	socket.send(JSON.stringify(eevent));
}

var socket = new WebSocket('ws://'+window.location.host+'/engine/ws');
socket.onmessage = function(event) {
	console.log(event.data);
	var event_msg = JSON.parse(event.data);
	if (event_msg["Type"] == "edit") {
		console.log("changestate");
		var event_data = event_msg["Data"];
		var id = event_data["Id"]
		var changes = event_data["State"]
		for (var change in changes) {
			obj[id][change] = changes[change];
		}
	} else if (event_msg["Type"] == "init") {
		var event_data = event_msg["Data"]
		obj = load_objects(event_data)
	}
}

