/* 
	logic engine
	(C) 2013 Jason Hunt
	nulluser@gmail.com
*/

/*** 
	Graphics
***/

"use strict";

var handle_size = 10;
var guide_size = 10;


function format(n)  { return(Number(n).toFixed(2)); } // Float to formatted string
function bformat(n) { return(Number(n).toFixed(0)); } // Float to single digit string

function get_x(x)       { return x*zoom+x_ofs;   }  // Get screen x from world x
function get_y(y)       { return y*zoom+y_ofs;   }  // Get screen y from world y
function get_world_x(x) { return (x-x_ofs)/zoom; }  // Get world x from screen x
function get_world_y(y) { return (y-y_ofs)/zoom; }  // Get world y from screen y


function draw_wire(ctx, objects, i1, i2, p1, p2)
{
	var o1 = objects[i1];
	var o2 = objects[i2];

	var x1 = o1.Xpos + o1.Xsize/2;
	var y1 = o1.Ypos + o1.Ysize/2;

	var x2 = o2.Xpos + o1.Xsize/2;
	var y2 = o2.Ypos + o1.Ysize/2;

	if (show_guide == 0)
	{
		if (o1.Dir == dir_type.left)  x1 += handle_size/2;
		if (o1.Dir == dir_type.right) x1 -= handle_size/2;
		if (o1.Dir == dir_type.up)    y1 += handle_size/2;
		if (o1.Dir == dir_type.down)  y1 -= handle_size/2;

		if (o2.Dir == dir_type.left)  x2 += handle_size/2;
		if (o2.Dir == dir_type.right) x2 -= handle_size/2;
		if (o2.Dir == dir_type.up)    y2 += handle_size/2;
		if (o2.Dir == dir_type.down)  y2 -= handle_size/2;
	}

 	ctx.moveTo(Math.round(get_x(x1)), Math.round(get_y(y1)));
 	ctx.lineTo(Math.round(get_x(x2)), Math.round(get_y(y2)));
}

function draw_object(ctx, o, x_size, y_size)
{
	if (o.Type == "guide" && show_guide == 0) return;

	if (o.selected)
	{			
		var old_fill  = ctx.fillStyle;

		var border = 4;

		ctx.fillStyle = "rgb(255,0,0)";
		ctx.fillRect (Math.round(get_x(o.Xpos - border )), Math.round(get_y(o.Ypos - border)), Math.round((o.Xsize + 2*border)*zoom),Math.round((o.Ysize + 2*border)*zoom));

		ctx.fillStyle = old_fill;
	}

	o.draw_icon(ctx);
	
	o.draw_properties(ctx, o.Xpos, o.Ypos);
}

function draw_objects(ctx, objects, x_size, y_size)
{
	ctx.clearRect (0, 0, x_size, y_size);

	ctx.lineWidth = 1;
	ctx.strokeStyle = "rgb(0,0,0)";
	ctx.fillStyle = "rgb(255,255,255)";

	// Draw all wires
	ctx.beginPath();
	for (var i in objects)
	{
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
	End of Graphics
***/

