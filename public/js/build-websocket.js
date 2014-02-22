/* site.js */
var re = /\u001B\[([0-9]+;?)*[Km]/g;

var styles = new Array();
var formatLine = function(s) {
	// Check for newline and early exit?
	s = s.replace(/</g, "&lt;");
	s = s.replace(/>/g, "&gt;");

	var final = "";
	var current = 0;
	while (m = re.exec(s)) {	
			var part = s.substring(current, m.index+1);
			current = re.lastIndex;

			var token = s.substr(m.index, re.lastIndex - m.index);
			var code = token.substr(2, token.length-2);

			var pre = "";
			var post = "";

			switch (code) {
			case 'm': 
			case '0m':
				var len = styles.length;
				for (var i=0; i < len; i++) {
					styles.pop();
					post += "</span>"
				}
				break;
			case '30;42m': pre = '<span style="color:black;background:lime">'; break;
			case '36m':
			case '36;1m': pre = '<span style="color:cyan;">'; break;
			case '31m':
			case '31;31m': pre = '<span style="color:red;">'; break;
			case '33m': 
			case '33;33m': pre = '<span style="color:yellow;">'; break;
			case '32m':
			case '0;32m': pre = '<span style="color:lime;">'; break;
			case '90m': pre = '<span style="color:gray;">'; break;
			case 'K': 
			case '0K': 
			case '1K':
			case '2K': break;
			}		

			if (pre !== "") {
				styles.push(pre);
			}

			final += part + pre + post;
	}

	var part = s.substring(current, s.length);
	final += part;
	return final;
};

$(function () {
    var pgChars = ['-', '\\', '|', '/'],
        pgIndex = 0,
        $waiting = $("span.waiting"),
        $console = $("pre.console");

    var addr = $.trim($("#name").text());

    function writeConsole(line) {
        $console.html($console.html() + line);
    }


    var start_build = function () {
        var $project = $(":input[name=project]").val(),
            $branch = "master",
            $goos = "windows", // FIXME: read from html
            $goarch = "amd64",
            initMessage = {
                project: $project,
                branch: $branch,
                goos: $goos,
                goarch: $goarch
            };
        var ws = new WebSocket(wsUri);
        var message = JSON.stringify(initMessage);
		var pgId = 0;
        ws.onerror = function (e) { alert("websocket connect error"); };
        ws.onopen = function (e) {
            ws.send(message);
			// progress bar
			pgId = setInterval(function () {
				$waiting.html(pgChars[pgIndex]);
				pgIndex = (pgIndex + 1) % 4;
			}, 200);
        };
        ws.onmessage = function (e) {
            var json = JSON.parse(e.data);
            console.log(e.data);
            writeConsole(formatLine(json.data));
            //writeConsole(json.data); // leave it for debug
			window.scrollTo(0, document.body.scrollHeight)
            return true;
        };
        ws.onclose = function (e) {
            console.log("close");
            clearInterval(pgId);
            $waiting.html("--DONE--").css("color", "#a6e1ec");
        }
    };
	$(".start-build").click(function(){
		$console.html("");
		start_build(); // make a call
	});
});
