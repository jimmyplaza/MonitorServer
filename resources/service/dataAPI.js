angular.module( 'api' , [] )

	.factory( 'DataAPI' , ['$http' , 'Setting' , function($http , Setting){
		
		var _sites = [
			{ sid : 1 ,url:'https://g2.nexusguard.com' , name:'g2'},
			{ sid : 2 ,url:'http://g2demo.nxg.me' , name:'g2demo'},
			{ sid : 3 ,url:'https://g2api.nexusguard.com' , name:'g2api' },
			{ sid : 4 ,url:'https://g2api2.nexusguard.com' , name:'g2api2' },
			{ sid : 5 ,url:'http://gc.nexusguard.com' , name:'gc'},
			{ sid : 6 ,url:'http://g2na1.nexusguard.com' , name:'g2na1'},
			{ sid : 7 ,url:'http://g2na2.nexusguard.com:8080' , name:'g2na2'},
			{ sid : 8 ,url:'http://g2ws.nexusguard.com' , name:'g2ws'},
			{ sid : 9 ,url:'https://www.nexusguard.com' , name:'www'},
			{ sid : 10 ,url:'https://portal.nexusguard.com' , name:'portal'}
		];
		
		var avgs = [];
		
		var getHCConfig = function( data , option){
			//console.log(  data );
			return {  "options":{"chart":{"type":"line"}},
					  "series":[{ 
						  "data":data,"id":"alive","type":"line","name":option.title,"dashStyle":"Solid","connectNulls":false,'color':option.color }],
					  "title":{"text":option.title},
					  "size": { "width": "500" },
					  "yAxis":{'title':""  },
					  "xAxis":{ 'type':'datetime' },
					  "useHighStocks":true
					}
		}
		
		var getMixHCConfig = function( status , resp , url ){
			//console.log(  data );
			
			return {  "options":{"chart":{"type":"line" }, "scrollbar":{"enabled":false} , "rangeSelector":{"enabled":false}},
					  "series":[
						{"data":status,"id":"status","type":"column","name":'Status',"dashStyle":"Solid","connectNulls":false,'color':'#FFFF80' },
						{"data":resp,"id":"responseTime","type":"line","name":'Response Time',"dashStyle":"Solid","connectNulls":false,'color':'#FF0000' }
					  ],
					  "title":{"text":''},
					  "yAxis":{'title':""  },
					  "xAxis":{ 'type':'datetime' },
					  "useHighStocks":true,
					  "size":{"width":"350" , 'height':'310'}
					}
		}
		
		var getAVG = function(  ){
			
		}
				
		var utcConvertor = function( dateStr ){
			var _utc ;	
			//2014-09-03 10:23:57.160641662 +0800 CST 
			var _c = dateStr.split( " " );
			var _date = _c[0].split("-");
			var _time = _c[1].split(":");
			return Date.UTC(_date[0] , parseInt(_date[1])-1 , _date[2] , _time[0] , _time[1] , _time[2]);
		}
		
		return {
			
				httpGet : function( url , success , error ){
					$http({method: 'GET', url: url }).
					success(success).
					error(error);	
				},
				
				dataParser: function ( data ){
					var _data = ("[" + data.slice( 0 , data.length-2 ) + "]");
					return JSON.parse( _data );
				},
				
				dataCollector : function( data , option ){
					var _d = [];
					for( var i=0 ;i<data.length;i++ ){
						for( var j=0;j<_sites.length;j++ ){
							if( data[i].Url == _sites[j].url ){
								if( !Array.isArray( _d[_sites[j].sid-1] ) ) _d[_sites[j].sid-1] = [];
								_d[_sites[j].sid-1].push( data[i] );
								if( _d[_sites[j].sid-1][_d[_sites[j].sid-1].length-1].ResponseTime.slice( _d[_sites[j].sid-1][_d[_sites[j].sid-1].length-1].ResponseTime.length-2 , _d[_sites[j].sid-1][_d[_sites[j].sid-1].length-1].ResponseTime.length ) == "ms" ){
									_d[_sites[j].sid-1][_d[_sites[j].sid-1].length-1].ResponseTime = parseFloat(_d[_sites[j].sid-1][_d[_sites[j].sid-1].length-1].ResponseTime)*0.001 + "s";
								}
								_d[_sites[j].sid-1][_d[_sites[j].sid-1].length-1].name = _sites[j].name;
								_d[_sites[j].sid-1][_d[_sites[j].sid-1].length-1].url = _sites[j].url;
								_d[_sites[j].sid-1][_d[_sites[j].sid-1].length-1].sid = _sites[j].sid;
							}
						}
					}
					
					//GET AVG
					/*for( var k=0;k<_d.length;k++ ){
						var _c = _d[i];
						var _avg = 0;
						for( var m=0;m<_c.length;m++ ){
							parseFloat(_c[m].Status)
						}
						//avgs
					}*/
					
					if( option.dtype == Setting.DATA_COLLECTION_TYPE_1 ){
						
						return _d;
						
					}else if ( option.dtype == Setting.DATA_COLLECTION_TYPE_2 ){
						var _data = [];
						for( var i=0;i<_d.length;i++ ){
							var _c = _d[i];
							console.log(_c)
							var _cd = { Status:[] , ResponseTime:[] , Name:'' , Url:'' , DataMix:[] , StatusAvg:0 , ResponseAvg:0 };
							_cd.Name = _c[0].name;
							_cd.Url = _c[0].url;
							for( var j=0;j<_c.length;j++ ){
								_cd.Status.push( [  utcConvertor(_c[j].LogTime) , parseFloat(_c[j].Status) ] );
								_cd.ResponseTime.push( [ utcConvertor(_c[j].LogTime) , parseFloat(_c[j].ResponseTime) ] );
							}
							_data.push( _cd );
							_data[i].DataMix = getMixHCConfig( _data[i].Status , _data[i].ResponseTime , _data[i].Url);
							_data[i].Status = getHCConfig( _data[i].Status , {title:'Status' , color:'#8080C0'  } );
							_data[i].ResponseTime = getHCConfig( _data[i].ResponseTime , {title:'Response Time' , color:'#0080FF'} );
						}
						return _data
					}
					
					return [];
				}
			
			}
		
		
	}])
	