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