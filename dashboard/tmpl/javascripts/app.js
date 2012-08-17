function activateTab($tab) {
  var $activeTab = $tab.closest('dl').find('dd.active'),
      contentLocation = $tab.children('a').attr("href") + 'Tab';

  // Strip off the current url that IE adds
  contentLocation = contentLocation.replace(/^.+#/, '#');

  //Make Tab Active
  $activeTab.removeClass('active');
  $tab.addClass('active');

  //Show Tab Content
  $(contentLocation).closest('.tabs-content').children('li').removeClass('active').hide();
  $(contentLocation).css('display', 'block').addClass('active');
}

jQuery(document).ready(function ($) {

  /* Use this js doc for all application specific JS */

  /* TABS --------------------------------- */
  /* Remove if you don't need :) */


  $('dl.tabs dd a').live('click.fndtn', function (event) {
    activateTab($(this).parent('dd'));
  });

  if (window.location.hash) {
    activateTab($('a[href="' + window.location.hash + '"]').parent('dd'));
    $.foundation.customForms.appendCustomMarkup();
  }

  /* ALERT BOXES ------------ */
  $(".alert-box").delegate("a.close", "click", function(event) {
    event.preventDefault();
    $(this).closest(".alert-box").fadeOut(function(event){
      $(this).remove();
    });
  });

  /* PLACEHOLDER FOR FORMS ------------- */
  /* Remove this and jquery.placeholder.min.js if you don't need :) */
  $('input, textarea').placeholder();

  /* TOOLTIPS ------------ */
  $(this).tooltips();

  /* UNCOMMENT THE LINE YOU WANT BELOW IF YOU WANT IE6/7/8 SUPPORT AND ARE USING .block-grids */
  //  $('.block-grid.two-up>li:nth-child(2n+1)').css({clear: 'left'});
  //  $('.block-grid.three-up>li:nth-child(3n+1)').css({clear: 'left'});
  //  $('.block-grid.four-up>li:nth-child(4n+1)').css({clear: 'left'});
  //  $('.block-grid.five-up>li:nth-child(5n+1)').css({clear: 'left'});


  /* DROPDOWN NAV ------------- */

  var lockNavBar = false;
  /* Windows Phone, sadly, does not register touch events :( */
  if (Modernizr.touch || navigator.userAgent.match(/Windows Phone/i)) {
    $('.nav-bar a.flyout-toggle').on('click.fndtn touchstart.fndtn', function(e) {
      e.preventDefault();
      var flyout = $(this).siblings('.flyout').first();
      if (lockNavBar === false) {
        $('.nav-bar .flyout').not(flyout).slideUp(500);
        flyout.slideToggle(500, function(){
          lockNavBar = false;
        });
      }
      lockNavBar = true;
    });
    $('.nav-bar>li.has-flyout').addClass('is-touch');
  } else {
    $('.nav-bar>li.has-flyout').hover(function() {
      $(this).children('.flyout').show();
    }, function() {
      $(this).children('.flyout').hide();
    });
  }

  /* DISABLED BUTTONS ------------- */
  /* Gives elements with a class of 'disabled' a return: false; */
  $('.button.disabled').on('click.fndtn', function (event) {
    event.preventDefault();
  });
  

  /* SPLIT BUTTONS/DROPDOWNS */
  $('.button.dropdown > ul').addClass('no-hover');

  $('.button.dropdown').on('click.fndtn touchstart.fndtn', function (e) {
    e.stopPropagation();
  });
  $('.button.dropdown.split span').on('click.fndtn touchstart.fndtn', function (e) {
    e.preventDefault();
    $('.button.dropdown').not($(this).parent()).children('ul').removeClass('show-dropdown');
    $(this).siblings('ul').toggleClass('show-dropdown');
  });
  $('.button.dropdown').not('.split').on('click.fndtn touchstart.fndtn', function (e) {
    $('.button.dropdown').not(this).children('ul').removeClass('show-dropdown');
    $(this).children('ul').toggleClass('show-dropdown');
  });
  $('body, html').on('click.fndtn touchstart.fndtn', function () {
    $('.button.dropdown ul').removeClass('show-dropdown');
  });

  // Positioning the Flyout List
  var normalButtonHeight  = $('.button.dropdown:not(.large):not(.small):not(.tiny)').outerHeight() - 1,
      largeButtonHeight   = $('.button.large.dropdown').outerHeight() - 1,
      smallButtonHeight   = $('.button.small.dropdown').outerHeight() - 1,
      tinyButtonHeight    = $('.button.tiny.dropdown').outerHeight() - 1;

  $('.button.dropdown:not(.large):not(.small):not(.tiny) > ul').css('top', normalButtonHeight);
  $('.button.dropdown.large > ul').css('top', largeButtonHeight);
  $('.button.dropdown.small > ul').css('top', smallButtonHeight);
  $('.button.dropdown.tiny > ul').css('top', tinyButtonHeight);
  
  $('.button.dropdown.up:not(.large):not(.small):not(.tiny) > ul').css('top', 'auto').css('bottom', normalButtonHeight - 2);
  $('.button.dropdown.up.large > ul').css('top', 'auto').css('bottom', largeButtonHeight - 2);
  $('.button.dropdown.up.small > ul').css('top', 'auto').css('bottom', smallButtonHeight - 2);
  $('.button.dropdown.up.tiny > ul').css('top', 'auto').css('bottom', tinyButtonHeight - 2);

  /* CUSTOM FORMS */
  $.foundation.customForms.appendCustomMarkup();


  /* LOGS */

  var log = $("#log");

  if(log.size() > 0) {
    (function(log){
      var highlight;
      var conn;
      var msg = $("#search-term");
      var highlightEnabled = false;

      function highlight(el, term)
      {
          var highlightedRegex = new RegExp('<span class="highlight">([^<]*)</span>', 'g');

          // Clear old highlighting
          var html = el.html();
          html = html.replace(highlightedRegex, '$1');

          // Ensure we are only replacing the content of tags and not tags themselves
          if(el.children().length > 0){
            var regex = new RegExp('(' + term + ')(?![a-zA-Z]*>)','gi');
            el.html(html.replace(regex,'<span class="highlight">$1</span>'));
          } else {
            // No inner elements so we can just replace the text
            var regex = new RegExp('(' + term + ')','gi');

            el.html(html.replace(regex,'<span class="highlight">$1</span>'));
          }
      }

      function appendLog(msg) {
        if (msg.text().indexOf(highlight) >= 0) {
          msg.css('background-color', "rgb(144, 238, 144)");
        }
        msg.appendTo(log);
      }

      $("#log-search").click(function() {
        highlightEnabled = false;

        if (!conn) {
          return false;
        }
        log.children().each(function(){
          $(this).remove()
        });
        var selected;
        $("#database-collection option:selected").each(function () {
            selected = $(this).text();
        });
        conn.send( JSON.stringify({ filter: msg.val(), collection: selected }));
        return false
      });

      $("#log-highlight").click(function() {
        highlightEnabled = true;
        highlight(log, msg.val());

        return false
      });

      if (window["WebSocket"]) {
        conn = new WebSocket("ws://" + document.location.host + "/logs/ws");
        conn.onclose = function(evt) {
          appendLog($("<div><b>Connection closed.</b></div>"))
        }
        conn.onmessage = function(evt) {
          logMsg = $("<div/>").text(evt.data);
          
          if(highlightEnabled){
            highlight(logMsg, msg.val());
          }

          appendLog(logMsg);
          log.children()[log.children().length-1].scrollIntoViewIfNeeded()
        }
      } else {
        appendLog($("<div><b>Your browser does not support WebSockets.</b></div>"))
      }
    })(log);

  }

});
