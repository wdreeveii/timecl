/* 
	logic engine
	(C) 2013 Jason Hunt
	nulluser@gmail.com
*/


/*
	Object 
*/

"use strict";

var dir_type = new Object();

dir_type.none = 0;
dir_type.up = 1;
dir_type.down = 2;
dir_type.left = 3;
dir_type.right = 4;

var object_list = Array(); // List of all defined objects

function object_type(name) {
	this.Xpos = 0;
	this.Ypos = 0;
	this.Xsize = 50;
	this.Ysize = 50;

	this.Type = "none";

	this.Id = 0;

	this.input_termcount = 0;
	this.output_termcount = 0;

	this.Selected = 0;
	this.Terminals = new Array;
	this.Attached = -1;
	this.Dir = dir_type.none;
	this.Output = 0;
	this.NextOutput = 0;
	this.Source = -1;

	this.show_name = 0;
	this.show_analog = 0;
	this.show_output = 0;

	this.PropertyCount = 0;
	this.PropertyNames = new Array;
	this.PropertyTypes = new Array;
	this.PropertyValues = new Array;

	this.draw_icon = function(ctx) {
		bounding_rect(ctx, this);
	}

	this.set_output = function(output) {
		this.Output = output;

		backend_setoutput(this.Id, this.Output);
	}

	this.save_properties = function() {
		backend_setproperties(this.Id,
			this.PropertyCount,
			this.PropertyNames,
			this.PropertyTypes,
			this.PropertyValues);
	}


	this.draw_properties = function(ctx, x, y) {
		var old_fill = ctx.fillStyle;

		ctx.fillStyle = "rgb(0,0,0)";
		ctx.font = "16pt Arial";

		var f_size = 12 * zoom;

		ctx.font = format(f_size) + "pt Arial";

		var name = this.get_property("name");

		ctx.fillText(name, get_x(x + this.Xsize * 0.1), get_y(y) - f_size / 2);

		if (this.ShowOutput)
			ctx.fillText(format(this.Output), get_x(x + this.Xsize * 0.3), get_y(y + this.Ysize / 2) + f_size / 2);

		if (this.ShowAnalog)
			ctx.fillText(format(this.Output), get_x(x + this.Xsize * 0.1), get_y(y + this.Ysize / 2) + f_size / 2);

		if (this.ShowName)
			ctx.fillText(this.show_name, get_x(x + this.Xsize * 0.1), get_y(y + this.Ysize / 2) - f_size / 2 * 1.5);

		ctx.fillStyle = old_fill;
	}

	this.get_property = function(property) {
		if (this.PropertyCount <= 0) return "";


		for (var i = 0; i < this.PropertyCount; i++) {
			if (this.PropertyNames[i] == property)
				return this.PropertyValues[i];
		}
		return "";
	}
}

function bounding_rect(ctx, o) {
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

function construct(new_object) {
	for (var i = 0; i < object_list.length; i++) {
		if (new_object.Type == object_list[i]) {
			window[new_object.Type + "_type"](new_object);
			break;
		}
	}
}

function load_object(newobject) {
	var o = new object_type;
	$.extend(o, newobject);
	construct(o);
	return o;
}

function load_objects(objects) {
	var results = Array();
	for (var i = 0; i < objects.length; i++) {
		var o = new object_type;
		$.extend(o, objects[i]);
		construct(o);
		results[o["Id"]] = o;
	}
	return results;
}

function find_object(x, y) {
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

function object_connect(o1, o2) {
	obj[o2].Source = o1;
	backend_hookobject(obj[o2].Id, obj[o1].Id)
}

function object_toggle(o) {
	var tmp = 0;

	if (o.Output > 0.5)
		tmp = 0;
	else
		tmp = 1;

	o.set_output(tmp);

	o.Output = tmp;
}

var handle_size = 10;
var guide_size = 10;


function format(n) {
	return (Number(n).toFixed(2));
} // Float to formatted string
function bformat(n) {
	return (Number(n).toFixed(0));
} // Float to single digit string

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


function draw_wire(ctx, objects, i1, i2, p1, p2) {
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

function draw_object(ctx, o, x_size, y_size) {
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

function draw_objects(ctx, objects, x_size, y_size) {
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

/***
	End of object methods 
***/