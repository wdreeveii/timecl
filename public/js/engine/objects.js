/* 
	logic engine
	(C) 2013 Jason Hunt
	nulluser@gmail.com
*/


/*
	Object 
*/

"use strict";

object_list.push("binput");

function binput_type (o)
{
	o.Xsize = 60;
	o.Ysize = 60;
	o.show_output = 1;

	o.input_termcount = 0;
	o.output_termcount = 1;

	if (o.PropertyCount == 0)
	{
		o.add_property("name", "string", "");
		o.add_property("value", "float", "0");
		o.add_property("port", "port", "None");
	}

	o.save_properties = function()
	{
		backend_setproperties(o.Id, o.PropertyCount, o.PropertyNames, o.PropertyTypes, o.PropertyValues);
		
		for (var i = 0; i < o.PropertyCount; i++)
		{
			if (o.PropertyNames[i] == "value")
			{
				backend_setoutput(o.Id, o.PropertyValues[i]);
				break;
			}
		}
	}

	o.set_output = function(output)
	{
		o.Output = output;
		
		backend_setoutput(o.Id, o.Output);
		
		for (var i = 0; i < o.PropertyCount; i++)
		{
			if (o.PropertyNames[i] == "value")
			{
				o.PropertyValues[i] = o.Output;
				break;
			}
		}		
		
		backend_setproperties(o.Id, o.PropertyCount, o.PropertyNames, o.PropertyTypes, o.PropertyValues);
	}
					 
	o.draw_icon = function(ctx) 
	{
		ctx.beginPath();
		
		ctx.arc(get_x(this.Xpos + this.Xsize/2), 
				get_y(this.Ypos + this.Ysize/2), 
				(this.Xsize/2)*zoom, 0, Math.PI*2, true);
		
		ctx.fill();
		ctx.stroke();				
	}
}

object_list.push("ainput");

function ainput_type (o)
{
	o.Xsize = 60;
	o.Ysize = 60;
	o.show_output = 1;

	o.input_termcount = 0;
	o.output_termcount = 1;

	if (o.PropertyCount == 0)
	{
		o.add_property("name", "string", "");
		o.add_property("value", "float", "0");
		o.add_property("port", "port", "None");
		o.add_property("Auto scale - Max", "float", 5);
		o.add_property("Auto scale - Min", "float", 0);
	}

	o.save_properties = function()
	{
		backend_setproperties(o.Id, o.PropertyCount, o.PropertyNames, o.PropertyTypes, o.PropertyValues);
		
		for (var i = 0; i < o.PropertyCount; i++)
		{
			if (o.PropertyNames[i] == "value")
			{
				backend_setoutput(o.Id, o.PropertyValues[i]);
				break;
			}
		}
	}

	o.set_output = function(output)
	{
		o.Output = output;
		
		backend_setoutput(o.Id, o.Output);
		
		for (var i = 0; i < o.PropertyCount; i++)
		{
			if (o.PropertyNames[i] == "value")
			{
				o.PropertyValues[i] = o.Output;
				break;
			}
		}		
		
		backend_setproperties(o.Id, o.PropertyCount, o.PropertyNames, o.PropertyTypes, o.PropertyValues);
	}
					 
	o.draw_icon = function(ctx) 
	{
		ctx.beginPath();
		
		ctx.arc(get_x(this.Xpos + this.Xsize/2), 
				get_y(this.Ypos + this.Ysize/2), 
				(this.Xsize/2)*zoom, 0, Math.PI*2, true);
		
		ctx.fill();
		ctx.stroke();				
	}
}

object_list.push("boutput");

function boutput_type (o)
{
	o.Xsize = 80;
	o.Ysize = 30;
	o.show_output = 1;
	
	if (o.PropertyCount == 0)
	{
		o.add_property("name", "string", "");
		o.add_property("port", "port", "None");
	}					 
	
	o.input_termcount = 1;
	o.output_termcount = 0;
			
	o.draw_icon = function(ctx) 
	{
		bounding_rect(ctx, this);
	}	
}

object_list.push("aoutput");

function aoutput_type (o)
{
	o.Xsize = 80;
	o.Ysize = 30;
		
	o.show_analog = 1;
	
	o.input_termcount = 1;
	o.output_termcount = 0;
	
	if (o.PropertyCount == 0)
	{
		o.add_property("name", "string", "");
		o.add_property("port", "port", "None");
	}

	o.draw_icon = function(ctx) 
	{
		bounding_rect(ctx, this);
	}
}

object_list.push("notgate");

function notgate_type (o)
{
	o.show_output = 1;
	if (o.PropertyCount == 0)
	{
		o.add_property("name", "string", "");
	}					 
	o.input_termcount = 1;
	o.output_termcount = 1;
	
	o.draw_icon = function(ctx) 
	{
		ctx.beginPath();	
		ctx.moveTo(Math.round(get_x(this.Xpos)), Math.round(get_y(this.Ypos)));		
		ctx.lineTo(Math.round(get_x(this.Xpos)), Math.round(get_y(this.Ypos + this.Ysize)));
		ctx.lineTo(Math.round(get_x(this.Xpos + this.Xsize)), Math.round(get_y(this.Ypos + this.Ysize/2)));
		ctx.closePath();	
		
		ctx.fill();
		ctx.stroke();	
	}
}

object_list.push("andgate");

function andgate_type (o)
{
	o.show_output = 1;  
       	 	
	o.input_termcount = 2;
	o.output_termcount = 1;
       	 	       	 	
	if (o.PropertyCount == 0)
	{
		o.add_property("name", "string", "");
	}
	
	o.draw_icon = function(ctx) 
	{
		ctx.beginPath();	
		ctx.moveTo(get_x(this.Xpos), get_y(this.Ypos));
		ctx.lineTo(get_x(this.Xpos), get_y(this.Ypos + this.Ysize  ));		
		ctx.arc(get_x(this.Xpos + this.Xsize/2), get_y(this.Ypos + this.Ysize/2), this.Xsize/2*zoom, 0.5 * Math.PI, 1.5 * Math.PI, true);

		ctx.closePath();	

		ctx.fill();
		ctx.stroke();	
	}	
}

object_list.push("orgate");

function orgate_type (o)
{
	o.show_output = 1;  
        	
	if (o.PropertyCount == 0)
	{
		o.add_property("name", "string", "");
	}					 

	o.input_termcount = 2;
	o.output_termcount = 1;

	o.draw_icon = function(ctx) 
	{
		ctx.beginPath();	
		
		ctx.arc(get_x(this.Xpos - this.Xsize * 1.15), 
		        get_y(this.Ypos + this.Ysize / 2), 
		        this.Xsize*1.2*zoom, 
		        0.10 * Math.PI, 1.90 * Math.PI, true);
		
		ctx.arc(get_x(this.Xpos + this.Xsize * 0.05), 
		        get_y(this.Ypos + this.Ysize / 2 +  this.Ysize * 0.6 ), 
		        this.Xsize*1.1*zoom, 
		        1.45 * Math.PI, 1.80 * Math.PI, false);
		
		ctx.arc(get_x(this.Xpos + this.Xsize * 0.05), 
		        get_y(this.Ypos + this.Ysize / 2 -  this.Ysize * 0.6 ), 
		        this.Xsize*1.1*zoom, 
		        0.20 * Math.PI, 0.55 * Math.PI, false);

		ctx.closePath();	
				
		ctx.fill();
		ctx.stroke();
	}   	
}

object_list.push("xorgate");

function xorgate_type (o)
{
	o.show_output = 1;  
        	
	if (o.PropertyCount == 0)
	{
		o.add_property("name", "string", "");
	}					 
	
	o.input_termcount = 2;
	o.output_termcount = 1;
	
	o.draw_icon = function(ctx) 
	{
		ctx.beginPath();	

		ctx.arc(get_x(this.Xpos - this.Xsize * 1.15), 
		        get_y(this.Ypos + this.Ysize / 2), 
		        this.Xsize*1.2*zoom, 
		        0.10 * Math.PI, 1.90 * Math.PI, true);
		
		ctx.arc(get_x(this.Xpos + this.Xsize * 0.05), 
		        get_y(this.Ypos + this.Ysize / 2 +  this.Ysize * 0.6 ), 
		        this.Xsize*1.1*zoom, 
		        1.45 * Math.PI, 1.80 * Math.PI, false);
		
		ctx.arc(get_x(this.Xpos + this.Xsize * 0.05), 
		        get_y(this.Ypos + this.Ysize / 2 -  this.Ysize * 0.6 ), 
		        this.Xsize*1.1*zoom, 
		        0.20 * Math.PI, 0.55 * Math.PI, false);
		ctx.closePath();	
		ctx.fill();
		ctx.stroke();

		ctx.beginPath();
		
		ctx.arc(get_x(this.Xpos - this.Xsize * 1.15 + 5), 
		        get_y(this.Ypos + this.Ysize / 2), 
		        this.Xsize*1.2*zoom, 
		        0.10 * Math.PI, 1.90 * Math.PI, true);

		ctx.stroke();
	}  	
}

object_list.push("mult");

function mult_type (o)
{
	o.show_analog = 1;
	o.show_name = "Mult";		
    	
	if (o.PropertyCount == 0)
	{
		o.add_property("name", "string", "");
	}					 

	o.input_termcount = 2;
	o.output_termcount = 1;
	
	o.draw_icon = function(ctx) 
	{
		bounding_rect(ctx, this);
	}
}

object_list.push("div");

function div_type (o)
{
	o.show_analog = 1;
	o.show_name = "Div";		
    	
	if (o.PropertyCount == 0)
	{
		o.add_property("name", "string", "");
	}					 
	
	o.input_termcount = 2;
	o.output_termcount = 1;
	
	o.draw_icon = function(ctx) 
	{
		bounding_rect(ctx, this);
	}
}

object_list.push("add");

function add_type (o)
{
	o.show_analog = 1;
	o.show_name = "Add";		
    	
	if (o.PropertyCount == 0)
	{
		o.add_property("name", "string", "");
	}					 
   	
	o.input_termcount = 2;
	o.output_termcount = 1;
		
	o.draw_icon = function(ctx) 
	{
		bounding_rect(ctx, this);
	}  	
}

object_list.push("sub");


function sub_type (o)
{
	o.show_analog = 1;
	o.show_name = "Sub";		
    	
	if (o.PropertyCount == 0)
	{
		o.add_property("name", "string", "");
	}
	
	o.input_termcount = 2;
	o.output_termcount = 1;
	
	o.draw_icon = function(ctx) 
	{
		bounding_rect(ctx, this);
	}  	
}

object_list.push("power");

function power_type (o)
{
	o.show_analog = 1;
	o.show_name = "Power";		
    	
	if (o.PropertyCount == 0)
	{
		o.add_property("name", "string", "");
	}					 
    
	o.input_termcount = 2;
	o.output_termcount = 1;

	o.draw_icon = function(ctx) 
	{
		bounding_rect(ctx, this);
	}  	
}

object_list.push("sine");

function sine_type (o)
{
	o.show_analog = 1;
	o.show_name = "Sine";
	
	if (o.PropertyCount == 0)
	{
		o.add_property("name", "string", "");
	}					 
	
	o.input_termcount = 1;
	o.output_termcount = 1;
	
	o.draw_icon = function(ctx) 
	{
		bounding_rect(ctx, this);
	}  	
}

object_list.push("cosine");

function cosine_type (o)
{
	o.show_analog = 1;
	o.show_name = "Cosine";
   	
	if (o.PropertyCount == 0)
	{
		o.add_property("name", "string", "");
	}					 
    
	o.input_termcount = 1;
	o.output_termcount = 1;
	
	o.draw_icon = function(ctx) 
	{
		bounding_rect(ctx, this);
	} 	
}

object_list.push("agtb");

function agtb_type (o)
{
	o.show_analog = 1;
	o.show_name = "A > B";
	
	if (o.PropertyCount == 0)
	{
		o.add_property("name", "string", "");
	}					 
	
	o.input_termcount = 2;
	o.output_termcount = 1;
	
	o.draw_icon = function(ctx) 
	{
		bounding_rect(ctx, this);
	}  	
}

object_list.push("agteb");

function agteb_type (o)
{
	o.show_analog = 1;
	o.show_name = "A >= B";
	
	if (o.PropertyCount == 0)
	{
		o.add_property("name", "string", "");
	}					 
	
	o.input_termcount = 2;
	o.output_termcount = 1;
	
	o.draw_icon = function(ctx) 
	{
		bounding_rect(ctx, this);
	}  	
}

object_list.push("altb");

function altb_type (o)
{
	o.show_analog = 1;
	o.show_name = "A < B";
	
	if (o.PropertyCount == 0)
	{
		o.add_property("name", "string", "");
	}					 
	
	o.input_termcount = 2;
	o.output_termcount = 1;
	
	o.draw_icon = function(ctx) 
	{
		bounding_rect(ctx, this);
	}  	
}

object_list.push("alteb");

function alteb_type (o)
{
	o.show_analog = 1;
	o.show_name = "A <= B";
	
	if (o.PropertyCount == 0)
	{
		o.add_property("name", "string", "");
	}					 
	
	o.input_termcount = 2;
	o.output_termcount = 1;
	
	o.draw_icon = function(ctx) 
	{
		bounding_rect(ctx, this);
	}  	
}

object_list.push("aeqb");

function aeqb_type (o)
{
	o.show_analog = 1;
	o.show_name = "A == B";
	
	if (o.PropertyCount == 0)
	{
		o.add_property("name", "string", "");
	}					 
	
	o.input_termcount = 2;
	o.output_termcount = 1;
	
	o.draw_icon = function(ctx) 
	{
		bounding_rect(ctx, this);
	}  	
}

object_list.push("aneqb");

function aneqb_type (o)
{
	o.show_analog = 1;
	o.show_name = "A != B";
	
	if (o.PropertyCount == 0)
	{
		o.add_property("name", "string", "");
	}					 
	
	o.input_termcount = 2;
	o.output_termcount = 1;
	
	o.draw_icon = function(ctx) 
	{
		bounding_rect(ctx, this);
	}  	
}

object_list.push("xyscope");

function xyscope_type (o)
{
	o.Xsize = 200;
	o.Ysize = 200;
		
	if (o.PropertyCount == 0)
	{
		o.add_property("name", "string", "");
	} 	
	 
	o.input_termcount = 2;
	o.output_termcount = 0; 
	 
	o.draw_icon = function(ctx) 
	{ 
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

		ctx.arc(get_x(this.Xpos + this.Xsize/2  +  x * this.Xsize/2.5), 
				get_y(this.Ypos + this.Ysize/2  +  y * this.Ysize/2.5), 
				(this.Xsize/2)*dot_scale*zoom, 0, Math.PI*2, true);
		
		ctx.fill();
		ctx.stroke();
	
		ctx.strokeStyle = old_stroke; 
		ctx.fillStyle = old_fill;
	} 		
}

object_list.push("guide");

function guide_type (o)
{
	o.Xsize = guide_size;
	o.Ysize = guide_size;

	o.draw_icon = function(ctx) 
	{	
		var old_fill = ctx.fillStyle;
		
		if (o.Output > 0)
	    	ctx.fillStyle = "rgb(20, 20, 190)";
	    else
		    ctx.fillStyle = "rgb(190, 20, 20)";
	     	     
  		//ctx.fillRect (get_x(this.Xpos), get_y(this.Ypos), this.Xsize*zoom, this.Ysize*zoom);	
		
		bounding_rect(ctx, this);
		
		ctx.fillStyle = old_fill;
	}
}

object_list.push("block");

function block_type (o)
{
	o.Xsize = 30;
	o.Ysize = 30;

	o.draw_icon = function(ctx) 
	{
		var old_fill = ctx.fillStyle;

		ctx.fillStyle = "rgb(60, 60, 60)";

		bounding_rect(ctx, this);			
		
		ctx.fillStyle = old_fill;
	}
}

object_list.push("vbar");

function vbar_type (o)
{
	o.Xsize = 10;
	o.Ysize = 100;
	
	o.draw_icon = function(ctx) 
	{
		var old_fill = ctx.fillStyle;

		ctx.fillStyle = "rgb(60, 60, 60)";

		bounding_rect(ctx, this);			
		
		ctx.fillStyle = old_fill;
	}
}

object_list.push("hbar");

function hbar_type (o)
{
	o.Xsize = 100;
	o.Ysize = 10;
	
	o.draw_icon = function(ctx) 
	{
		var old_fill = ctx.fillStyle;

		ctx.fillStyle = "rgb(60, 60, 60)";

		bounding_rect(ctx, this);			
		
		ctx.fillStyle = old_fill;
	}
}

object_list.push("timebase");

function timebase_type (o)
{
	o.Xsize = 60;
	o.Ysize = 60;

	o.show_output = 1;
	
	o.input_termcount = 0;
	o.output_termcount = 1;
		
	if (o.PropertyCount == 0)
	{
		o.add_property("name", "string", "");
	}
	
	o.draw_icon = function(ctx) 
	{
		ctx.beginPath();
				
		ctx.arc(get_x(this.Xpos + this.Xsize/2), 
				get_y(this.Ypos + this.Ysize/2), 
				(this.Xsize/2)*zoom, 0, Math.PI*2, true);				
		
		ctx.fill();		
		ctx.stroke();
		
		ctx.closePath();			
	}

}

object_list.push("timerange");

function timerange_type (o)
{
	o.Xsize = 80;
	o.Ysize = 40;
	o.show_output = 1;

//	o.add_output_terminal(0);		
	
	o.input_termcount = 0;
	o.output_termcount = 1; 

	if (o.PropertyCount == 0)
	{
		o.add_property("name", "string", "");
		o.add_property("on", "time", "8:00");
		o.add_property("off", "time", "18:00");
		o.add_property("timezone", "timezone", "");
	}					 
	
	o.draw_icon = function(ctx) 
	{
		bounding_rect(ctx, this);					
	}
	
	o.draw_properties = function(ctx, x, y)
	{
		//alert("draw");
		var on = this.get_property("on");
		var off = o.get_property("off");
		var name = o.get_property("name");
	
		var old_fill  = ctx.fillStyle;
	
		ctx.fillStyle = "rgb(0,0,0)";
		ctx.font = "16pt Arial";

		var f_size = 12 * zoom;

		ctx.font = format(f_size) + "pt Arial";

	    ctx.fillText(name, get_x(x + this.Xsize * 0.1 ), get_y(y) - f_size  / 2);

	    ctx.fillText(on, get_x(x + this.Xsize * 0.1 ), get_y(y + this.Ysize/2) - f_size  / 2);
	
	    ctx.fillText(off, get_x(x + this.Xsize * 0.1 ), get_y(y + this.Ysize/2) + f_size  / 2 * 1.5);

		ctx.fillStyle = old_fill;
	}
}


object_list.push("timer");

function timer_type (o)
{
	o.Xsize = 80;
	o.Ysize = 60;
	o.show_output = 1;

//	o.add_output_terminal(0);		
	
	o.input_termcount = 0;
	o.output_termcount = 1; 

	if (o.PropertyCount == 0)
	{
		o.add_property("name", "string", "");
		o.add_property("on duration", "time", "2s");
		o.add_property("off duration", "time", "2s");
	}					 
	
	o.draw_icon = function(ctx) 
	{
		bounding_rect(ctx, this);					
	}
	
	o.draw_properties = function(ctx, x, y)
	{
		//alert("draw");
		var on = this.get_property("on duration");
		var off = o.get_property("off duration");
		var name = o.get_property("name");
	
		var old_fill  = ctx.fillStyle;
	
		ctx.fillStyle = "rgb(0,0,0)";
		ctx.font = "16pt Arial";

		var f_size = 12 * zoom;

		ctx.font = format(f_size) + "pt Arial";

	    ctx.fillText(name, get_x(x + this.Xsize * 0.1 ), get_y(y) - f_size  / 2);

	    ctx.fillText("Timer", get_x(x + this.Xsize * 0.1 ), get_y(y + this.Ysize/2) - f_size);
	    
	    ctx.fillText("  ON: " + on, get_x(x + this.Xsize * 0.1 ), get_y(y + this.Ysize/2) + f_size  / 2 * 1.5);
	
	    ctx.fillText("OFF: " + off, get_x(x + this.Xsize * 0.1 ), get_y(y + this.Ysize/2) + f_size  * 2);
	
		ctx.fillStyle = old_fill;
	}
}

object_list.push("conversion");

function conversion_type (o)
{
	o.Xsize = 140;
	o.Ysize = 60;

	o.show_analog = 1;
	o.show_name = "conversion";
	//o.show_output = 1;
	
	if (o.PropertyCount == 0)
	{
		o.add_property("name", "string", "");
		o.add_property("a", "float", 0);
		o.add_property("b", "float", 0);
		o.add_property("c", "float", 0);
	}					 
	
	o.input_termcount = 1;
	o.output_termcount = 1;
	
	o.draw_icon = function(ctx) 
	{
		bounding_rect(ctx, this);
	}
	o.draw_properties = function(ctx, x, y)
	{
		//alert("draw");
		var a = this.get_property("a");
		var b = o.get_property("b");
		var c = o.get_property("c");
		var name = o.get_property("name");
	
		var old_fill  = ctx.fillStyle;
	
		ctx.fillStyle = "rgb(0,0,0)";
		ctx.font = "16pt Arial";

		var f_size = 12 * zoom;

		ctx.font = format(f_size) + "pt Arial";

	    ctx.fillText(name, get_x(x + this.Xsize * 0.1 ), get_y(y) - f_size  / 2);

	    ctx.fillText("Conversion", get_x(x + this.Xsize * 0.1 ), get_y(y + this.Ysize/2) - f_size);
	    
	    ctx.fillText(a + "x²" + (b<0?'':'+') + b + "x" + (c<0?'':'+') + c, get_x(x + this.Xsize * 0.1 ), get_y(y + this.Ysize/2) + f_size  / 2 * 1.5);
	
	    ctx.fillText(format(this.Output), get_x(x + this.Xsize * 0.1 ), get_y(y + this.Ysize/2) + f_size  * 2);
	
		ctx.fillStyle = old_fill;
	}	
}

object_list.push("logger");

function logger_type (o)
{
	//o.show_output = 1;
	o.show_analog = 1;
	o.show_name = "Log"
	if (o.PropertyCount == 0)
	{
		o.add_property("name", "string", "");
		o.add_property("frequency", "float", 300);
	}					 
	
	o.input_termcount = 1;
	o.output_termcount = 0;
			
	o.draw_icon = function(ctx) 
	{
		bounding_rect(ctx, this);
	}
	//o.draw_properties	
}

object_list.push("delay");

function delay_type (o)
{
	o.show_analog = 1;
	o.show_name = "Delay"
	
	if (o.PropertyCount == 0)
	{
		o.add_property("name", "string", "");
		o.add_property("delay", "float", 0);
		o.add_property("min on", "float", 0);
	}					 
	
	o.input_termcount = 1;
	o.output_termcount = 1;
			
	o.draw_icon = function(ctx) 
	{
		bounding_rect(ctx, this);
	}
}







