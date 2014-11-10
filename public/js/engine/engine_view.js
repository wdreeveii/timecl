/* 
	logic engine
	(C) 2013 Jason Hunt
	nulluser@gmail.com
	With large modification from Whitham D. Reeve II
	thetawaves@gmail.com
*/
"use strict";

function format(n) {
	return (Number(n).toFixed(2));
} // Float to formatted string

var EngineView = function() {
	var obj = new Array();

	var ui_mode = "none";
	var ui_addtype = "";
	var sel_obj = -1;
	var obj_x_ofs = 0;
	var obj_y_ofs = 0;
	var has_moved = 0;
	var show_guide = 1;
	var mouse_state = "up";
	var mouse_x = 0;
	var mouse_y = 0;
	var x_ofs_start = 0;
	var y_ofs_start = 0;
	var container_x = 0;
	var container_y = 0;

	// Pan / Zoom
	var zoom = 1;
	var x_ofs = 0;
	var y_ofs = 0;

	var start = function() {
		var canvas = document.getElementById('canvas')
		// Map mouse functions
		$(canvas).mouseup(mouse_up);
		$(canvas).mousedown(mouse_down);
		$(canvas).mousemove(mouse_move);
		$(canvas).mouseout(mouse_out);

		resize_canvas();
	}

	var mouse_pos = function(ev) {
		var coords = $('#canvas').offset();
		if (ev.pageX || ev.pageY) {
			return {
				x: ev.pageX - coords.left,
				y: ev.pageY - coords.top
			};
		}
		return {
			x: ev.clientX + document.body.scrollLeft - document.body.clientLeft - canvas_x_ofs,
			y: ev.clientY + document.body.scrollTop - document.body.clientTop - canvas_y_ofs
		};
	}

	var mouse_out = function(ev) {
		/*	mouse_state = "up";
		if (ui_mode == "moving") {
			set_mode("none");
		}*/
	}

	var mouse_up = function(ev) {
		var pos = mouse_pos(ev);
		mouse_state = "up";

		if (ui_mode == "moving") {
			if (!has_moved) {
				select_object(obj[sel_obj]);
			} else {
				resize_canvas();
				backend_moveobject(obj[sel_obj].Id, obj[sel_obj].Xpos, obj[sel_obj].Ypos);

				for (var i in obj[sel_obj].Terminals) {
					var k = obj[sel_obj].Terminals[i];
					backend_moveobject(obj[k].Id, obj[k].Xpos, obj[k].Ypos);
				}
			}
			has_moved = 0;
			set_mode("none");
		}
		requestAnimationFrame(draw_display);
	}

	var mouse_down = function(ev) {
		var pos = mouse_pos(ev);
		mouse_state = "down";
		container_x = $('#canvas_container').scrollLeft();
		container_y = $('#canvas_container').scrollTop();
		x_ofs_start = x_ofs;
		y_ofs_start = y_ofs;
		mouse_x = pos.x;
		mouse_y = pos.y;

		if (pos.x < 0 || pos.y < 0) return;
		if (ui_mode == "none") // No mode, either find an obj or clear mode
		{
			var i = find_object(pos.x, pos.y);
			if (i == -1) // No object found, go clear selection
			{
				select_none();
			} else {
				ui_move_object(pos, i);
			}

		} else
		if (ui_mode == "add_object") ui_add_object(pos);
		else // Add generic object
		if (ui_mode == "add_pipe") ui_add_pipe1(pos);
		else // Select first object for adding wire
		if (ui_mode == "add_pipe2") ui_add_pipe2(pos);
		else // Select second object for adding wire
		if (ui_mode == "delete") ui_delete_object(pos);
		else // Delete
		if (ui_mode == "unhook") ui_unhook_object(pos);
		else // Unhook
		if (ui_mode == "moving") {} else
			set_mode("none");

		requestAnimationFrame(draw_display);
	}

	var mouse_move = function(ev) {
		var pos = mouse_pos(ev);
		if (ui_mode == "add_pipe2") {

		} else
		if (ui_mode == "moving") {
			if (mouse_state == "down") {
				has_moved = 1;
			}
			var new_x = snap_val(get_world_x(pos.x) - obj_x_ofs);
			var new_y = snap_val(get_world_y(pos.y) - obj_y_ofs);

			var delta_x = new_x - obj[sel_obj].Xpos;
			var delta_y = new_y - obj[sel_obj].Ypos;

			obj[sel_obj].Xpos = new_x;
			obj[sel_obj].Ypos = new_y;


			for (var j in obj[sel_obj].Terminals) {
				var k = obj[sel_obj].Terminals[j];

				obj[k].Xpos += delta_x;
				obj[k].Ypos += delta_y;
			}
			//resize_canvas();
			requestAnimationFrame(draw_display);
		}
		/* else
		// Pan grid if dragging mouse
		if (mouse_state == "down")
		{
			var dx = pos.x - mouse_x;		
			var dy = pos.y - mouse_y;
			//x_ofs = x_ofs_start + dx;
			//y_ofs = y_ofs_start + dy;

			//$('#canvas_container').scrollLeft(container_x - dx);
			//$('#canvas_container').scrollTop(container_y - dy);

		}*/
	}
	var find_extent = function() {
		var max_x = 0;
		var min_x = 0;
		var max_y = 0;
		var min_y = 0;
		for (var prop in obj) {
			if (obj.hasOwnProperty(prop)) {
				if (obj[prop].Xpos > max_x) {
					max_x = obj[prop].Xpos;
				} else if (obj[prop].Xpos < min_x) {
					min_x = obj[prop].Xpos;
				}
				if (obj[prop].Ypos > max_y) {
					max_y = obj[prop].Ypos;
				} else if (obj[prop].Ypos < min_y) {
					min_y = obj[prop].Ypos;
				}
			}
		}
		return [max_x + 100, min_x - 100, max_y + 100, min_y - 100];
	}

	var zoom_extent = function() {
		var extents = find_extent();
		var container = $(document.getElementById("canvas_container"));
		var x = container.innerWidth();
		var y = container.innerHeight();
		var content_x_size = extents[0] - extents[1] + 100;
		var content_y_size = extents[2] - extents[3] + 100;
		var zoom_x = x / content_x_size;
		var zoom_y = y / content_y_size;
		var new_zoom = zoom_x;
		if (zoom_y < new_zoom) {
			new_zoom = zoom_y;
		}
		zoom = new_zoom;
		x_ofs = zoom * (-extents[1]);
		y_ofs = zoom * (-extents[3]);
		resize_canvas();
	}

	var timer = null;
	var resize_canvas = function() {
		if (timer != null) {
			clearTimeout(timer);
			timer = setTimeout(_resize_canvas, 500);
		} else {
			_resize_canvas();
		}

	}

	var _resize_canvas = function() {
		var container = $(document.getElementById("canvas_container"));
		var canvas = document.getElementById("canvas");
		var x = container.innerWidth();
		var y = container.innerHeight();

		var save_x_ofs = x_ofs;
		var save_y_ofs = y_ofs;

		var extents = find_extent();
		x_ofs = zoom * (-extents[1]);
		y_ofs = zoom * (-extents[3]);

		var x_size = zoom * ((extents[0] - extents[1]) + 100);
		var y_size = zoom * ((extents[2] - extents[3]) + 100);

		if (x_size > x) {
			x = x_size;
		}
		if (y_size > y) {
			y = y_size;
		}
		if (container.innerHeight() < y) {
			x -= 20;
		}
		if (container.innerWidth() < x) {
			y -= 20;
		}

		canvas.width = x;
		canvas.height = y;
		requestAnimationFrame(draw_display);
	}

	var zoom_in = function() {
		zoom += 0.1;
		if (zoom > max_zoom) {
			zoom = max_zoom;
		}
		resize_canvas();
	}

	var zoom_out = function() {
		zoom -= 0.1;
		if (zoom < min_zoom) {
			zoom = min_zoom;
		}
		resize_canvas();
	}

	var ui_add_pipe1 = function(pos) {
		//add_object(pos.x, pos.y, "pipe");

		var i = find_object(pos.x, pos.y)
		if (i != -1 && obj[i].Type == "guide") {
			obj[i].Selected = 1;

			set_mode("add_pipe2");
			sel_obj = i;

			requestAnimationFrame(draw_display);
		} else {
			set_mode("none");
		}
	}

	var ui_add_pipe2 = function(pos) {
		var i = find_object(pos.x, pos.y)
		if (i != -1 && i != sel_obj) {
			if (obj[sel_obj].Type == "guide" && obj[i].Type == "guide") {
				object_connect(sel_obj, i);
				requestAnimationFrame(draw_display);
			}

			obj[i].Selected = 0;
			obj[sel_obj].Selected = 0;

			set_mode("add_pipe");
		} else {
			set_mode("none");
		}
	}

	var ui_move_object = function(pos, i) {
		if (obj[i].Attached < 1) {
			obj_x_ofs = get_world_x(pos.x) - obj[i].Xpos;
			obj_y_ofs = get_world_y(pos.y) - obj[i].Ypos;

			//obj[i].Selected = 1;//!obj[i].Selected;

			sel_obj = i;
			set_mode("moving");

			requestAnimationFrame(draw_display);

		} else {

		}
	}

	var ui_delete_object = function(pos) {
		var i = find_object(pos.x, pos.y);
		if (i >= 0 && obj[i].Attached < 1) {
			backend_deleteobject(i);
			resize_canvas();
		} else {
			set_mode("none");
		}
	}

	var ui_unhook_object = function(pos) {
		var i = find_object(pos.x, pos.y);

		if (i != -1) {
			backend_unhookobject(i);
		} else {
			set_mode("none");
		}
	}

	var ui_add_object = function(pos) {
		backend_addobject(ui_addtype, pos.x, pos.y);
		set_mode("none");
		ui_addtype = "";
	}

	var set_guide = function(s) {
		if (s == "show") show_guide = 1;
		if (s == "hide") show_guide = 0;

		requestAnimationFrame(draw_display);
	}

	var select_none = function() {
		// more like blank_properties
		for (var i in obj)
			obj[i].Selected = 0;

		sel_obj = -1;
		property_window.set('current_obj', undefined);
	}

	var select_object = function(o) {
		if (o.Selected == 1) {
			object_toggle(o);
			return;
		}

		// Clear all selections
		for (var i in obj)
			obj[i].Selected = 0;


		o.Selected = 1;
		property_window.set('current_obj', o);

		requestAnimationFrame(draw_display);
	}

	var save_properties = function(sel_obs) {
		if (sel_obj == -1) return;

		var o = obj[sel_obj];

		// Save all properties
		for (var i = 0; i < o.PropertyCount; i++) {
			o.PropertyValues[i] = document.getElementById(o.PropertyNames[i] + "_field").value;
		}

		// Need to save peoperties to database
		o.save_properties();

		requestAnimationFrame(draw_display);

		return false;
	}

	var set_mode = function(m) {
		$('.tool').removeClass('active');
		if (m == 'none' || m == 'moving')
			$('#ptr_btn').addClass('active');
		else if (m == 'delete')
			$('#delete_btn').addClass('active');
		else if (m == 'unhook')
			$('#unhook_btn').addClass('active');
		else if (m == 'add_pipe' || m == 'add_pipe2')
			$('#wire_btn').addClass('active');

		if (m == 'delete' ||
			m == 'unhook' ||
			m == 'add_pipe' ||
			m == 'add_pipe2')
			$('#canvas').css('cursor', 'crosshair');
		else {
			$('#canvas').css('cursor', 'auto');
		}

		ui_mode = m;
	}

	var set_addmode = function(obj_type) {
		backend_addobject(obj_type, get_world_x(100), get_world_y(100));
	}

	var reset = function() {
		zoom = 1;
		resize_canvas();
	}
	var handle_size = 10;
	var guide_size = 10;

	function get_x(x) {
		return x * zoom + x_ofs;
	} // Get screen x from world x
	function get_y(y) {
		return y * zoom + y_ofs;
	} // Get screen y from world y
	function get_world_x(x) {
		return (x - x_ofs) / zoom;
	} // Get world x from screen x
	function get_world_y(y) {
		return (y - y_ofs) / zoom;
	} // Get world y from screen y


	var draw_wire = function(ctx, objects, i1, i2, p1, p2) {
		var o1 = objects[i1];
		var o2 = objects[i2];

		var x1 = o1.Xpos + o1.Xsize / 2;
		var y1 = o1.Ypos + o1.Ysize / 2;

		var x2 = o2.Xpos + o1.Xsize / 2;
		var y2 = o2.Ypos + o1.Ysize / 2;

		if (show_guide == 0) {
			if (o1.Dir == dir_type.left) x1 += handle_size / 2;
			if (o1.Dir == dir_type.right) x1 -= handle_size / 2;
			if (o1.Dir == dir_type.up) y1 += handle_size / 2;
			if (o1.Dir == dir_type.down) y1 -= handle_size / 2;

			if (o2.Dir == dir_type.left) x2 += handle_size / 2;
			if (o2.Dir == dir_type.right) x2 -= handle_size / 2;
			if (o2.Dir == dir_type.up) y2 += handle_size / 2;
			if (o2.Dir == dir_type.down) y2 -= handle_size / 2;
		}

		ctx.moveTo(Math.round(get_x(x1)), Math.round(get_y(y1)));
		ctx.lineTo(Math.round(get_x(x2)), Math.round(get_y(y2)));
	}

	var draw_object = function(ctx, o, x_size, y_size) {
		if (o.Type == "guide" && show_guide == 0) return;

		if (o.Selected) {
			var old_fill = ctx.fillStyle;

			var border = 4;

			ctx.fillStyle = "rgb(255,0,0)";
			ctx.fillRect(Math.round(get_x(o.Xpos - border)), Math.round(get_y(o.Ypos - border)), Math.round((o.Xsize + 2 * border) * zoom), Math.round((o.Ysize + 2 * border) * zoom));

			ctx.fillStyle = old_fill;
		}

		o.draw_icon(ctx);

		o.draw_properties(ctx, o.Xpos, o.Ypos);
	}

	var draw_objects = function(ctx, objects, x_size, y_size) {
		ctx.clearRect(0, 0, x_size, y_size);

		ctx.lineWidth = 1;
		ctx.strokeStyle = "rgb(0,0,0)";
		ctx.fillStyle = "rgb(255,255,255)";

		// Draw all wires
		ctx.beginPath();
		for (var i in objects) {
			var idx = objects[i].Source;
			if (idx >= 0)
				draw_wire(ctx, objects, i, idx, 1, 0);
		}
		ctx.stroke();
		ctx.closePath();


		ctx.lineWidth = 2;
		ctx.strokeStyle = "rgb(0,0,0)";
		ctx.fillStyle = "rgb(255,255,255)";


		for (var i in objects)
			draw_object(ctx, objects[i], x_size, y_size);
	}

	var draw_display = function() {
		var canvas = document.getElementById("canvas");
		var ctx = canvas.getContext("2d");

		draw_objects(ctx, obj, canvas.width, canvas.height);
	}

	var find_object = function(x, y) {
		// convert to world cords
		x = get_world_x(x);
		y = get_world_y(y);
		var keys = [];
		var ii = 0;
		for (var k in obj) {
			keys[ii++] = parseInt(k);
		}
		keys.reverse();

		for (var k in keys) {
			var i = keys[k];
			if (x >= obj[i].Xpos &&
				y >= obj[i].Ypos &&
				x <= obj[i].Xpos + obj[i].Xsize &&
				y <= obj[i].Ypos + obj[i].Ysize) {

				return (i);
			}
		}

		return (-1);
	}

	var process_messages = function(bufferedMsgs) {
		while (bufferedMsgs.length > 0) {
			var event_msg = JSON.parse(bufferedMsgs.shift());
			//console.log(event_msg);
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
	var get_zoom = function() {
		return zoom;
	}
	var bounding_rect = function(ctx, o) {
		ctx.lineWidth = 0.25;
		ctx.beginPath();
		ctx.moveTo(Math.round(get_x(o.Xpos)), Math.round(get_y(o.Ypos)));
		ctx.lineTo(Math.round(get_x(o.Xpos + o.Xsize)), Math.round(get_y(o.Ypos)));
		ctx.lineTo(Math.round(get_x(o.Xpos + o.Xsize)), Math.round(get_y(o.Ypos + o.Ysize)));
		ctx.lineTo(Math.round(get_x(o.Xpos)), Math.round(get_y(o.Ypos + o.Ysize)));

		ctx.closePath();

		ctx.fill();
		ctx.stroke();
	}

	var construct = function(new_object) {
		if (typeof(window[new_object.Type + "_type"]) == 'function') {
			window[new_object.Type + "_type"](new_object);
		}
	}

	var load_object = function(newobject) {
		var o = new object_type;
		$.extend(o, newobject);
		construct(o);
		return o;
	}

	var load_objects = function(objects) {
		var results = Array();
		for (var i = 0; i < objects.length; i++) {
			var o = new object_type;
			$.extend(o, objects[i]);
			construct(o);
			results[o["Id"]] = o;
		}
		return results;
	}

	var object_connect = function(o1, o2) {
		obj[o2].Source = o1;
		backend_hookobject(obj[o2].Id, obj[o1].Id)
	}

	var object_toggle = function(o) {
		o.set_output(!o.Output);
	}

	return {
		resize: resize_canvas,
		start: start,
		zoom_extent: zoom_extent,
		draw_display: draw_display,
		process_messages: process_messages,
		get_x: get_x,
		get_y: get_y,
		get_zoom: get_zoom,
		bounding_rect: bounding_rect,
		add_obj: set_addmode,
		save_properties: save_properties,
		set_mode: set_mode,
		zoom_in: zoom_in,
		zoom_out: zoom_out,
	};
}