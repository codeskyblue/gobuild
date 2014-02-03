/* home */
$(function(){
	$("#address").on('input',(function(){
		$("#form").attr("action", "./download/"+$(this).val());
	}));
});

