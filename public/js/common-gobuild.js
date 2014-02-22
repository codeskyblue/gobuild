/* command js function */
var getosarch = function(){
	var mapOS = {m:"darwin", l:"linux", w:"windows"};
	var mapArch = {darwin:"amd64", linux:"amd64", windows:"386"};
	var platform = navigator.platform;
	var goos = mapOS[platform.toLowerCase().substring(0, 1)] || "linux";
	var goarch = mapArch[goos];
	return {os: goos, arch: goarch};
}

function renderMarkdown() {
    var $md = $('.markdown');
    $md.each(function (i, item) {
        $(item).html(marked($(item).html().replace(/&gt;/g, '>')));
    });
    var code = $md.find('pre code');
    if (code.length) {
        $("<link>").attr({ rel: "stylesheet", type: "text/css", href: "/css/highlight.css"}).appendTo("head");
        $.getScript("/js/highlight.min.js", function () {
            code.each(function (i, item) {
                hljs.highlightBlock(item)
            });
        });
    }
}

$(function(){
    renderMarkdown();
});
