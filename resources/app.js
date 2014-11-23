angular.module( 'MonitorApp' , [ 'ngRoute' , 'config' , 'api' , 'highcharts-ng' ] )
	.config( [ '$routeProvider' , function( $routeProvider ){
		
		$routeProvider
			.when( '/Detail' , {
				templateUrl : './views/detail.html'
			} )
			.when( '/Monitor' , {
				templateUrl : './views/monitor.html'
			} )
			.otherwise( { redirectTo: '/Monitor' } );
			
	}] )