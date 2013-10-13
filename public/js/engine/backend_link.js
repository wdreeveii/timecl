/* 
	logic engine
	(C) 2013 Jason Hunt
	nulluser@gmail.com
*/

"use strict";

function backend_update( )
{
	backend_getstates( );
	/*for (var i in obj)
	{
		if (obj[i].type == "httpsource")
			data_source_request(i, obj[i].source_name);
		
	}*/
}

function backend_hookobject(index, id, source)
{
	var cmd = "id=" + id + "&" +
		      "source=" + source;
	
	$.ajax({
		url: "/engine/hook",
		context: document.body,
		type: "POST",
		data: cmd
	});
}

function backend_setoutput(index, id, output)
{
	var cmd = "id=" + id + "&" +
		      "output=" + output;

	$.ajax({
		url: "/engine/set/output",
		context: document.body,
		type: "POST",
		data: cmd
	});
}

function backend_setproperties(index, id, property_count, property_names, property_types, property_values)
{
	var cmd = "id=" + id + "&";
		      
	for (var ii = 0; ii < property_count; ii++) {
		cmd += "property_names["+ii+"]=" + property_names[ii] + "&";
	}
	for (var ii = 0; ii < property_count; ii++) {
		cmd += "property_types["+ii+"]=" + property_types[ii] + "&";
	}
	for (var ii = 0; ii < property_count; ii++) {
		cmd += "property_values["+ii+"]=" + property_values[ii] + "&";
	}
	cmd += "property_count=" + property_count;

	$.ajax({
		url: "/engine/set/properties",
		context: document.body,
		type: "POST",
		data: cmd
	});
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

function backend_unhookobject(index, id)
{
	var cmd = "id=" + id ;
	
	$.ajax({
		url: "/engine/unhook",
		context: document.body,
		type: "POST",
		data: cmd
	});
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
	console.log("sending1");
	socket.send(JSON.stringify(event));
	console.log("sending2");
}

function backend_moveobject(index, id, x_pos, y_pos)
{
	var cmd = "id=" + id + "&" + 
    	      "x_pos=" + x_pos + "&" +
    	      "y_pos=" + y_pos;
    
	$.ajax({
		url: "/engine/move",
		context: document.body,
		type: "POST",
		data: cmd
	});
}

function backend_setguides(index, id, Terminals)
{
	var cmd = "id=" + id  + "&" +
		      "Terminals=" + Terminals;
	
	$.ajax({
		url: "/engine/set/guides",
		context: document.body,
		type: "POST",
		data: cmd
	}).done(function(response) {
		
	});
}

function backend_load( )
{
	reset();
	
	$.ajax({
		url: "/engine/list",
		context: document.body,
	}).done(function(response) {
		console.log(response);
		var tmp = eval(response);
		console.log(tmp);
		console.log(tmp.length);
		if (tmp.length == 0) return;
	
		for (var j = 0; j < tmp.length; j++)
		{
			var id = parseInt(tmp[j].Id);
			var type = tmp[j].Type;
			var x_pos = parseInt(tmp[j].Xpos);
			var y_pos = parseInt(tmp[j].Ypos);
			var attached = parseInt(tmp[j].Attached);
			var dir = parseInt(tmp[j].Dir);
			var source_id = parseInt(tmp[j].Source);
			
			var property_count = parseInt(tmp[j].PropertyCount);
			var property_names = tmp[j].PropertyNames;
			var property_types = tmp[j].PropertyTypes;
			var property_values = tmp[j].PropertyValues;

			var output = tmp[j].Output;
				
			id = parseInt(id);
		
			x_pos = parseInt(x_pos);
			y_pos = parseInt(y_pos);
		
			attached = parseInt(attached);
			dir = parseInt(dir);

			
			var index = load_object(obj, x_pos, y_pos, type, attached, dir);
	
			obj[index].id = id;
			console.log("setting " + String(index));
			obj[index].tmp_Terminals = tmp[j].Terminals;
			obj[index].source_id = source_id;
			
			obj[index].property_count = property_count;
			obj[index].property_names = property_names
			obj[index].property_types = property_types
			obj[index].property_values = property_values


			obj[index].output = parseFloat(output);
		}
	
	
		// add objects added, need to decode guidelist and sources
		// very bad, O^2
		
		for (var j in obj)
		{
			var guide_list = obj[j].tmp_Terminals;
			if (guide_list) {
				for (var k = 0; k < guide_list.length; k++)
				{
					for (var l in obj)
					{
						if (obj[l].id == guide_list[k])
						{
							obj[j].Terminals.push(l);
						}
					}
				}
			}
			for (var l in obj)
			{
					if (obj[l].id == obj[j].source_id)
					{
						obj[j].source = parseInt(l);
					}			
			}
		}
	});
}

function backend_getstates( )
{
	//reset();
	$.ajax({
		url: "/engine/states",
		context: document.body,
		type: "GET",
		data: {state: 1},
		cache: false,
	}).done(function(response) {
		if (response == null) return;
		
		var tmp = eval(response);
	
		if (tmp == null || tmp.length == 0) return;
	
		for (var i = 0; i < tmp.length; i++)
		{
			var id = parseInt(tmp[i].Id);
			var output = parseFloat(tmp[i].Output);
			
			var index = -1;

			for (var j in obj)
			{
				if (obj[j].id == id)
					index  = j;
			}

			if (index >= 0)
			{
				obj[index].output = parseFloat(output);
			}
		}
	});
}
var socket = new WebSocket('ws://'+window.location.host+'/engine/ws');
socket.onmessage = function(event) {
	console.log(event.data)
}
function backend_start( )
{
	setInterval(function() { backend_update(); }.bind(this), 10000);
	var testevent = {
		Type: "edit",
		Data: {
			Id: 1,
			State: {
				"test": 123
			}
		},
		Timestamp: 100000000
	}
	setInterval(function() { socket.send(JSON.stringify(testevent));}.bind(this), 2000);
	backend_load();
}


