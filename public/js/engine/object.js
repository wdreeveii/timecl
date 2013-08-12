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

var object_list = Array();  // List of all defined objects

function object_type(name)
{
	this.x_pos = 0;
	this.y_pos = 0;
	this.x_size = 50;
	this.y_size = 50;

	this.type = "none";
	
	this.id = 0;
	this.index = 0;

	this.input_termcount = 0;
	this.output_termcount = 0;

	this.selected = 0;
	this.guides = new Array;
	this.attached = -1;
	this.solved = 0;
	this.dir = dir_type.none;
	this.output = 0;
	this.next_output = 0;
	this.source = -1;

	this.show_name = 0;
	this.show_analog = 0;
	this.show_output = 0;

	this.draw_icon = function(ctx) { 		bounding_rect(ctx, this); } 

	this.property_count = 0;
	this.property_names = new Array;
	this.property_types = new Array;
	this.property_values = new Array;
	
	this.root_id = -1;
	
	this.add_output_terminal = function(objects, pos)
	{
		var index = add_object(objects,
							   this.x_pos + this.x_size,
							   this.y_pos + this.y_size/2 - guide_size/2 + pos * (guide_size+2), 
						       "guide", 1, dir_type.right);

		this.guides.push(index);
	}


	this.add_input_terminal = function(objects, pos)
	{
		var index = add_object(objects,
						      this.x_pos-guide_size, 
						      this.y_pos + this.y_size/2 - guide_size/2 + pos * (guide_size+2), 
							  "guide", 1, dir_type.left);

		this.guides.push(index);
	}
	
	this.set_output = function(output)
	{
		this.output = output;
		
		backend_setoutput(this.index, this.id, this.output);
	}

	this.save_properties = function()
	{
		backend_setproperties(this.index, this.id, 
								this.property_count, 
								this.property_names.join(','), 
								this.property_types.join(','), 
								this.property_values.join(','));
	}


	this.add_property = function(name, type, value)
	{
		this.property_names[this.property_count] = name;
		this.property_types[this.property_count] = type;
		this.property_values[this.property_count] = value;
		
		this.property_count++;
	}
	
	
	this.draw_properties = function(ctx, x, y)
	{
	//alert("draw");
		var old_fill  = ctx.fillStyle;
	
		ctx.fillStyle = "rgb(0,0,0)";
		ctx.font = "16pt Arial";

		var f_size = 12 * zoom;

		ctx.font = format(f_size) + "pt Arial";
	
		var name = this.get_property("name");
	
	   	ctx.fillText(name, get_x(x + this.x_size * 0.1 ), get_y(y) - f_size  / 2);
	
		if (this.show_output)
	    	ctx.fillText(bformat(this.output), get_x(x + this.x_size * 0.3 ), get_y(y + this.y_size/2) + f_size  / 2);
	
		if (this.show_analog)
	    	ctx.fillText(format(this.output), get_x(x + this.x_size * 0.1 ), get_y(y + this.y_size/2) + f_size  / 2);
	
		if (this.show_name)
	    	ctx.fillText(this.type, get_x(x + this.x_size * 0.1 ), get_y(y + this.y_size/2) - f_size  / 2 * 1.5);
		    
		ctx.fillStyle = old_fill;
	}	

	this.backend_add = function ()
	{
		var o = this;
		var donefunc = function (response) {
			o.id = response;
			if (o.guides) {
				for (var ii = 0; ii < o.guides.length; ii++) {
					var idx = o.guides[ii];
					obj[idx].root_id = response;
					obj[idx].backend_add();
				}
			}
		}
		backend_addobject(this, donefunc);
	}	
	
	this.get_property = function (property)
	{
		if (this.property_count <= 0) return "";

	
		for (var i = 0; i < this.property_count; i++)
		{
			if (this.property_names[i] == property)
				return this.property_values[i];
		}
		return "";
	}
}

function bounding_rect(ctx, o)
{
	ctx.beginPath();	
	ctx.moveTo(get_x(o.x_pos), get_y(o.y_pos));		
	ctx.lineTo(get_x(o.x_pos+o.x_size), get_y(o.y_pos));
	ctx.lineTo(get_x(o.x_pos+o.x_size), get_y(o.y_pos+o.y_size));
	ctx.lineTo(get_x(o.x_pos), get_y(o.y_pos+o.y_size));
		
	ctx.closePath();	
		
	ctx.fill();
	ctx.stroke();	
}
function init_obj(x, y, type, attached, dir, index) {
	var o = new object_type;
	o.type = type;
	o.x_pos = x;
	o.y_pos = y;
	o.dir = dir;
	o.index = index;
	if (type == "guide")
		o.attached = attached;
	return o;
}

function construct(new_object) {
	for (var i = 0; i < object_list.length; i++) {
		if (new_object.type == object_list[i]) {
			window[new_object.type + "_type"](new_object);
			break;
		}
	}
}

function add_object(objects, x, y, type, attached, dir)
{
	var index = objects.length;
	var o = init_obj(x, y, type, attached, dir, index);
	objects.push(o);
	construct(o);
	
	// Add terminals
	// HACK
	for (var i = 0; i < o.input_termcount; i++)
	{
		if (o.input_termcount == 1)
			o.add_input_terminal(objects, 0);
		
		if (o.input_termcount == 2)
			o.add_input_terminal(objects, 2*i - 1);
	}
		
	for (var i = 0; i < o.output_termcount; i++)
	{
		if (o.output_termcount == 1)
			o.add_output_terminal(objects, 0);

		if (o.output_termcount == 2)
			o.add_output_terminal(objects, 2*i - 1);
	}
	console.log(o.guides);
	return index;
}

function load_object(objects, x, y, type, attached, dir)
{
	var index = objects.length;
	var o = init_obj(x, y, type, attached, dir, index);
	objects.push(o);
	construct(o);
	return index;
}

function find_object(x, y)
{
	// convert to world cords
	x = get_world_x(x);
	y = get_world_y(y);

	for (var i in obj)
	{
		if (x > obj[i].x_pos &&
			y > obj[i].y_pos &&
			x < obj[i].x_pos + obj[i].x_size &&
			y < obj[i].y_pos + obj[i].y_size )
			return(i);
	}

	return (-1);
}

function unhook_object(i)
{
	// unhook objects linked to this object
	for (var j in obj)
		if (obj[j].source == i)
		{
			obj[j].source = -1;
		
			backend_unhookobject(j, obj[j].id);
		}
			
			
	// unhook main source
	obj[i].source = -1;
		
	backend_unhookobject(i, obj[i].id);
	
	for (var j in obj[i].guides)
	{
		unhook_guide(obj[i].guides[j]);
		
		obj[obj[i].guides[j]].source=-1;
		
		backend_unhookobject(obj[i].guides[j], obj[obj[i].guides[j]].id);
		
	}
}

function unhook_guide(i)
{	
	for (var j in obj)
		if (obj[j].source == i)	
		{	
			obj[j].source = -1;
			backend_unhookobject(j, obj[j].id);
		}			
			
	obj[i].source = -1;
}

function delete_object(i)
{
 	unhook_object(i);

	for (var j in obj[i].guides)
		delete_object(obj[i].guides[j]);

	backend_deleteobject(i, obj[i].id);

	delete obj[i];
//	draw_objects();
}

function object_connect(o1, o2)
{	
	obj[o2].source = o1;
	backend_hookobject(o1, obj[o2].id, obj[o1].id)
}

function object_toggle(o)
{
	var tmp = 0;

	if (o.output > 0.5) 
		tmp = 0;
	else
		tmp = 1;
		
	o.set_output(tmp);
		
	o.output = tmp;
}

/***
	End of object methods 
***/




