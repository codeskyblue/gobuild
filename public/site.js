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

	// progress bar	
	var pgId = setInterval(function(){
		$waiting.html(pgChars[pgIndex]);
		pgIndex = (pgIndex + 1) % 4;
	}, 200);

	var start_build = function(){
		var $project = "github.com/shxsun/fswatch";
		var $branch = "master";
		var $goos = "linux";
		var $goarch = "amd64";
		var initMessage = {
			project: $project,
			branch: $branch,
			goos: $goos,
			goarch: $goarch,
		};
		var ws =new WebSocket(wsUri);
		var message = JSON.stringify(initMessage)
		ws.onerror = function(e) { alert("websocket connect error"); }
		ws.onopen = function(e){ ws.send(message); }
		ws.onmessage = function(e){
			var json = JSON.parse(e.data);
			console.log(json.data);
			writeConsole(json.data);
			return true;
		}
		ws.onclose = function(e) {
			console.log("close")
			clearInterval(pgId);
			$waiting.html("--DONE--").css("color", "blue");
		}
	};
	start_build(); // make a call
});
