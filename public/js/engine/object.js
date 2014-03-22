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
	this.Xpos = 0;
	this.Ypos = 0;
	this.Xsize = 50;
	this.Ysize = 50;

	this.Type = "none";
	
	this.Id = 0;
	this.index = 0;

	this.input_termcount = 0;
	this.output_termcount = 0;

	this.selected = 0;
	this.Terminals = new Array;
	this.Attached = -1;
	this.solved = 0;
	this.Dir = dir_type.none;
	this.Output = 0;
	this.NextOutput = 0;
	this.Source = -1;

	this.show_name = 0;
	this.show_analog = 0;
	this.show_output = 0;

	this.draw_icon = function(ctx) { 		bounding_rect(ctx, this); } 

	this.PropertyCount = 0;
	this.PropertyNames = new Array;
	this.PropertyTypes = new Array;
	this.PropertyValues = new Array;
	
	this.RootId = -1;
	
	this.add_output_terminal = function(objects, pos, root_id)
	{
		var index = add_object(objects,
							   this.Xpos + this.Xsize,
							   this.Ypos + this.Ysize/2 - guide_size/2 + pos * (guide_size+2), 
						       "guide", 1, dir_type.right, root_id);

		this.Terminals.push(index);
	}


	this.add_input_terminal = function(objects, pos, root_id)
	{
		var index = add_object(objects,
						      this.Xpos-guide_size, 
						      this.Ypos + this.Ysize/2 - guide_size/2 + pos * (guide_size+2), 
							  "guide", 1, dir_type.left, root_id);

		this.Terminals.push(index);
	}
	
	this.set_output = function(output)
	{
		this.Output = output;
		
		backend_setoutput(this.Id, this.Output);
	}

	this.save_properties = function()
	{
		backend_setproperties(this.Id, 
								this.PropertyCount, 
								this.PropertyNames, 
								this.PropertyTypes, 
								this.PropertyValues);
	}


	this.add_property = function(name, type, value)
	{
		this.PropertyNames[this.PropertyCount] = name;
		this.PropertyTypes[this.PropertyCount] = type;
		this.PropertyValues[this.PropertyCount] = value;
		
		this.PropertyCount++;
	}
	
	
	this.draw_properties = function(ctx, x, y)
	{
		var old_fill  = ctx.fillStyle;
	
		ctx.fillStyle = "rgb(0,0,0)";
		ctx.font = "16pt Arial";

		var f_size = 12 * zoom;

		ctx.font = format(f_size) + "pt Arial";
	
		var name = this.get_property("name");
	
	   	ctx.fillText(name, get_x(x + this.Xsize * 0.1 ), get_y(y) - f_size  / 2);
	
		if (this.show_output)
	    	ctx.fillText(format(this.Output), get_x(x + this.Xsize * 0.3 ), get_y(y + this.Ysize/2) + f_size  / 2);
	
		if (this.show_analog)
	    	ctx.fillText(format(this.Output), get_x(x + this.Xsize * 0.1 ), get_y(y + this.Ysize/2) + f_size  / 2);
	
		if (this.show_name)
	    	ctx.fillText(this.show_name, get_x(x + this.Xsize * 0.1 ), get_y(y + this.Ysize/2) - f_size  / 2 * 1.5);
	    
		ctx.fillStyle = old_fill;
	}	

	this.backend_add = function ()
	{
		backend_addobject(this);
	}	
	
	this.get_property = function (property)
	{
		if (this.PropertyCount <= 0) return "";

	
		for (var i = 0; i < this.PropertyCount; i++)
		{
			if (this.PropertyNames[i] == property)
				return this.PropertyValues[i];
		}
		return "";
	}
}

function bounding_rect(ctx, o)
{
	ctx.lineWidth=0.25;
	ctx.beginPath();	
	ctx.moveTo(Math.round(get_x(o.Xpos)), Math.round(get_y(o.Ypos)));		
	ctx.lineTo(Math.round(get_x(o.Xpos+o.Xsize)), Math.round(get_y(o.Ypos)));
	ctx.lineTo(Math.round(get_x(o.Xpos+o.Xsize)), Math.round(get_y(o.Ypos+o.Ysize)));
	ctx.lineTo(Math.round(get_x(o.Xpos)), Math.round(get_y(o.Ypos+o.Ysize)));
		
	ctx.closePath();	
		
	ctx.fill();
	ctx.stroke();	
}
function init_obj(x, y, type, attached, dir, index, root_id) {
	var o = new object_type;
	o.Type = type;
	o.Xpos = x;
	o.Ypos = y;
	o.Dir = dir;
	o.Id = index;
	o.RootId = root_id;
	o.index = index;
	if (type == "guide")
		o.Attached = attached;
	return o;
}

function construct(new_object) {
	for (var i = 0; i < object_list.length; i++) {
		if (new_object.Type == object_list[i]) {
			window[new_object.Type + "_type"](new_object);
			break;
		}
	}
}

function add_object(objects, x, y, type, attached, dir, root_id)
{
	var index = objects.length;
	var o = init_obj(x, y, type, attached, dir, index, root_id);
	objects.push(o);
	construct(o);
	// Add terminals
	// HACK
	for (var i = 0; i < o.input_termcount; i++)
	{
		if (o.input_termcount == 1)
			o.add_input_terminal(objects, 0, index);
		
		if (o.input_termcount == 2)
			o.add_input_terminal(objects, 2*i - 1, index);
	}
		
	for (var i = 0; i < o.output_termcount; i++)
	{
		if (o.output_termcount == 1)
			o.add_output_terminal(objects, 0, index);

		if (o.output_termcount == 2)
			o.add_output_terminal(objects, 2*i - 1, index);
	}
	objects[index].backend_add()
	return index;
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

function find_object(x, y)
{
	// convert to world cords
	x = get_world_x(x);
	y = get_world_y(y);
	var keys = [], ii = 0;
	for (keys[ii++] in obj) {}
	keys.reverse();
	for (var k in keys)
	{
		var i = keys[k];
		if (x >= obj[i].Xpos &&
			y >= obj[i].Ypos &&
			x <= obj[i].Xpos + obj[i].Xsize &&
			y <= obj[i].Ypos + obj[i].Ysize ) {
			
			return(i);
		}
	}

	return (-1);
}

function unhook_object(i)
{
	if (i in obj) {
		// unhook objects linked to this object
		for (var j in obj)
			if (obj[j].Source == i)
			{
				obj[j].Source = -1;
			
				backend_unhookobject(obj[j].Id);
			}
				
				
		// unhook main source
		obj[i].Source = -1;
			
		backend_unhookobject(obj[i].Id);
		
		for (var j in obj[i].Terminals)
		{
			unhook_guide(obj[i].Terminals[j]);
			if (obj[i].Terminals[j] in obj) {
				obj[obj[i].Terminals[j]].source=-1;
				backend_unhookobject(obj[obj[i].Terminals[j]].Id);
			}
		}
	}
}

function unhook_guide(i)
{	
	for (var j in obj)
		if (obj[j].Source == i)	
		{	
			obj[j].Source = -1;
			backend_unhookobject(obj[j].Id);
		}			
	if (i in obj) {		
		obj[i].Source = -1;
	}
}

function delete_object(i)
{
	if (i in obj) {
	 	unhook_object(i);

		for (var j in obj[i].Terminals)
			delete_object(obj[i].Terminals[j]);

		backend_deleteobject(obj[i].Id);

		delete obj[i];
	}
}

function object_connect(o1, o2)
{	
	obj[o2].Source = o1;
	backend_hookobject(obj[o2].Id, obj[o1].Id)
}

function object_toggle(o)
{
	var tmp = 0;

	if (o.Output > 0.5) 
		tmp = 0;
	else
		tmp = 1;
		
	o.set_output(tmp);
		
	o.Output = tmp;
}

/***
	End of object methods 
***/




