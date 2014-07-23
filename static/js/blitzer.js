(function() {

  var app = angular.module('blitzer', []);
  app.controller('HistoryController', [ '$http', function($http){
    var history = this;
    this.events = [];
    $http.get(window.location.pathname.trimRight('/') + '/history/0').success(function(data) {
      history.events = data.result;
    });
  }]);

})();
