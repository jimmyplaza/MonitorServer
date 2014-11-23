
Date.prototype.Format = function (fmt) { //author: meizz 
var o = {
    "M+": this.getMonth() + 1, //月份 
    "d+": this.getDate(), //日 
    "h+": this.getHours(), //小时 
    "m+": this.getMinutes(), //分 
    "s+": this.getSeconds(), //秒 
    "q+": Math.floor((this.getMonth() + 3) / 3), //季度 
    "S": this.getMilliseconds() //毫秒 
};
if (/(y+)/.test(fmt)) fmt = fmt.replace(RegExp.$1, (this.getFullYear() + "").substr(4 - RegExp.$1.length));
for (var k in o)
if (new RegExp("(" + k + ")").test(fmt)) fmt = fmt.replace(RegExp.$1, (RegExp.$1.length == 1) ? (o[k]) : (("00" + o[k]).substr(("" + o[k]).length)));
return fmt;
}

angular.module( 'MonitorApp' )

	.controller( 'MonitorContainer' , [ '$scope' , 'Setting' , 'DataAPI' , function( $scope , Setting , DataAPI ){
		
		$scope.monitorDatas = [];
		$scope.hcConfig = [];

		var today = new Date().Format("yyyyMMdd");
		var Data_date = Setting.SITE_DATA + today + "_" + "allSite.json";
		//console.log(Data_date)

		
		var getData = function(){
			DataAPI.httpGet( Data_date, function(successData){
			   $scope.monitorDatas = DataAPI.dataCollector( DataAPI.dataParser(successData) , { dtype:Setting.DATA_COLLECTION_TYPE_2 });
			}, function(error){} );
		}
		getData();
		
		var _1MInterval = setInterval( getData , Setting.GET_DATA_TIME*1000 );
			
	}] );
		