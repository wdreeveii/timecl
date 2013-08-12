/* 
	logic engine
	(C) 2013 Jason Hunt
	nulluser@gmail.com
*/



"use strict";

function data_source_request(i, data_name)
{
/*	var xmlhttp;
	
	if (window.XMLHttpRequest)
  	{	// code for IE7+, Firefox, Chrome, Opera, Safari
  		xmlhttp=new XMLHttpRequest();
  	}
	else
  	{	// code for IE6, IE5
  		xmlhttp=new ActiveXObject("Microsoft.XMLHTTP");
  	}

	xmlhttp.onreadystatechange=function()
	{
		if (xmlhttp.readyState==4 && xmlhttp.status==200)
		{
			data_source_process(i, data_name, xmlhttp.responseText);
		}
	}
	xmlhttp.open("GET","get_value.php?name=" + data_name, true);
	xmlhttp.send();*/
}




function data_source_process(i, data_name, data)
{
	/*//debug(data);

	if (i < 0) return;
	if (i >= obj.length) return;

	
	if (obj[i].type != "httpsource") return;
	
	if (obj[i].source_name != data_name) return;
	
	obj[i].output = data;	*/
	
	
	
}


function data_source_update( )
{
/*	for (var i in obj)
	{
		if (obj[i].type == "httpsource")
			data_source_request(i, obj[i].source_name);
		
	}*/
}


function data_source_start( )
{
	//setInterval(function() { data_source_update(); }.bind(this), 100);
}

