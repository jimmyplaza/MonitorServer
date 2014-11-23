angular.module( 'config' , [] )
	
	.factory( 'Setting' , function(){
		
		return{
			
				MONITOR_COUNT					:10 , 
				SITE_DATA						:'./assets/',
				CHART_WIDTH						:600,
				CHART_HEIGHT					:100,
				CHART_DETAIL_WIDTH				:800,
				CHART_DETAIL_HEIGHT				:600,
				CHART_TYPE_1					:'alive',
				CHART_TYPE_2					:'responseTime',
				DATA_COLLECTION_TYPE_1			:'dct1',
				DATA_COLLECTION_TYPE_2			:'dct2',
				GET_DATA_TIME					:60										//sec
			}
		
	}); 
