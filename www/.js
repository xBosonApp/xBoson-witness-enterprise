jQuery(function($) {

$('.smartform').each(smartForm);

//
// 把 url 参数压入 input
//
var pp = location.search.substring(1).split("&");
pp.forEach(function(p) {
  var pn = p.split('=');
  $("input[name='"+ pn[0] +"']").val(pn[1]);
});


function smartForm() {
  var form = $(this);
  var msg = form.find("message");
  var act = form.attr("action");
  var successpage = form.attr("successpage");

  form.submit(function() {
    var parm = form.serialize();
    $.get(act, parm, function(ret) {
      msg.html(ret.msg);
      if (ret.code == 0 && successpage) {
        location.href = successpage;
      }
    }, 'json');
    return false;
  });
}

});