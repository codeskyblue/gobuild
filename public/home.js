/* home */
$(function(){
	$("#address").on('input',(function(){
		$("#form").attr("action", "./build/"+$(this).val());
	}));
});

