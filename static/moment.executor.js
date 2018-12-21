$( document ).ready(function() {
  var now = moment();
  $('time').each(function(i, e) {
    var time = moment($(e).attr('datetime'));
    $(e).html('<span>' + time.from(now) + '</span>');
  });
});
