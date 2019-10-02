// magic.js
$(document).ready(function() {

  $('form').submit(function(event) {

    $('.form-group').removeClass('has-error');
    $('.help-block').remove();
    $('.alert-success').remove();

    var formData = {
      'sku' : $('input[name=sku]').val()
    };

    $.ajax({
      type 		: 'POST',
      url 		: 'buyStuff',
      data 		: formData,
      dataType 	: 'json',
      encode 	: true
    })
    .success(function(data) {
      console.log(data);
      $('form').append('<div class="alert alert-success">' + data.message + '</div>');
    })

    .fail(function(data) {
      if(data.responseJSON.error) {
        $('#sku-group').addClass('has-error'); // add the error class to show red input
        $('#sku-group').append('<div class="help-block">' + data.responseJSON.error + '</div>'); // add the actual error message under our input
      }
    });

    event.preventDefault();
  });

});
