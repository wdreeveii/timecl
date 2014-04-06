/*
	Tamed Web Sockets
	Whitham D. Reeve II
*/

"use strict";

var TameSocket = function(usrConfig) {
	if (!(usrConfig instanceof Object)) {
		console.log("TameSocket: missing required configuration parameters");
		return undefined;
	}
	var config = {
		target: null,
		sleepTimeout: 60000,
		reconnectDelay: 5000,
		msgMinInterval: 1000,
		msgProcessor: null,
	};
	$.extend(config, usrConfig);

	if (!(typeof config.target == "string")) {
		console.log("TameSocket: no target address provided");
		return undefined;
	}

	var socket;
	var socketMsgBuffer = new Array();

	var doMsgProcessing = function() {
		if (config.msgProcessor instanceof Function) {
			config.msgProcessor(socketMsgBuffer);
		}
	}
	var msgRateLimit = null;
	var flushMsgBuffer = function() {
		if (socketMsgBuffer.length > 0) {
			msgRateLimit = setTimeout(flushMsgBuffer, config.msgMinInterval);
			doMsgProcessing();
		} else {
			msgRateLimit = null;
		}
	}
	var bufferMsgs = function(event) {
		socketMsgBuffer.push(event.data);
		if (msgRateLimit == null) {
			msgRateLimit = setTimeout(flushMsgBuffer, config.msgMinInterval);
			doMsgProcessing();
		}
	}

	var resetTimeout = null;
	var setupSocket = function() {
		resetTimeout = null;
		socket = new WebSocket(config.target)
		socket.onmessage = bufferMsgs;
		socket.onerror = function(event) {
			console.log("TameSocket error: " + event);
		}
		socket.onclose = function(event) {
			resetSocket();
		}
	}
	var resetSocket = function() {
		if (resetTimeout == null) {
			socket.close();
			resetTimeout = setTimeout(setupSocket, config.reconnectDelay);
		}
	}

	var sleepDetectStart = new Date();
	var sleepDetectTimeout = setInterval(function() {
		if (sleepDetectStart.getTime() + config.sleepTimeout < new Date().getTime()) {
			console.log("TameSocket detected sleep: reconnecting to websocket...");
			resetSocket();
		}
		sleepDetectStart = new Date();
	}, 1000);

	setupSocket();
	$(document).on('online', function(event) {
		console.log("TameSocket coming back online: reconnecting to websocket...");
		resetSocket()
	})
	return socket
}