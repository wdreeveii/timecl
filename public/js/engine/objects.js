/* 
	logic engine
	(C) 2013 Jason Hunt
	nulluser@gmail.com
*/


/*
	Object 
*/

"use strict";

function binput_type(o) {
	o.save_properties = function() {
		backend_setproperties(o.Id, o.PropertyCount, o.PropertyNames, o.PropertyTypes, o.PropertyValues);

		for (var i = 0; i < o.PropertyCount; i++) {
			if (o.PropertyNames[i] == "value") {
				backend_setoutput(o.Id, o.PropertyValues[i]);
				break;
			}
		}
	}

	o.set_output = function(output) {
		o.Output = output;

		backend_setoutput(o.Id, o.Output);

		for (var i = 0; i < o.PropertyCount; i++) {
			if (o.PropertyNames[i] == "value") {
				o.PropertyValues[i] = o.Output;
				break;
			}
		}

		backend_setproperties(o.Id, o.PropertyCount, o.PropertyNames, o.PropertyTypes, o.PropertyValues);
	}

	o.draw_icon = function(ctx) {
		ctx.beginPath();

		ctx.arc(engine.get_x(this.Xpos + this.Xsize / 2),
		engine.get_y(this.Ypos + this.Ysize / 2), (this.Xsize / 2) * engine.get_zoom(), 0, Math.PI * 2, true);

		ctx.fill();
		ctx.stroke();
	}
}

function ainput_type(o) {
	o.save_properties = function() {
		backend_setproperties(o.Id, o.PropertyCount, o.PropertyNames, o.PropertyTypes, o.PropertyValues);

		for (var i = 0; i < o.PropertyCount; i++) {
			if (o.PropertyNames[i] == "value") {
				backend_setoutput(o.Id, o.PropertyValues[i]);
				break;
			}
		}
	}

	o.set_output = function(output) {
		o.Output = output;

		backend_setoutput(o.Id, o.Output);

		for (var i = 0; i < o.PropertyCount; i++) {
			if (o.PropertyNames[i] == "value") {
				o.PropertyValues[i] = o.Output;
				break;
			}
		}

		backend_setproperties(o.Id, o.PropertyCount, o.PropertyNames, o.PropertyTypes, o.PropertyValues);
	}

	o.draw_icon = function(ctx) {
		ctx.beginPath();

		ctx.arc(engine.get_x(this.Xpos + this.Xsize / 2),
			engine.get_y(this.Ypos + this.Ysize / 2), (this.Xsize / 2) * engine.get_zoom(), 0, Math.PI * 2, true);

		ctx.fill();
		ctx.stroke();
	}
}

function boutput_type(o) {
	o.draw_icon = function(ctx) {
		bounding_rect(ctx, this);
	}
}

function aoutput_type(o) {
	o.draw_icon = function(ctx) {
		bounding_rect(ctx, this);
	}
}

function notgate_type(o) {
	o.draw_icon = function(ctx) {
		ctx.beginPath();
		ctx.moveTo(Math.round(engine.get_x(this.Xpos)), Math.round(engine.get_y(this.Ypos)));
		ctx.lineTo(Math.round(engine.get_x(this.Xpos)), Math.round(engine.get_y(this.Ypos + this.Ysize)));
		ctx.lineTo(Math.round(engine.get_x(this.Xpos + this.Xsize)), Math.round(engine.get_y(this.Ypos + this.Ysize / 2)));
		ctx.closePath();

		ctx.fill();
		ctx.stroke();
	}
}

function andgate_type(o) {
	o.draw_icon = function(ctx) {
		ctx.beginPath();
		ctx.moveTo(engine.get_x(this.Xpos), engine.get_y(this.Ypos));
		ctx.lineTo(engine.get_x(this.Xpos), engine.get_y(this.Ypos + this.Ysize));
		ctx.arc(engine.get_x(this.Xpos + this.Xsize / 2), engine.get_y(this.Ypos + this.Ysize / 2), this.Xsize / 2 * engine.get_zoom(), 0.5 * Math.PI, 1.5 * Math.PI, true);

		ctx.closePath();

		ctx.fill();
		ctx.stroke();
	}
}

function orgate_type(o) {
	o.draw_icon = function(ctx) {
		ctx.beginPath();

		ctx.arc(engine.get_x(this.Xpos - this.Xsize * 1.15),
			engine.get_y(this.Ypos + this.Ysize / 2),
			this.Xsize * 1.2 * engine.get_zoom(),
			0.10 * Math.PI, 1.90 * Math.PI, true);

		ctx.arc(engine.get_x(this.Xpos + this.Xsize * 0.05),
			engine.get_y(this.Ypos + this.Ysize / 2 + this.Ysize * 0.6),
			this.Xsize * 1.1 * engine.get_zoom(),
			1.45 * Math.PI, 1.80 * Math.PI, false);

		ctx.arc(engine.get_x(this.Xpos + this.Xsize * 0.05),
			engine.get_y(this.Ypos + this.Ysize / 2 - this.Ysize * 0.6),
		this.Xsize * 1.1 * engine.get_zoom(),
			0.20 * Math.PI, 0.55 * Math.PI, false);

		ctx.closePath();

		ctx.fill();
		ctx.stroke();
	}
}

function xorgate_type(o) {
	o.draw_icon = function(ctx) {
		ctx.beginPath();

		ctx.arc(engine.get_x(this.Xpos - this.Xsize * 1.15),
			engine.get_y(this.Ypos + this.Ysize / 2),
			this.Xsize * 1.2 * engine.get_zoom(),
			0.10 * Math.PI, 1.90 * Math.PI, true);

		ctx.arc(engine.get_x(this.Xpos + this.Xsize * 0.05),
			engine.get_y(this.Ypos + this.Ysize / 2 + this.Ysize * 0.6),
			this.Xsize * 1.1 * engine.get_zoom(),
			1.45 * Math.PI, 1.80 * Math.PI, false);

		ctx.arc(engine.get_x(this.Xpos + this.Xsize * 0.05),
			engine.get_y(this.Ypos + this.Ysize / 2 - this.Ysize * 0.6),
			this.Xsize * 1.1 * engine.get_zoom(),
			0.20 * Math.PI, 0.55 * Math.PI, false);
		ctx.closePath();
		ctx.fill();
		ctx.stroke();

		ctx.beginPath();

		ctx.arc(engine.get_x(this.Xpos - this.Xsize * 1.15 + 5),
			engine.get_y(this.Ypos + this.Ysize / 2),
			this.Xsize * 1.2 * engine.get_zoom(),
			0.10 * Math.PI, 1.90 * Math.PI, true);

		ctx.stroke();
	}
}

function mult_type(o) {
	o.draw_icon = function(ctx) {
		bounding_rect(ctx, this);
	}
}

function div_type(o) {
	o.draw_icon = function(ctx) {
		bounding_rect(ctx, this);
	}
}

function add_type(o) {
	o.draw_icon = function(ctx) {
		bounding_rect(ctx, this);
	}
}

function sub_type(o) {
	o.draw_icon = function(ctx) {
		bounding_rect(ctx, this);
	}
}

function power_type(o) {
	o.draw_icon = function(ctx) {
		bounding_rect(ctx, this);
	}
}

function sine_type(o) {
	o.draw_icon = function(ctx) {
		bounding_rect(ctx, this);
	}
}

function cosine_type(o) {
	o.draw_icon = function(ctx) {
		bounding_rect(ctx, this);
	}
}

function agtb_type(o) {
	o.draw_icon = function(ctx) {
		bounding_rect(ctx, this);
	}
}

function agteb_type(o) {
	o.draw_icon = function(ctx) {
		bounding_rect(ctx, this);
	}
}

function altb_type(o) {
	o.draw_icon = function(ctx) {
		bounding_rect(ctx, this);
	}
}

function alteb_type(o) {
	o.draw_icon = function(ctx) {
		bounding_rect(ctx, this);
	}
}

function aeqb_type(o) {
	o.draw_icon = function(ctx) {
		bounding_rect(ctx, this);
	}
}

function aneqb_type(o) {
	o.draw_icon = function(ctx) {
		bounding_rect(ctx, this);
	}
}

function xyscope_type(o) {
	o.draw_icon = function(ctx) {
		bounding_rect(ctx, this);

		if (this.Terminals.length < 2) return;

		var dot_scale = 0.03;

		var x = obj[this.Terminals[0]].Output;
		var y = obj[this.Terminals[1]].Output;

		if (x < -1) x = -1;
		if (x > 1) x = 1;

		if (y < -1) y = -1;
		if (y > 1) y = 1;

		ctx.beginPath();

		var old_fill = ctx.fillStyle;
		var old_stroke = ctx.strokeStyle;

		ctx.strokeStyle = "rgb(0,0,0)";
		ctx.fillStyle = "rgb(0, 0, 0)";

		ctx.arc(engine.get_x(this.Xpos + this.Xsize / 2 + x * this.Xsize / 2.5),
			engine.get_y(this.Ypos + this.Ysize / 2 + y * this.Ysize / 2.5), (this.Xsize / 2) * dot_scale * engine.get_zoom(), 0, Math.PI * 2, true);

		ctx.fill();
		ctx.stroke();

		ctx.strokeStyle = old_stroke;
		ctx.fillStyle = old_fill;
	}
}

function guide_type(o) {
	o.draw_icon = function(ctx) {
		var old_fill = ctx.fillStyle;

		if (o.Output > 0)
			ctx.fillStyle = "rgb(20, 20, 190)";
		else
			ctx.fillStyle = "rgb(190, 20, 20)";

		//ctx.fillRect (engine.get_x(this.Xpos), engine.get_y(this.Ypos), this.Xsize*zoom, this.Ysize*zoom);	

		bounding_rect(ctx, this);

		ctx.fillStyle = old_fill;
	}
}

function timebase_type(o) {
	o.draw_icon = function(ctx) {
		ctx.beginPath();

		ctx.arc(engine.get_x(this.Xpos + this.Xsize / 2),
			engine.get_y(this.Ypos + this.Ysize / 2), (this.Xsize / 2) * engine.get_zoom(), 0, Math.PI * 2, true);

		ctx.fill();
		ctx.stroke();

		ctx.closePath();
	}

}

function timerange_type(o) {
	o.draw_icon = function(ctx) {
		bounding_rect(ctx, this);
	}

	o.draw_properties = function(ctx, x, y) {
		//alert("draw");
		var on = this.get_property("on");
		var off = o.get_property("off");
		var name = o.get_property("name");

		var old_fill = ctx.fillStyle;

		ctx.fillStyle = "rgb(0,0,0)";
		ctx.font = "16pt Arial";

		var f_size = 12 * engine.get_zoom();

		ctx.font = format(f_size) + "pt Arial";

		ctx.fillText(name, engine.get_x(x + this.Xsize * 0.1), engine.get_y(y) - f_size / 2);

		ctx.fillText(on, engine.get_x(x + this.Xsize * 0.1), engine.get_y(y + this.Ysize / 2) - f_size / 2);

		ctx.fillText(off, engine.get_x(x + this.Xsize * 0.1), engine.get_y(y + this.Ysize / 2) + f_size / 2 * 1.5);

		ctx.fillStyle = old_fill;
	}
}

function timer_type(o) {
	o.draw_icon = function(ctx) {
		bounding_rect(ctx, this);
	}

	o.draw_properties = function(ctx, x, y) {
		//alert("draw");
		var on = this.get_property("on duration");
		var off = o.get_property("off duration");
		var name = o.get_property("name");

		var old_fill = ctx.fillStyle;

		ctx.fillStyle = "rgb(0,0,0)";
		ctx.font = "16pt Arial";

		var f_size = 12 * engine.get_zoom();

		ctx.font = format(f_size) + "pt Arial";

		ctx.fillText(name, engine.get_x(x + this.Xsize * 0.1), engine.get_y(y) - f_size / 2);

		ctx.fillText("Timer", engine.get_x(x + this.Xsize * 0.1), engine.get_y(y + this.Ysize / 2) - f_size);

		ctx.fillText("  ON: " + on, engine.get_x(x + this.Xsize * 0.1), engine.get_y(y + this.Ysize / 2) + f_size / 2 * 1.5);

		ctx.fillText("OFF: " + off, engine.get_x(x + this.Xsize * 0.1), engine.get_y(y + this.Ysize / 2) + f_size * 2);

		ctx.fillStyle = old_fill;
	}
}

function conversion_type(o) {
	o.draw_icon = function(ctx) {
		bounding_rect(ctx, this);
	}
	o.draw_properties = function(ctx, x, y) {
		//alert("draw");
		var a = this.get_property("a");
		var b = o.get_property("b");
		var c = o.get_property("c");
		var name = o.get_property("name");

		var old_fill = ctx.fillStyle;

		ctx.fillStyle = "rgb(0,0,0)";
		ctx.font = "16pt Arial";

		var f_size = 12 * engine.get_zoom();

		ctx.font = format(f_size) + "pt Arial";

		ctx.fillText(name, engine.get_x(x + this.Xsize * 0.1), engine.get_y(y) - f_size / 2);

		ctx.fillText("Conversion", engine.get_x(x + this.Xsize * 0.1), engine.get_y(y + this.Ysize / 2) - f_size);

		ctx.fillText(a + "x²" + (b < 0 ? '' : '+') + b + "x" + (c < 0 ? '' : '+') + c, engine.get_x(x + this.Xsize * 0.1), engine.get_y(y + this.Ysize / 2) + f_size / 2 * 1.5);

		ctx.fillText(format(this.Output), engine.get_x(x + this.Xsize * 0.1), engine.get_y(y + this.Ysize / 2) + f_size * 2);

		ctx.fillStyle = old_fill;
	}
}

function logger_type(o) {
	o.draw_icon = function(ctx) {
		bounding_rect(ctx, this);
	}
	//o.draw_properties	
}

function alert_type(o) {
	o.draw_icon = function(ctx) {
		bounding_rect(ctx, this);
	}
	//o.draw_properties	
}

function delay_type(o) {
	o.draw_icon = function(ctx) {
		bounding_rect(ctx, this);
	}
}