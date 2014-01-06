/* site.js */

$(function(){
	var pgChars = ['-', '\\', '|', '/'];
	pgIndex = 0;
	var $waiting = $("span.waiting");
	var $console = $("span.console");
	var addr = $("#name").text();
	addr = $.trim(addr);
	
	function writeConsole(line) {
		$console.html($console.html() + line);
	}
	
	// websocket address
	var wsUri ="ws://localhost:3000/websocket"; 
	
	var ws =new WebSocket(wsUri);
	ws.onopen = function(e){
		console.log("open");
		console.log(e);
		var message = JSON.stringify({
			error: null,
			data: addr,
		})
		ws.send(message)
		console.log("send msg: hello")
	}
	ws.onmessage = function(e){
		console.log(e.data);
		var json = JSON.parse(e.data);
		console.log(json.data);
		writeConsole(json.data);
		return true;
	}
	ws.onclose = function(e) {
		console.log("close")
		clearInterval(pgId);
	}
	ws.onerror = function(e) {
		console.log("error")
	}
	
	var pgId = setInterval(function(){
		$waiting.html(pgChars[pgIndex]);
		pgIndex = (pgIndex + 1) % 4;
		
		//$console.html($console.html() + "good\n");
	}, 200);
	
	setTimeout(function(){
		clearInterval(pgId);
		$waiting.html("");
	}, 20000);
});
