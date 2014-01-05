/* site.js */

$(function(){
	var pgChars = ['-', '\\', '|', '/'];
	pgIndex = 0;
	var $waiting = $("span.waiting");
	var $console = $("span.console");
	
	var wsUri ="ws://localhost:3000/websocket"; 
	var ws =new WebSocket(wsUri);
	ws.onopen = function(e){
		console.log("open");
		console.log(e);
	}
	ws.onmessage = function(e){
		console.log(e.data);
	}
	ws.onclose = function(e) {
		console.log("close")
	}
	ws.onerror = function(e) {
		console.log("error")
	}
	
	var pgId = setInterval(function(){
		$waiting.html(pgChars[pgIndex]);
		pgIndex = (pgIndex + 1) % 4;
		
		$console.html($console.html() + "good\n");
	}, 200);
	
	setTimeout(function(){
		clearInterval(pgId);
		$waiting.html("");
	}, 2000);
});
