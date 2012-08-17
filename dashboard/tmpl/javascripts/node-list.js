// Code is simplified by just asking model for HTML rather than also introducing Backbone Views, may change this later
jQuery(document).ready(function ($) {
    function sanitizeName(name){
      return name.replace(/[\s_.:]/g,"-").toLowerCase();
    }

    Mustache.tags = ["<%", "%>"];

    var Templates = {
      tab: $('#tab-template').html(),
      region: $('#region-template').html(),
      node: $('#node-template').html(),
      instance: $('#instance-template').html()
    };


    var Instance = Backbone.Model.extend({
      initialize: function(){
        this.on('change', function(i){
          $("#" + i.htmlId()).replaceWith(i.contentHTML());
        });
      },

      contentHTML: function(){
        return Mustache.to_html(Templates.instance, this);
      },

      htmlId: function(){
        return sanitizeName(this.get('address'));
      },

      averageResponseTime: function(){
        if(this.get('averageResponseTime') === 0){
          return "";
        }
      },

      lastRequest: function(){
        if(this.get('lastRequest') === 0){
          return "";
        }
      },

      uptime: function(){
        var startTime = this.get("startTime");
        return startTime == "" ? "" : $.timeago(startTime);
      }
    });

    var InstanceCollection = Backbone.Collection.extend({
      model: Instance,

      initialize: function(models, options){
        this.on('add', function(instance){
          $("#" + instance.get('node').regionHtmlId() + "-" + instance.get('node').htmlId() + " table tbody").append(instance.contentHTML());

          $("#" + instance.htmlId()).effect("highlight", {}, 600);
        });

        this.on('remove', function(i){
          $("#" + i.htmlId()).effect("highlight", {}, 600, function(){
            $("#" + i.htmlId()).remove();
          });
        });
      }
    });

    var Node = Backbone.Model.extend({
      initialize: function(models, options){
        this.set('instances', new InstanceCollection());
      },

      contentHTML: function(){
        var html = $(Mustache.to_html(Templates.node, this));
        var models = this.get('instances').models;

        for(var i in models){
          html.find('table tbody').append(models[i].contentHTML());
        }

        return $("<div />").append(html).html();
      },

      regionHtmlId: function(){
        return this.get('region').htmlId();
      },

      htmlId: function(){
        return sanitizeName(this.get('name'));
      },

      addInstance: function(instance, silent){
        this.get('instances').add({
          name: instance.Config.ServiceAddr.IPAddress + ":" + instance.Config.ServiceAddr.Port,
          id: instance.Config.ServiceAddr.IPAddress + ":" + instance.Config.ServiceAddr.Port,
          service: instance.Config.Name,
          version: instance.Config.Version,
          address: instance.Config.ServiceAddr.IPAddress + ":" + instance.Config.ServiceAddr.Port,
          adminAddress: instance.Config.ServiceAddr.IPAddress + ":" + instance.Config.ServiceAddr.Port,
          registered: instance.Registered,
          clients: instance.Clients,
          startTime: instance.StartTime,
          lastRequest: instance.LastRequest,
          averageResponseTime: instance.AverageResponseTime,
          node: this
        }, {silent: silent});
      },

      removeInstance: function(instance){
        var instances = this.get('instances');
        var i = instances.get(instance.Config.ServiceAddr.IPAddress + ":" + instance.Config.ServiceAddr.Port);

        if(i){
          instances.remove(i);
        }

        if(instances.size() === 0){
          this.get('region').get('nodes').remove(this);
        }
      },

      updateInstance: function(instance){
        var instances = this.get('instances');
        var i = instances.get(instance.Config.ServiceAddr.IPAddress + ":" + instance.Config.ServiceAddr.Port);

        if(i){
          i.set({
            name: instance.Config.ServiceAddr.IPAddress + ":" + instance.Config.ServiceAddr.Port,
            id: instance.Config.ServiceAddr.IPAddress + ":" + instance.Config.ServiceAddr.Port,
            service: instance.Config.Name,
            version: instance.Config.Version,
            address: instance.Config.ServiceAddr.IPAddress + ":" + instance.Config.ServiceAddr.Port,
            adminAddress: instance.Config.ServiceAddr.IPAddress + ":" + instance.Config.ServiceAddr.Port,
            registered: instance.Registered,
            clients: instance.Clients,
            startTime: instance.StartTime,
            lastRequest: instance.LastRequest,
            averageResponseTime: instance.AverageResponseTime,
            node: this,
          }, {silent: false});
        }
      }
    });

    var NodeCollection = Backbone.Collection.extend({
      model: Node,
      initialize: function(models, options){
        this.on('add', function(node){
          $("#region-" + node.get('region').htmlId() + "Tab").append(node.contentHTML());
          $("#" + node.get('region').htmlId() + "-" + node.htmlId()).effect("highlight", {}, 600);
        });

        this.on('remove', function(node){
          $("#" + node.get('region').htmlId() + "-" + node.htmlId()).effect("highlight", {}, 600, function(){
            $("#" + node.get('region').htmlId() + "-" + node.htmlId()).remove();
          });
        });
      }
    });

    var Region = Backbone.Model.extend({
      initialize: function(models, options) {
        this.set('nodes', new NodeCollection());
      },

      tabHTML: function(){
        return Mustache.to_html(Templates.tab, this);
      },

      contentHTML: function(){
        var html = $(Mustache.to_html(Templates.region, this));
        var contentHTML = "";
        var models = this.get('nodes').models;

        for(var i in models){
          contentHTML = contentHTML + models[i].contentHTML();
        }

        html.append(contentHTML);

        return $("<div />").append(html).html();
      },

      htmlId: function(){
        return sanitizeName(this.get('name'));
      },

      addInstance: function(instance, silent){
        var nodeName = instance.Config.ServiceAddr.IPAddress;
        var nodes = this.get('nodes');
        var node = nodes.get(nodeName);

        if(!node){
          nodes.add({id: nodeName, name: nodeName, region: this}, {silent: silent});
          node = nodes.get(nodeName);
        }

        node.addInstance(instance);
      },

      removeInstance: function(instance){
        var nodeName = instance.Config.ServiceAddr.IPAddress;
        var nodes = this.get('nodes');
        var node = nodes.get(nodeName);

        if(node){
          node.removeInstance(instance);

          if(nodes.size() === 0){
            regions.remove(this);
          }
        }
      },

      updateInstance: function(instance){
        var nodeName = instance.Config.ServiceAddr.IPAddress;
        var nodes = this.get('nodes');
        var node = nodes.get(nodeName);

        if(node){
          node.updateInstance(instance);
        }
      }
    })

    var RegionCollection = Backbone.Collection.extend({
        model: Region,

        initialize: function(models, options){
          this.on('add', function(region){
            $("#region-tabs").append(region.tabHTML());
            $("#instance-list").append(region.contentHTML());

            $("#region-" + region.htmlId() + "-tab").effect("highlight", {}, 600);
          });

          this.on('remove', function(region){
            if($("#region-" + region.htmlId() + "-tab").hasClass('active')){
                activateTab($('#region-tabs dd').first());
            }

            $("#region-" + region.htmlId() + "-tab").effect("highlight", {}, 600, function(){
              $("#region-" + region.htmlId() + "-tab").remove();
              $("#region-" + region.htmlId() + "Tab").remove();
            });
          });
        },

        render: function(){
          var tabHTML = "";
          var contentHTML = "";

          for(var i in this.models){
            tabHTML = tabHTML + this.models[i].tabHTML();

            contentHTML = contentHTML + this.models[i].contentHTML();
          }


          $("#region-tabs").append(tabHTML);
          $("#instance-list").append(contentHTML);
        }
    });

    var regions = new RegionCollection();

    function findOrCreateRegion(regionName, silent){
      var region = regions.get(regionName);

      if(!region){
        regions.add({name: regionName, id: regionName}, {silent: silent});
        region = regions.get(regionName);
      }

      return region;
    }

    function parseNotification(notification){
      if(notification.Action === "List"){
          regions.reset();

          for(var path in notification.Data){
            var region = findOrCreateRegion(notification.Data[path].Service.Config.Region, true);

            region.addInstance(notification.Data[path].Service, true);
          }

          // Hide loading animation and activate first tab
          $("#loading").hide();
          regions.render();
          $("#region-tabs").show();
          $("#instance-filter").show();
          $(".timeago").timeago();
          $("#instance-list").show();
          activateTab($('#region-tabs dd').first());

      } else if(notification.Action === "Update"){

        for(var path in notification.Data){
          var update = notification.Data[path];

          switch(update.Type) {
            case "InstanceAddNotification":
                var region = findOrCreateRegion(update.Service.Config.Region, false);
                region.addInstance(notification.Data[path].Service, false);
              break;

            case "InstanceUpdateNotification":
              var region = regions.get(update.Service.Config.Region);

              if(region){
                region.updateInstance(update.Service);
              }
              break;

            case "InstanceRemoveNotification":
              var region = regions.get(update.Service.Config.Region);

              if(region){
                region.removeInstance(update.Service);
              }
              break;
          }
        }
      }
    }


    // Monitor instances through WebSocket
    function openWebSocket(retryCount){
      var ticker;
      var conn = new WebSocket("ws://" + document.location.host + "/instances/ws");

      conn.onopen = function(evt){
          // Keep connection alive
          ticker = window.setInterval(function(){
            conn.send('{"Action": "Heartbeat"}');  
          }, 5000);
      };

      conn.onclose = function(evt) {
          if(ticker){
            window.clearInterval(ticker);
            ticker = null;
          }

          // we need to recreate/connect to the websocket so this page stays live
          if(retryCount < 5){
            openWebSocket(retryCount+1);
          }
      };

      conn.onmessage = function(evt) {
          var e =  $.parseJSON(evt.data);
          parseNotification(e);
      }


      return conn;
    }


    $.timeago.settings.strings.suffixAgo = "";
    // Set this to 0 because it sets a timer for each element, so this could grow a lot as elements are added/removed.
    // lets do it a smarter way
    $.timeago.settings.refreshMillis = 0;

    var timeagoInterval = setInterval(function(){
      $(".timeago").timeago();
    }, 6000);

    if(window["WebSocket"]){
      var conn = openWebSocket(0);

      $(".filter-button").live('click', function(evt){
        // Clear current list and let the socket know aboutt our new filter
        regions.reset();
        $("#region-tabs").empty();
        $("#instance-list").empty().hide();
        $("#instance-filter").hide();
        $("#loading").show();


        // Set filter
        $("#instance-filter dd").removeClass('active');
        $("#" + evt.target.id).parent('dd').addClass('active');

        switch(evt.target.id){
          case "filter-active":
            conn.send('{"Action": "Filter", "Data": {"Registered": true}}');  
            break;
          case "filter-inactive":
            conn.send('{"Action": "Filter", "Data": {"Registered": false}}');  
            break;

          case "filter-all":
            conn.send('{"Action": "Filter", "Data": {"Reset": true}}');  
            break;
        }

      });

    } else {
      $("#loading").html("Your browser does not support WebSockets")
    }

});
