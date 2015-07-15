var siteApp = angular.module('whitelistApp', ['ui.bootstrap']);

siteApp.controller('WhitelistController', function($scope,$modal,$http) {
  $scope.rows = [];
  $scope.blacklist = [];

  $scope.loadRows = function () {
    // Our parameters
    
    $http.get("rest/whitelist")
    .success(function(result) {
      console.log(result);
      $scope.rows = result.rows;
    })
    .error(function(data,status,headers,config) {
      console.log('failure');
      toastr.error ( "Could not contact server" );
    })
    $http.get("rest/blocklist")
    .success(function(result) {
      console.log(result);
      $scope.blacklist = result.rows;
    })
    .error(function(data,status,headers,config) {
      console.log('failure');
      toastr.error ( "Could not contact server" );
    })
  }

  $scope.addSite = function() {
    $scope.site = {};
    $modal.open({
      templateUrl: 'site.html',
      scope: $scope,
      controller: function($scope,$modalInstance) {
        console.log("Create Site Modal");
        $scope.save = function() {
          var c;
            c = $http.post("/rest/whitelist", $scope.site)
          c.success(function() {
            $modalInstance.dismiss();
            $scope.loadRows();
          })
          .error(function(data,status,headers,config) {
            toastr.error("Could not save site to server");
          });
        };
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
