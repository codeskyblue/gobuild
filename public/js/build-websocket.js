/* site.js */

$(function () {
    var pgChars = ['-', '\\', '|', '/'],
        pgIndex = 0,
        $waiting = $("span.waiting"),
        $console = $("pre.console");

    var addr = $.trim($("#name").text());

    function writeConsole(line) {
        $console.html($console.html() + line);
    }

    // progress bar
    var pgId = setInterval(function () {
        $waiting.html(pgChars[pgIndex]);
        pgIndex = (pgIndex + 1) % 4;
    }, 200);

    var start_build = function () {
        var $project = "github.com/shxsun/fswatch",
            $branch = "master",
            $goos = "linux",
            $goarch = "amd64",
            initMessage = {
                project: $project,
                branch: $branch,
                goos: $goos,
                goarch: $goarch
            };
        var ws = new WebSocket(wsUri);
        var message = JSON.stringify(initMessage);
        ws.onerror = function (e) {
            alert("websocket connect error");
        };
        ws.onopen = function (e) {
            ws.send(message);
        };
        ws.onmessage = function (e) {
            var json = JSON.parse(e.data);
            console.log(e.data);
            writeConsole(json.data);
            return true;
        };
        ws.onclose = function (e) {
            console.log("close");
            clearInterval(pgId);
            $waiting.html("--DONE--").css("color", "#a6e1ec");
        }
    };
    start_build(); // make a call
});
