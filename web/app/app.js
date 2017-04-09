'use strict';

// Declare app level module which depends on views, and components
angular.module('myApp', [
    'ngRoute',
    'ui-notification',
    'zingchart-angularjs',
    'myApp.home',
    'myApp.wallets',
    'myApp.settings',
    'myApp.version',
    'myApp.buy',
    'myApp.sell',
    'myApp.enabledExchanges',
    'myApp.buyOrders',
    'myApp.sellOrders',
    'myApp.stringUtils',
    'myApp.webSocket',
    'myApp.charts.market-depth'
]).
config(['$locationProvider', '$routeProvider', 'NotificationProvider', function($locationProvider, $routeProvider, NotificationProvider) {
    NotificationProvider.setOptions({
        delay: 5000,
        startTop: 60,
        startRight: 10,
        verticalSpacing: 10,
        horizontalSpacing: 20,
        positionX: 'right',
        positionY: 'top'
    });

    $locationProvider.hashPrefix('!');

    $routeProvider.otherwise({ redirectTo: '/' });
}]);