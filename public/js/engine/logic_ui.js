/* 
	logic engine
	(C) 2013 Jason Hunt
	nulluser@gmail.com
*/

"use strict";

// Parameters
var snap = 5;
var min_zoom = 0.20;
var max_zoom = 2.0;
var update_rate = 3;

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
var	x_ofs_start = 0;
var	y_ofs_start = 0;
var container_x = 0;
var container_y = 0;

// Pan / Zoom
var zoom = 1;
var x_ofs = 0;
var y_ofs = 0;

var port_list = new Array();

var property_window;
$(function() {
	$('.tool').tooltip();
	$('#property_and_canvas').layout({applyDefaultStyles: false,
				center__onresize: resize_canvas,
	});
	$('#propertyTemplate').load('/public/property_window.html', function() {
		property_window = new Ractive({
			el: $('#property_sidebar'),
			template: $('#propertyTemplate').html(),
			data: {
				current_obj: false,
				tzdb: tzdb,
				current_timezone: jstz.determine().name(),
				port_list: new Array(),
				netmgrTypes: function(t) {
					var types = new Array("binput", "ainput", "boutput", "aoutput");
					t = parseInt(t);
					return types[t];
				}
			}
		});
		backend_start();
		start();
	});
});

/* 
	High level UI 
*/
function start()
{
	var canvas = document.getElementById('canvas')
	// Map mouse functions
	$(canvas).mouseup(mouse_up); 
	$(canvas).mousedown(mouse_down); 
	$(canvas).mousemove(mouse_move); 
	$(canvas).mouseout(mouse_out);
	// Mouse wheel, FF
	/*if (canvas.addEventListener)
        canvas.addEventListener('DOMMouseScroll', mouse_wheel, false);

	// Mouse Wheel IE
	window.onmousewheel = document.onmousewheel = mouse_wheel;
	*/
	resize_canvas();
	//data_source_start();
}

/* 
	Utility 
*/
function snap_val(x)
{
	return (Math.round((x)/snap)*snap);
}

function get_value(name)
{
	//return Number(document.getElementById(name).value);
	return document.getElementById(name).value;
}

function set_value(name, value)
{
	document.getElementById(name).value = value;
}
/* 
	End Utility 
*/

/* 
	UI
*/
function find_extent() {
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
	return [max_x, min_x, max_y, min_y];
}

function resize_canvas()
{
	var container = $(document.getElementById("canvas_container"));
	var canvas = document.getElementById("canvas");
	var x = container.innerWidth();
	var y = container.innerHeight();
	var save_x_ofs = x_ofs;
	var save_y_ofs = y_ofs;

	var extents = find_extent();
	x_ofs = -(extents[1] - 100);
	y_ofs = -(extents[3] - 100);

	var x_size = zoom * ((extents[0] - extents[1]) + 200);
	var y_size = zoom * ((extents[2] - extents[3]) + 200);

	if (x_size > x) {
		if (save_x_ofs < x_ofs) {
			container.animate({scrollLeft:0}, 10);
		} else if (save_x_ofs == x_ofs) {
			container.animate({scrollLeft:x_size - x}, 5);
		}
		x = x_size;
	}
	if (y_size > y) {
		if (save_y_ofs < y_ofs) {
			container.animate({scrollTop:0}, 10);
		} else if (save_y_ofs == y_ofs) {
			container.animate({scrollTop:y_size - y}, 5);
		}
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

function draw_display() {
 	var canvas = document.getElementById("canvas");
 	var ctx = canvas.getContext("2d");
	
	draw_objects(ctx, obj, canvas.width, canvas.height);
}

/* 
	End of high level UI 
*/

/*
	Keyboard / Mouse 
*/

function mouse_pos(ev)
{
	var coords = $('#canvas').offset();
	if(ev.pageX || ev.pageY)
	{ 
		return {x:ev.pageX - coords.left, y:ev.pageY - coords.top}; 
	} 
	return 	{ x:ev.clientX + document.body.scrollLeft - document.body.clientLeft  - canvas_x_ofs, 
			  y:ev.clientY + document.body.scrollTop  - document.body.clientTop   - canvas_y_ofs}; 
}
function mouse_out(ev) {
/*	mouse_state = "up";
	if (ui_mode == "moving") {
		set_mode("none");
	}*/
}
function mouse_up(ev)
{ 
	var pos = mouse_pos(ev);	
	mouse_state = "up";
	if (ui_mode == "moving")
	{
		if (!has_moved)
		{
			select_object(obj[sel_obj]);
		}
		else
		{
			resize_canvas();
			backend_moveobject(obj[sel_obj].Id, obj[sel_obj].Xpos, obj[sel_obj].Ypos);
		
			for (var i in obj[sel_obj].Terminals)
			{
				var k = obj[sel_obj].Terminals[i];
				backend_moveobject(obj[k].Id, obj[k].Xpos, obj[k].Ypos);
			}
		}
		has_moved = 0;
		set_mode("none");
	}
	requestAnimationFrame(draw_display);
} 

function mouse_down(ev)
{ 
 	var pos = mouse_pos(ev); 
	mouse_state = "down";
	container_x = $('#canvas_container').scrollLeft();
	container_y = $('#canvas_container').scrollTop();
	x_ofs_start = x_ofs;	
	y_ofs_start = y_ofs;
	mouse_x = pos.x;
	mouse_y = pos.y;

	if (pos.x < 0 || pos.y < 0) return;

	if (ui_mode == "none")// No mode, either find an obj or clear mode
	{
		var i = find_object(pos.x, pos.y);
		if (i == -1) // No object found, go clear selection
		{
			select_none();
		} else
		{
			ui_move_object(pos, i);
		}

	} else		
	if (ui_mode == "add_object") ui_add_object(pos); else		// Add generic object
	if (ui_mode == "add_pipe")   ui_add_pipe1(pos); else 		// Select first object for adding wire
	if (ui_mode == "add_pipe2")  ui_add_pipe2(pos); else		// Select second object for adding wire
	if (ui_mode == "delete")     ui_delete_object(pos); else	// Delete
	if (ui_mode == "unhook")     ui_unhook_object(pos); else	// Unhook
	if (ui_mode == "moving") {} else
		set_mode("none");

	requestAnimationFrame(draw_display);
} 

function mouse_move(ev)
{
 	var pos = mouse_pos(ev); 
 	if (ui_mode == "add_pipe2") {

 	} else
	if (ui_mode == "moving")
	{
		var new_x = snap_val(get_world_x(pos.x) - obj_x_ofs);
		var new_y = snap_val(get_world_y(pos.y) - obj_y_ofs);

		var delta_x = new_x - obj[sel_obj].Xpos ;
		var delta_y = new_y - obj[sel_obj].Ypos;

		obj[sel_obj].Xpos = new_x;
		obj[sel_obj].Ypos = new_y;


		for(var j in obj[sel_obj].Terminals)
		{
			var k = obj[sel_obj].Terminals[j];

			obj[k].Xpos += delta_x;
			obj[k].Ypos += delta_y;
		}

		requestAnimationFrame(draw_display);
	} else
	// Pan grid if dragging mouse
	if (mouse_state == "down")
	{
		var dx = pos.x - mouse_x;		
		var dy = pos.y - mouse_y;
		
		//x_ofs = x_ofs_start + dx;
		//y_ofs = y_ofs_start + dy;

		//$('#canvas_container').scrollLeft(container_x - dx);
		//$('#canvas_container').scrollTop(container_y - dy);

	}
	has_moved = 1;
}

function mouse_wheel( event )
{
	console.log("scrolling");
	var delta = 0;
	if (!event)
   		event = window.event;
   	
   	if (event.wheelDelta) 
    { 
    	delta = event.wheelDelta/120;
    } else 
    if (event.detail) 
    {
    	delta = -event.detail/3;
    }

	var zoom_factor = 1 + delta * 0.05;

    zoom *= zoom_factor;       

    if (zoom < min_zoom) 
    	zoom = min_zoom;
    if (zoom > max_zoom)
        zoom = max_zoom;
    resize_canvas();    
	// Recompute offsets for new zoom 
	var canvas = document.getElementById("canvas");
        
    var cx = canvas.width / 2.0;
	var cy = canvas.height / 2.0;    
        
    x_ofs = (x_ofs - cx) * zoom_factor + cx;
    y_ofs = (y_ofs - cy) * zoom_factor + cy;
    console.log(x_ofs);
    console.log(y_ofs);    
                
    if (event.preventDefault) event.preventDefault();
	
	event.returnValue = false;
	requestAnimationFrame(draw_display);
}

/*
	End keyboard / Mouse
*/

/*
	UI Service 
*/
function ui_add_pipe1(pos) 
{
	//add_object(pos.x, pos.y, "pipe");

	var i = find_object(pos.x, pos.y) 
	if (i != -1 && obj[i].Type == "guide")
	{
		console.log("adding 1 pipe");
		obj[i].selected  =1;

		set_mode("add_pipe2");
		sel_obj = i;

		requestAnimationFrame(draw_display);
	} else
		set_mode("none");
}

function ui_add_pipe2(pos)
{
	var i = find_object(pos.x, pos.y) 
	console.log("find2:", i);
	if (i != -1 && i != sel_obj)
	{
		if (obj[sel_obj].Type == "guide" && obj[i].Type == "guide")
		{
			object_connect(sel_obj, i);
			requestAnimationFrame(draw_display);
		}

		obj[i].selected  =0;
		obj[sel_obj].selected  =0;

		//set_mode("add_pipe");
		set_mode("none");
	} /*else
		set_mode ("none");*/
}

function ui_move_object(pos, i)
{
	if (obj[i].Attached == -1)
	{
		obj_x_ofs = get_world_x(pos.x) - obj[i].Xpos;
		obj_y_ofs = get_world_y(pos.y) - obj[i].Ypos;

		//obj[i].selected = 1;//!obj[i].selected;

		requestAnimationFrame(draw_display);

		sel_obj = i;
		set_mode("moving");
	} else {

	}
}

function ui_delete_object(pos)
{
	var i = find_object(pos.x, pos.y);

	if (i != -1)
		delete_object(i);

	set_mode("none");
}

function ui_unhook_object(pos)
{
	var i = find_object(pos.x, pos.y);

	if (i != -1)
	{
		unhook_object(i);
	}

	set_mode("none");
}

function ui_add_object(pos)
{
	var index = add_object(obj, pos.x, pos.y, ui_addtype, 0, -1);
	set_mode("none");
	ui_addtype = "";

}

function set_guide(s)
{
	if (s == "show") show_guide = 1;
	if (s == "hide") show_guide = 0;

	requestAnimationFrame(draw_display);
}

function select_none()
{
	// more like blank_properties
	for (var i in obj)
		obj[i].selected = 0;
	
	sel_obj = -1;
	property_window.set('current_obj', {});
}

function select_object( o )
{
	if (o.selected == 1)
	{
		object_toggle(o);
		return;
	}

	// Clear all selections
	for (var i in obj)
		obj[i].selected = 0;
	
		
	o.selected = 1;
	property_window.set('current_obj', o);	

	requestAnimationFrame(draw_display);
}

function save_properties(sel_obs)
{
	if (sel_obj == -1) return;
	
	var o = obj[sel_obj];

	// Save all properties
	for (var i = 0; i < o.PropertyCount; i++)
	{
		o.PropertyValues[i] = get_value(o.PropertyNames[i] + "_field");
	}

	// Need to save peoperties to database

	o.save_properties();
		
	requestAnimationFrame(draw_display);

	return false;
}

function set_mode(m)
{
	if (m == 'delete' || m == 'unhook' || m == 'add_pipe' || m == 'add_pipe2')
		$('#canvas').css('cursor', 'crosshair');
	else
		$('#canvas').css('cursor', 'auto');
	ui_mode = m;
}

function set_addmode(obj_type)
{
	var index = add_object(obj, get_world_x(100), get_world_y(100), obj_type, 0, 0, -1);
}

function reset()
{
	//obj = [];
	zoom = 1;
	x_ofs = 0;
	y_ofs = 0;
	//requestAnimationFrame(draw_display);
}
