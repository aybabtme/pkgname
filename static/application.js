$(function() {
  var pkgnameField = $('#pkgname');
  var pkgnameForm = $('#pkgnameform');
  var validMessage = $('#validmessage');
  var invalidMessage = $('#invalidmessage');
  var historyMessage = $('#historymessage');

  function hideMessages() {
    validMessage.hide();
    invalidMessage.hide();
    invalidMessage.find('ul').empty();
    historyMessage.hide();
    historyMessage.find('ul').empty();
  }

  function showValidPkgname(name) {
    validMessage.find('.name')
      .attr('href', '/?pkgname=' + encodeURIComponent(name))
      .text(name);
    validMessage.show();
  }

  function showInvalidPkgname(name, causes) {
    var ul = invalidMessage.find('ul');

    causes.forEach(function(cause) {
      var li = document.createElement('li');
      $(li).text(cause);
      ul.append(li);
    });

    invalidMessage.find('.name')
      .attr('href', '/?pkgname=' + encodeURIComponent(name))
      .text(name);
    invalidMessage.show();
  }

  function afterValidate(data) {
    hideMessages();

    if (data.error != '') {
      console.log(data.error);
      return;
    }

    if (data.success === false) {
      showInvalidPkgname(data.pkgname, data.causes);
    } else {
      showValidPkgname(data.pkgname);
    }
  }

  function afterHistory(data) {
    hideMessages();

    var ul = historyMessage.find('ul');
    var badsUl = ul.first();
    var goodsUl = ul.last();

    data.bads.forEach(function(bad) {
      var li = document.createElement('li');
      $(li).text(bad);
      badsUl.append(li);
    });

    data.goods.forEach(function(good) {
      var li = document.createElement('li');
      $(li).text(good);
      goodsUl.append(li);
    });

    historyMessage.show();
  }

  pkgnameForm.on('submit', function(e) {
    e.preventDefault();

    var name = pkgnameField.val().trim();
    if (name == '') return;

    $.ajax({
      url: '/validate',
      type: 'post',
      data: { pkgname: name },
      success: afterValidate
    });
  });

  $('#example').on('click', function(e) {
    e.preventDefault();
    $.get('/generate', afterValidate);
  });

  $('#history').on('click', function(e) {
    e.preventDefault();
    $.get('/history', afterHistory);
  });

  if (window.location.search.length) {
    var pkgParam = null;
    var queryString = window.location.search.substr(1);
    queryString.split('&').forEach(function(p) {
      var pair = p.split('=');

      if (pair[0].toLowerCase() == 'pkgname') {
        pkgParam = pair[1];
      }
    });

    if (pkgParam) {
      $('.main').hide();
      $('#tryagainmessage').css('display', 'block');
      pkgnameField.val(pkgParam);
      pkgnameForm.trigger('submit');
    }
  }
})
