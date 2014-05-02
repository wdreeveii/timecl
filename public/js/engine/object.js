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

		var f_size = 12 * engine.get_zoom();

		ctx.font = format(f_size) + "pt Arial";

		var name = this.get_property("name");

		ctx.fillText(name, engine.get_x(x + this.Xsize * 0.1), engine.get_y(y) - f_size / 2);

		if (this.ShowOutput)
			ctx.fillText(format(this.Output), engine.get_x(x + this.Xsize * 0.3), engine.get_y(y + this.Ysize / 2) + f_size / 2);

		if (this.ShowAnalog)
			ctx.fillText(format(this.Output), engine.get_x(x + this.Xsize * 0.1), engine.get_y(y + this.Ysize / 2) + f_size / 2);

		if (this.ShowName)
			ctx.fillText(this.show_name, engine.get_x(x + this.Xsize * 0.1), engine.get_y(y + this.Ysize / 2) - f_size / 2 * 1.5);

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
	ctx.moveTo(Math.round(engine.get_x(o.Xpos)), Math.round(engine.get_y(o.Ypos)));
	ctx.lineTo(Math.round(engine.get_x(o.Xpos + o.Xsize)), Math.round(engine.get_y(o.Ypos)));
	ctx.lineTo(Math.round(engine.get_x(o.Xpos + o.Xsize)), Math.round(engine.get_y(o.Ypos + o.Ysize)));
	ctx.lineTo(Math.round(engine.get_x(o.Xpos)), Math.round(engine.get_y(o.Ypos + o.Ysize)));

	ctx.closePath();

	ctx.fill();
	ctx.stroke();
}

function construct(new_object) {
	if (typeof(window[new_object.Type + "_type"]) == 'function') {
		window[new_object.Type + "_type"](new_object);
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



function object_connect(o1, o2) {
	engine.obj[o2].Source = o1;
	backend_hookobject(engine.obj[o2].Id, engine.obj[o1].Id)
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



/***
	End of object methods 
***/