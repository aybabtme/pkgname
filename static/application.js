$(function() {
  var pkgnamefield = $('#pkgname');
  var validMessage = $('#validmessage');
  var invalidMessage = $('#invalidmessage');

  function hideMessages() {
    validMessage.hide();
    invalidMessage.hide();
    invalidMessage.find('ul').empty();
  }

  function showValidPkgname(name) {
    validMessage.find('.name').text(name);
    validMessage.show();
  }

  function showInvalidPkgname(name, causes) {
    var ul = invalidMessage.find('ul');

    causes.forEach(function(cause) {
      var li = document.createElement('li');
      li.innerText = cause;
      
      ul.append(li);
    });

    invalidMessage.find('.name').text(name);
    invalidMessage.show();
  }

  function afterValidate(data) {
    hideMessages();

    if (data.error != '') {
      console.log(data.error);
      return;
    }

    console.log(data);

    if (data.success === false) {
      showInvalidPkgname(data.pkgname, data.causes);
    } else {
      showValidPkgname(data.pkgname);
    }
  }

  $('#pkgnameform').on('submit', function(e) {
    e.preventDefault();

    var name = pkgnamefield.val().trim();
    if (name == '') return;

    $.ajax({
      url: '/validate',
      type: 'post',
      data: { pkgname: name },
      success: afterValidate
    });
  });

  $('#generate').on('click', function(e) {
    e.preventDefault();

    $.get('/generate', afterValidate);
  });
})
