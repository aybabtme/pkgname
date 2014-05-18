$(function() {
  var pkgnamefield = $('#pkgname');

  function afterValidate(data) {
    console.log(data);
  }

  $('#pkgnameform').on('submit', function(e) {
    e.preventDefault();

    $.ajax({
      url: '/validate',
      type: 'post',
      data: { pkgname: pkgnamefield.val() },
      success: afterValidate
    });
  });

  $('#generate').on('click', function(e) {
    e.preventDefault();

    $.get('/generate', function(data) {
      console.log(data);
    });
  });
})
