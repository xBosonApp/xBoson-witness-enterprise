jQuery(function($) {

$('.smartform').each(smartForm);
var jlog = $("#log");

if (jlog.length > 0) {
  startSysLogger();
}

//
// 把 url 参数压入 input
//
var pp = location.search.substring(1).split("&");
pp.forEach(function(p) {
  var pn = p.split('=');
  $("input[name='"+ pn[0] +"']").val(pn[1]);
});


var chain_select = selectGroup("#chain_list ul");
list_content(chain_select.root, 'chain_list', null, function(li, a, v) {
  a.click(function() {
    listChannel(v)
    chain_select.select(a);
    return false;
  }).attr('href', '#');
});


function listChannel(chainid) {
  var channel_select = selectGroup("#channel_list ul");
  var p = {chain: chainid};
  list_content(channel_select.root, 'channel_list', p, function(li, a, v) {
    a.click(function() {
      listBlock(chainid, v);
      channel_select.select(a);
      return false;
    }).attr('href', '#');
  });
}


function listBlock(chainid, channelid) {
  var block_select = selectGroup("#block_list ul");
  var p = {chain: chainid, channel: channelid};
  var next_key;
  requestBlocks();

  $("#next_page").off('click').click(function() {
    p.key = next_key;
    if (p.key) requestBlocks();
    return false;
  });

  $("#current_page").off('click').click(function() {
    delete p.key;
    requestBlocks();
    return false;
  });

  $("#search_block").off('click').click(function() {
    p.key = $("#input_key").val();
    requestBlocks();
    return false;
  });

  function requestBlocks() {
    list_content(block_select.root, 'get_block', p, function(li, a, v) {
      a.html(v.key).attr('href', '#').click(function() {
        showBlock(v);
        block_select.select(a);
        return false;
      });
      next_key = v.previousKey;
    });
  }
}


function selectGroup(root) {
  root = $(root);
  return {
    select : select,
    root   : root,
  };
  
  function select(jdom) {
    root.find(".select").removeClass('select');
    jdom.addClass('select');
  }
}


function showBlock(block) {
  var root = $('#block_info ul').html("");
  root.closest("section").show();
  var mapping = {
    create : { tr: tr_date },
    sign   : { tr: tr_sign },
    data   : { tr: tr_format_json },
    type   : { tr: tr_type },
  };
  for (var n in block) {
    var name = (mapping[n] && mapping[n].name) || n;
    var val  = (mapping[n] && mapping[n].tr && mapping[n].tr(block[n])) || block[n];
    var li = $("<li class='flex'>").appendTo(root);
    li.append($("<div class='name'>").html(name));
    li.append($("<div class='value'>").html(val));
  }
}


function tr_date(v) {
  return new Date(v).toLocaleString();
}


function tr_sign(v) {
  var r = [];
  v.forEach(function(s) {
    r.push("<div class='sign_id'>ID:&nbsp;", s.id, "</div>");
    r.push("<div class='sign_ct'>", s.si, "</div>");
  });
  return r.join('');
}


function tr_format_json(v) {
  try {
    var j = JSON.parse(Base64.decode(v));
    var s = JSON.stringify(j, 0, 2); 
    return $("<b>").text(s).html();
  } catch(e) {
    console.log("base64", e);
    return v;
  }
}


function tr_type(v) {
  return {
    1: '创世区块',
    2: '数据块',
    3: '加密块',
    4: '链码块',
    5: '消息块',
  }[v];
}


function list_content(jroot, api, parm, _every_li, _over) {
  jroot = $(jroot);
  if (jroot.length <= 0) {
    console.debug("list_content zero item");
    return;
  }
  jroot.closest("section").show();

  call(api, parm, function(err, ret) {
    if (err) return (_over && _over(err));
    jroot.html("")

    ret.data.forEach(function(v) {
      var a = $("<a>").html(v);
      var li = $("<li>").append(a);
      _every_li && _every_li(li, a, v);
      jroot.append(li);
    });
    _over && _over(null, ret, jroot);
  });
}


function call(api, parm, over) {
  $.ajax('../service/'+ api, {
    data: parm,
    dataType: 'json',

    success: function(ret) {
      if (ret.code == 0) {
        over(null, ret)
      } else {
        log(ret.msg);
        over(new Error(ret.msg))
      }
    },

    error: function(req, txt, err) {
      over(err);
    }
  });
}


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


function log() {
  var item = $("<div class='log'>");
  item.append("<span class='date'>"+ new Date().toLocaleString() +"</span>");
  item.append("<span class='content'>"
    + Array.prototype.join.call(arguments, ",") +"</span>");
  jlog.prepend(item);
}


function startSysLogger() {
  var LOG_MAX_COUNT = 1000;
  var log_count = 0;
  __do();
  
  function __do() {
    call('read_log', null, function(err, ret) {
      if (ret && ret.data) {
        ret.data.forEach(function(v) {
          log(v.join(' '))
        });
        log_count += ret.data.length;
      }
      if (log_count > LOG_MAX_COUNT) {
        log_count = 0;
        jlog.find(".log:gt(200)").remove();
      }
      setTimeout(__do, 100);
    });
  }
}

});