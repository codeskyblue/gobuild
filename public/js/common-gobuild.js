/* command js function */
var getosarch = function(){
	var mapOS = {m:"darwin", l:"linux", w:"windows"};
	var mapArch = {darwin:"amd64", linux:"amd64", windows:"386"};
	var platform = navigator.platform;
	var goos = mapOS[platform.toLowerCase().substring(0, 1)] || "linux";
	var goarch = mapArch[goos];
	return {os: goos, arch: goarch};
}
