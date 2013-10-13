/* 
	engine utils
	(C) 2013 Whitham Reeve II
	thetawaves@gmail.com
*/

"use strict";

function new_engine() {
	backend_new_engine();
}

/*function save_engine() {
	var dialogdiv = $('<div id="save_dialog" title="Save New Engine"></div>');
	dialogdiv.append("<p>Please provide a name for this engine instance.</p>");
	dialogdiv.append("<p><label for='id_name'>Name:</label><input class='form-control' type='text' id='id_name' name='name'/></p>");
	$('body').append(dialogdiv);
	dialogdiv.dialog({
		height: 220,
		width: 400,
		buttons: {
			Save: function() {
				var name = $( this ).find("input[name='name']").val();
				backend_save_engine(name);
				$( this ).dialog("close");
				$( this ).remove();
			},
			Cancel: function() {
				$( this ).dialog("close");
				$( this ).remove();
			}
		}
	});
}*/