$( document ).ready(function() {
  $('.toclipboard').click(function() {
      $(this).tooltip('enable');
      $(this).tooltip('show');
      $(this).children('input').select();
      document.execCommand("copy");
      window.getSelection().removeAllRanges();
      $(this).tooltip('disable');
  })
});