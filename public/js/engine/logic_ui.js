/* 
	logic engine
	(C) 2013 Jason Hunt
	nulluser@gmail.com
	With large modification from Whitham D. Reeve II
	thetawaves@gmail.com
*/
"use strict";

// Parameters
var snap = 5;
var min_zoom = 0.20;
var max_zoom = 2.0;

function snap_val(x) {
	return (Math.round((x) / snap) * snap);
}

var port_list = new Array();
var property_window = null;
var engine = EngineView();

$('.tool').tooltip();
$.noty.defaults["layout"] = 'bottom';
$.noty.defaults["type"] = 'error';
$('#property_and_canvas').layout({
	applyDefaultStyles: false,
	center__onresize: engine.resize,
});
$('#property_sidebar').on('shown.bs.collapse', function(event) {
	var inputs = $(event.target).find(':input');
	inputs.focus();
	inputs[0].selectionStart = inputs[0].value.length;
	inputs[0].selectionEnd = inputs[0].value.length;
});
$('#propertyTemplate').load('/public/property_window.html', function() {
	property_window = new Ractive({
		el: $('#property_sidebar'),
		template: $('#propertyTemplate').html(),
		data: {
			current_obj: false,
			tzdb: tzdb,
			current_timezone: jstz.determine().name(),
			port_list: port_list,
			netmgrTypes: function(t) {
				var types = new Array("binput", "ainput", "boutput", "aoutput");
				t = parseInt(t);
				return types[t];
			}
		}
	});
});
backend_start();
engine.start();