function initHomeEvent(){
    var $ipt = $('#go-ipt'),
        $iptTip = $('#go-ipt-tip'),
        typeInTip = "Please type in search keywords or project address",
        prjTip = "Please type in project address in github.com";
    $('#go-search').on("click", function (e) {
        e.preventDefault();
        if (!$ipt.val()) {
            $iptTip.text(typeInTip).show(200);
            $ipt.focus();
            return;
        }
        window.location = "/search?key="+$ipt.val();
    });
    $("#go-download").on('click', (function (e) {
        e.preventDefault();
        if (!$ipt.val()) {
            $iptTip.text(typeInTip).show(200);
            $ipt.focus();
            return;
        }
        if($ipt.val().toString().indexOf("github.com/") == -1){
            $iptTip.text(prjTip).show(200);
            $ipt.focus();
            return;
        }
        window.location = "/download/"+$ipt.val();
    }));
}