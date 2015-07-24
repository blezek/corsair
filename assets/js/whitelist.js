var siteApp = angular.module('whitelistApp', ['ui.bootstrap']);

siteApp.controller('WhitelistController', function($scope,$modal,$http,$interval,$anchorScroll) {
  $scope.wlLoadStatus = true;
  $scope.blLoadStatus = true;
  
  $scope.rows = [];
  $scope.blacklist = [];

  $scope.pageSize = 5;
  $scope.currentPage = 1;

  $scope.wlPageSize = 5;
  $scope.wlCurrentPage = 1;

  $scope.sort = {
    column: 'last_access',
    descending: true
  };

  $scope.getOrdering = function() {
    return $scope.sort.column;
  };
  $scope.setOrderBy = function(o) {
    if ( o === $scope.sort.column ) {
      $scope.sort.descending = !$scope.sort.descending;
    }
    $scope.sort.column = o;
  };
  $scope.sortingBy = function(o) {
    return $scope.sort.column == o;
  };
  
  $scope.loadRows = function () {
    // Our parameters
    var start = 0;
    console.log($scope.wlPageSize, $scope.wlCurrentPage)
    start = $scope.wlPageSize * ($scope.wlCurrentPage - 1 );
    var config = { params : {
      "start" : start,
      "page_size" : $scope.wlPageSize
    } };
    
    $http.get("rest/whitelist", config)
    .success(function(result) {
      console.log(result);
      $scope.rows = result.rows;
      $scope.wlNumberOfItems = result.total_rows;
      $scope.wlLoadStatus = true;
    })
    .error(function(data,status,headers,config) {
      console.log('failure');
      $scope.wlLoadStatus = false;
    })

    var start = 0;
    if ( $scope.currentPage ) {
      start = $scope.pageSize * ($scope.currentPage - 1 );
    }
    var config = { params : {
      "start" : start,
      "page_size" : $scope.pageSize,
      "sort_by" : $scope.sort.column,
      "descending" : $scope.sort.descending
    } };
    
    $http.get( "rest/blacklist", config )
    .success(function(result) {
      console.log(result);
      $scope.blacklist = result.rows;
      $scope.numberOfItems = result.total_rows;
      $scope.blLoadStatus = true;
    })
    .error(function(data,status,headers,config) {
      console.log('failure', status);
      $scope.blLoadStatus = false;
    })
  }
  // Reload every 5 seconds
  $interval($scope.loadRows, 5000);

  $scope.purgeCache = function() {
      var c = $http.delete("/rest/blacklist");
      c.success(function(result) {
        $scope.loadRows();
        toastr.info(result);
      })
      c.error(function(data,status,headers,config) {
        toastr.error("Could not remove cached item");
      });
  }
  
  $scope.addSite = function(site,id,edit) {

    var save = function() {
      var c;
      c = $http.post("/rest/whitelist", $scope.site)
      c.success(function() {
        $scope.loadRows();
        toastr.info ( "Added " + $scope.site.url + " to whitelist")
      }).error(function(data,status,headers,config) {
        toastr.error("Could not save site to server");
      });
    };

    if ( id ) {
      var c = $http.delete("/rest/blacklist/" + id);
      c.success(function(result) {
        $scope.loadRows();
      })
      c.error(function(data,status,headers,config) {
        toastr.error("Could not remove cached item");
      });
    }

    if ( site && !edit ) {
      $scope.site = {
        "url" : site
      };
      save();
      return;
    }
    $scope.site = {};
    if ( site && edit ) {
      $scope.site = {"url": site};
    }
    $modal.open({
      templateUrl: 'site.html',
      scope: $scope,
      controller: function($scope,$modalInstance) {
        console.log("Create Site Modal");
        $scope.save = function() { save(); $modalInstance.dismiss();};
        $scope.close = function() { $modalInstance.dismiss(); };
      }
    });

  }
  $scope.editSite = function(site) {
    $scope.createSite(site);
  };

  $scope.deleteSite = function(site) {
    $modal.open ({
      templateUrl: 'confirm.html',
      scope: $scope,
      controller: function($scope, $modalInstance) {
        $scope.title = "Remove " + site.url + " from whitelist?";
        $scope.ok = function() {
          console.log("Calling delete!");
          $http.delete("rest/whitelist/" + site.id).success(function() {
            console.log("delete success")
            $scope.loadRows();
            $modalInstance.dismiss();
            toastr.info ( "Removed " + $scope.site + " from whitelist")
          })
          .error(function(data,status,headers,config) {
            toastr.error("Could not remove " + site.url + " from the list...\n" + data);
          });
        };
        $scope.cancel = function() { $modalInstance.dismiss(); };
      }
    });
  };


  console.log("Loading")
  $scope.loadRows();

});
