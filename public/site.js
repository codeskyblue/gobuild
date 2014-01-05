/* site.js */

var wsChars = ['-', '\\', '|', '/'];
wsIndex = 0;
var $waiting = $("span.waiting");
var $console = $("span.console");

var wsId = setInterval(function(){
	$waiting.html(wsChars[wsIndex]);
	wsIndex = (wsIndex + 1) % 4;
	
	$console.html($console.html() + "good\n");
}, 200);

setTimeout(function(){
	clearInterval(wsId);
	$waiting.html("");
}, 2000);