function getParameterByName(name) {
    var match = RegExp('[?&]' + name + '=([^&]*)').exec(window.location.search);
    return match && decodeURIComponent(match[1].replace(/\+/g, ' '));
}

siteApp.controller('ForbiddenController', function($scope,$modal,$http,$interval) {

  // Find the parameters
  $scope.url = getParameterByName("url")
  $scope.destination = getParameterByName("destination");
  $scope.password = null;

  console.log("URL: ", $scope.url)
  console.log("Destination: ", $scope.destination);
  
  $scope.submit = function() {
    // Try it out
    toastr.info("Submitting: " + $scope.password);
    var item = {
      url: $scope.url,
      password: $scope.password
    };
      
    var c = $http.post("/rest/whitelist", item)
    c.success(function() {
      toastr.info ( "Added " + $scope.url + " to whitelist");
    }).error(function(data,status,headers,config) {
      toastr.error("Could not save site to server " + status + " \n" + data);
    });
  };
  
});

