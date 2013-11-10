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
	ctx.beginPath();

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


 	ctx.moveTo(get_x(x1), get_y(y1));
 	ctx.lineTo(get_x(x2), get_y(y2));

	ctx.stroke();
	ctx.closePath();
}




function draw_properties(ctx, o, x, y)
{
	var old_fill  = ctx.fillStyle;
	
	ctx.fillStyle = "rgb(0,0,0)";
	ctx.font = "16pt Arial";

	var f_size = 12 * zoom;

	ctx.font = format(f_size) + "pt Arial";

	if (o.show_output)
	    ctx.fillText(bformat(o.Output), get_x(x + o.Xsize * 0.3 ), get_y(y + o.Ysize/2) + f_size  / 2);

	if (o.show_analog)
	    ctx.fillText(format(o.Output), get_x(x + o.Xsize * 0.1 ), get_y(y + o.Ysize/2) + f_size  / 2);

	if (o.show_name)
	    ctx.fillText(o.Type, get_x(x + o.Xsize * 0.1 ), get_y(y + o.Ysize/2) - f_size  / 2 * 1.5);
	    
	ctx.fillStyle = old_fill;
	    
}


function draw_object(ctx, o, x_size, y_size)
{
	if (o.Type == "guide" && show_guide == 0) return;


	//if (o.Xpos + o.Xsize < Xsize) return;
	//if (o.Ypos + o.Xsize < Ysize) return;

//	if (o.Xpos  > Xsize) return;
	//if (o.Ypos  > Ysize) return;
	


	if (o.selected)
	{			
		var old_fill  = ctx.fillStyle;

		var border = 4;

		ctx.fillStyle = "rgb(255,0,0)";
		ctx.fillRect (get_x(o.Xpos - border ), get_y(o.Ypos - border), (o.Xsize + 2*border)*zoom,(o.Ysize + 2*border)*zoom);

		ctx.fillStyle = old_fill;
	}

	o.draw_icon(ctx);
	
	o.draw_properties(ctx, o.Xpos, o.Ypos);
	

	//draw_properties(ctx, o, o.Xpos, o.Ypos);
}


function draw_objects(ctx, objects, x_size, y_size)
{
	// Clear background
	ctx.fillStyle = "rgb(240,240,240)";

	ctx.fillStyle = "rgb(245,255,255)";

	ctx.fillRect (0, 0, x_size, y_size);


	ctx.lineWidth = 1;
	ctx.strokeStyle = "rgb(0,0,0)";
	ctx.fillStyle = "rgb(255,255,255)";


	// Draw all wires
	for (var i in objects)
	{
		var idx = objects[i].Source;
			
		if (idx >= 0)
			draw_wire(ctx, objects, i, idx, 1, 0);
	}


	ctx.lineWidth = 2;
	ctx.strokeStyle = "rgb(0,0,0)";
	ctx.fillStyle = "rgb(255,255,255)";


	for (var i in objects)
		draw_object(ctx, objects[i], x_size, y_size);
}


/*** 
	End of Graphics
***/

