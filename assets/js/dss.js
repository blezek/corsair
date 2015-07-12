var siteApp = angular.module('dssApp', ['ui.bootstrap']);

siteApp.controller('SiteController', function($scope,$modal,$http) {
  $scope.sites = [];

  $scope.loadSites = function () {
    $http.get("rest/v1/sites")
    .success(function(result) {
      console.log(result);
      $scope.sites = result;
    })
    .error(function(data,status,headers,config) {
      console.log('failure');
      toastr.error ( "Could not contact server" );
    })
  }

  $scope.createSite = function(site) {
    $scope.site = {};
    if ( site ) {
      $scope.site = site;
    }
    $modal.open({
      templateUrl: 'site.html',
      scope: $scope,
      controller: function($scope,$modalInstance) {
        console.log("Create Site Modal");
        $scope.save = function() {
          var c;
          if ( $scope.site._id ) {
            // Put
            c = $http.put("/rest/v1/sites/" + $scope.site._id, $scope.site);
          } else {
            c = $http.post("/rest/v1/sites", $scope.site)
          }
          c.success(function() {
$scope.loadSites();
            $modalInstance.dismiss();
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
        $scope.title = "Delete site at " + site.directory + "?";
        $scope.message = "Delete site at: " + site.directory + " on port " + site.port + "?";
        $scope.ok = function() {
          console.log("Calling delete!");
          $http.delete("rest/v1/sites/" + site._id).success(function() {
            console.log("delete success")
$scope.loadSites();
            $modalInstance.dismiss();
          })
          .error(function(data,status,headers,config) {
            toastr.error("Could not delete site");
          });
        };
        $scope.cancel = function() { $modalInstance.dismiss(); };
      }
    });
  };


  $scope.loadSites();

});
