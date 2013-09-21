/* 
	logic engine
	(C) 2013 Jason Hunt
	nulluser@gmail.com
*/

"use strict";

// Parameters
var snap = 5;
var min_zoom = 0.05;
var update_rate = 10;

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

var updates = 0;

// Pan / Zoom
var zoom = 1;
var x_ofs = 0;
var y_ofs = 0;

/* 
	High level UI 
*/
function start()
{
	// Map mouse functions
	document.onmouseup = mouse_up; 
	document.onmousedown = mouse_down; 
	document.onmousemove = mouse_move; 

	// Mouse wheel, FF
	if (window.addEventListener)
        window.addEventListener('DOMMouseScroll', mouse_wheel, false);

	// Mouse Wheel IE
	window.onmousewheel = document.onmousewheel = mouse_wheel;
	document.onkeydown = key_down; 

	window.onresize =  resize_canvas;

	resize_canvas();
	set_timer();
	update();
	//test_scope();
	backend_start();
	data_source_start();
}

/* 
	Utility 
*/
function snap_val(x)
{
	return (Math.round((x)/snap)*snap);
}

function debug(m)
{
	document.getElementById("debug").innerHTML =m + "<br>" + 	document.getElementById("debug").innerHTML;
	//alert(m);
}

function show_div(name)
{
	document.getElementById(name).style.display = "block";
}

function hide_div(name)
{
	document.getElementById(name).style.display = "none";
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
function resize_canvas()
{
	var canvas = document.getElementById("canvas");

	canvas.width = window.innerWidth - $(canvas).offset().left - 10;
	canvas.height = window.innerHeight- $(canvas).offset().top - 10;
}

function draw_display()
{
 	var canvas = document.getElementById("canvas");
 	var ctx = canvas.getContext("2d");
	
	draw_objects(ctx, obj, canvas.width, canvas.height);
}

function update( )
{
	draw_display();
	
	updates++;
	
	//debug("Update: " + updates);
}

function set_timer() 
{ 
//	debug("set timer");
		
	setInterval(function() { update(); }.bind(this), 1000/update_rate);
}

/* 
	End of high level UI 
*/

/*
	Keyboard / Mouse 
*/

function mouse_pos(ev)
{
	var canvas_x_ofs = document.getElementById("canvas").offsetLeft;
	var canvas_y_ofs = document.getElementById("canvas").offsetTop;

	//console.log("x_ofs: " + canvas_x_ofs);
	//console.log("y_ofs: " + canvas_y_ofs);
	if(ev.pageX || ev.pageY)
	{ 
		return {x:ev.pageX - canvas_x_ofs, y:ev.pageY - canvas_y_ofs}; 
	} 
	//console.log("clienttop: " + document.body.clientTop);
	return 	{ x:ev.clientX + document.body.scrollLeft - document.body.clientLeft  - canvas_x_ofs, 
			  y:ev.clientY + document.body.scrollTop  - document.body.clientTop   - canvas_y_ofs}; 
}

function mouse_up(ev)
{ 
	var pos = mouse_pos(ev);    		
	//document.getElementById("debug").innerHTML += "Up  " + pos.x + " " + pos.y +"<br>";
	mouse_state = "up";

	if (ui_mode == "moving")
	{
		if (!has_moved)
		{
			select_object(obj[sel_obj]);
		}
		else
		{
			backend_moveobject(sel_obj, obj[sel_obj].id, obj[sel_obj].x_pos, obj[sel_obj].y_pos);
		
			for (var i in obj[sel_obj].guides)
			{
				var k = obj[sel_obj].guides[i];
				//debug (i);
				backend_moveobject(k, obj[k].id, obj[k].x_pos, obj[k].y_pos);
			}
		}
			
		set_mode("none");
	}
	draw_display();
} 

function mouse_down(ev)
{ 
 	var pos = mouse_pos(ev); 

	mouse_state = "down";
	
	x_ofs_start = x_ofs;	
	y_ofs_start = y_ofs;
	console.log("x : " + pos.x);
	console.log("y : " + pos.y);
	mouse_x = pos.x;
	mouse_y = pos.y;
	
	//document.getElementById("debug").innerHTML += "Down  " + pos.x + " " + pos.y +"<br>";
	//debug("mode: " + ui_mode);

	has_moved = 0;
	
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

		set_mode("none");

	draw_display();
} 




function mouse_move(ev)
{ 
 	var pos = mouse_pos(ev); 


	if (ui_mode == "moving")
	{
		
		var new_x = snap_val(get_world_x(pos.x) - obj_x_ofs);
		var new_y = snap_val(get_world_y(pos.y) - obj_y_ofs);

		var delta_x = new_x - obj[sel_obj].x_pos ;
		var delta_y = new_y - obj[sel_obj].y_pos;


		obj[sel_obj].x_pos = new_x;
		obj[sel_obj].y_pos = new_y;


		for(var j in obj[sel_obj].guides)
		{
			var k = obj[sel_obj].guides[j];

			obj[k].x_pos += delta_x;
			obj[k].y_pos += delta_y;
		}

		draw_display();
	} else


	// Pan grid if dragging mouse
	if (mouse_state == "down")
	{
		var dx = pos.x - mouse_x;		
		var dy = pos.y - mouse_y;
		
	
		x_ofs = x_ofs_start + dx;
		y_ofs = y_ofs_start + dy;
		
		//debug(x_ofs + " " + y_ofs);
	}


	has_moved = 1;
}


function mouse_wheel( event )
{
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

	var zoom_factor = 1 + delta; // * 0.05;

    zoom *= zoom_factor;       

    if (zoom < min_zoom) 
    	zoom = min_zoom;
   
            
	// Recompute offsets for new zoom 
	var canvas = document.getElementById("canvas");
        
    var cx = canvas.width / 2.0;
	var cy = canvas.height / 2.0;    
        
    x_ofs = (x_ofs - cx) * zoom_factor + cx;
    y_ofs = (y_ofs - cy) * zoom_factor + cy;
           
                
    if (event.preventDefault) event.preventDefault();
	
	event.returnValue = false;

	//debug("zoom: " + zoom);
}


function key_down(ev)
{
 	var ch = (typeof ev.which == "number") ? ev.which : ev.keyCode;
 	
 /*	if (ch == 46)  // Delete key
	{
		if (sel_obj < 0) return;

		delete_object(sel_obj);

		sel_obj = -1;

		set_mode("none");
	} 	
	
 	if (ch == 85)  // Delete key
	{
		if (sel_obj < 0) return;

		unhook_object(sel_obj);

		sel_obj = -1;

		set_mode("none");
	} 	
	if (ch == 87) // Wire command
	{
		set_mode('add_pipe');
	}
	*/
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

	if (i != -1 && obj[i].type == "guide")
	{
		console.log("adding 1 pipe");
		obj[i].selected  =1;

		ui_mode = "add_pipe2";
		sel_obj = i;

		draw_display();
	} else
		ui_mode = "none";
}


function ui_add_pipe2(pos)
{
	var i = find_object(pos.x, pos.y) 

	if (i != -1 && i != sel_obj)
	{
		if (obj[sel_obj].type == "guide" && obj[i].type == "guide")
		{
			object_connect(sel_obj, i);
			draw_display();
		}

		obj[i].selected  =0;
		obj[sel_obj].selected  =0;

		ui_mode = "add_pipe";
	} else
		set_mode ("none");
}


function ui_move_object(pos, i)
{
	if (obj[i].attached == -1)
	{
		obj_x_ofs = get_world_x(pos.x) - obj[i].x_pos;
		obj_y_ofs = get_world_y(pos.y) - obj[i].y_pos;

		//obj[i].selected = 1;//!obj[i].selected;

		draw_display();

		sel_obj = i;
		set_mode("moving");
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
	var index = add_object(obj, pos.x, pos.y, ui_addtype, 0);
	obj[index].backend_add();
	ui_mode = "none";
	ui_addtype = "";

}

function set_guide(s)
{
	if (s == "show") show_guide = 1;
	if (s == "hide") show_guide = 0;

	draw_display();
}

function select_none()
{
	//hide_properties();
	// more like blank_properties
	for (var i in obj)
		obj[i].selected = 0;
	
	sel_obj = -1;
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
	
	show_properties(o);
	

	draw_display();
}

function show_properties(o)
{
	show_div("property_area");

	var p_str = "";

	// Create property table
	
	p_str += "<table class='property_table'><tbody>";
	p_str += "<tr><td><b>Type</b></td><td><b>" + o.type + "</b></td></tr>";

	for (var i = 0; i < o.property_count; i++)
	{
		p_str += "<tr><td><b>" + o.property_names[i] + " </b></td><td>";
		p_str += "<input id='"    + o.property_names[i] + "_field' size='6&quot;' type='text' value='" + o.property_values[i] + "'></td></tr>";
	}

	p_str += "</tbody></table>";
		
	if (o.property_count > 0)
		p_str += "<input value='Save' onclick='save_properties(" +sel_obj + ");' type='submit'>  ";
	
	document.getElementById("property_area").innerHTML = p_str;
}

function hide_properties()
{
	hide_div("property_area");
}

function save_properties(sel_obs)
{
	if (sel_obj == -1) return;
	
	var o = obj[sel_obj];

	// Save all properties
	for (var i = 0; i < o.property_count; i++)
	{
		o.property_values[i] = get_value(o.property_names[i] + "_field");
	}

	// Need to save peoperties to database

	o.save_properties();
	
	//hide_properties();
	
	draw_display();

	return false;
}


function set_mode(m)
{
	ui_mode = m;
}

function set_addmode(obj_type)
{
	//ui_mode = "add_object";
	//ui_addtype = obj_type;
	
	var index = add_object(obj, get_world_x(100), get_world_y(100), obj_type, 0, 0);
	obj[index].backend_add();
}

function reset()
{
	obj = [];

	draw_display();	
	
	zoom = 1;
	x_ofs = 0;
	y_ofs = 0;
}

$(start);
