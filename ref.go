package main
/*
		data := map[string]interface{}{}
		dec := json.NewDecoder(strings.NewReader(string(content)))
		dec.Decode(&data)
		jq := jsonq.NewQuery(data)
		liveThreatsChart, err := jq.ArrayOfArrays("liveThreatsChart", "Threats")
		NetflowBandwidth, err := jq.ArrayOfArrays("NetflowBandwidth")
		NetflowBandwidth = NetflowBandwidth[:len(NetflowBandwidth)-2]
		liveReqsChart, err := jq.ArrayOfArrays("liveReqsChart", "Reqs")
		CacheHit, err := jq.ArrayOfArrays("liveCacheChart", "CacheHit")
		Legitimated, err := jq.ArrayOfArrays("liveLegitimatedChart", "Legitimated")
		Upstream, err := jq.ArrayOfArrays("liveUpstreamChart", "Upstream")
		if err != nil {
			fmt.Println("jsonq Error: %s", err)
		}

		threats_min, threats_max, threats_avg := GetStatistic(liveThreatsChart)
		NetflowBandwidth_min, NetflowBandwidth_max, NetflowBandwidth_avg := GetStatistic(NetflowBandwidth)
		liveReqsChart_min, liveReqsChart_max, liveReqsChart_avg := GetStatistic(liveReqsChart)
		CacheHit_min, CacheHit_max, CacheHit_avg := GetStatistic(CacheHit)
		Legitimated_min, Legitimated_max, Legitimated_avg := GetStatistic(Legitimated)
		Upstream_min, Upstream_max, Upstream_avg := GetStatistic(Upstream)
	*/
	/*
		fmt.Println("threats_min: ", threats_min)
		fmt.Println("threats_max: ", threats_max)
		fmt.Println("threats_avg: ", threats_avg)

		fmt.Println("NetflowBandwidth_min: ", NetflowBandwidth_min)
		fmt.Println("NetflowBandwidth_max: ", NetflowBandwidth_max)
		fmt.Println("NetflowBandwidth_avg: ", NetflowBandwidth_avg)

		fmt.Println("liveReqsChart_min: ", liveReqsChart_min)
		fmt.Println("liveReqsChart_max: ", liveReqsChart_max)
		fmt.Println("liveReqsChart_avg: ", liveReqsChart_avg)

		fmt.Println("CacheHit_min: ", CacheHit_min)
		fmt.Println("CacheHit_max: ", CacheHit_max)
		fmt.Println("CacheHit_avg: ", CacheHit_avg)

		fmt.Println("Legitimated_min: ", Legitimated_min)
		fmt.Println("Legitimated_max: ", Legitimated_max)
		fmt.Println("Legitimated_avg: ", Legitimated_avg)

		fmt.Println("Upstream_min: ", Upstream_min)
		fmt.Println("Upstream_max: ", Upstream_max)
		fmt.Println("Upstream_avg: ", Upstream_avg)
	*/
	/*
		liveStatistic := "<br><br>LIVE REPORT: " +
			"<br>Threats min: " + humanize.Comma(int64(threats_min)) +
			"<br>Threats max: " + humanize.Comma(int64(threats_max)) +
			"<br>Threats avg: " + humanize.Comma(int64(threats_avg)) +
			"<br>Bandwidth_min: " + humanize.Comma(int64(NetflowBandwidth_min)) +
			"<br>Bandwidth_max: " + humanize.Comma(int64(NetflowBandwidth_max)) +
			"<br>Bandwidth_avg: " + humanize.Comma(int64(NetflowBandwidth_avg)) +
			"<br>Live Request min: " + humanize.Comma(int64(liveReqsChart_min)) +
			"<br>Live Request max: " + humanize.Comma(int64(liveReqsChart_max)) +
			"<br>Live Request avg: " + humanize.Comma(int64(liveReqsChart_avg)) +
			"<br>CachHit_min: " + humanize.Comma(int64(CacheHit_min)) +
			"<br>CachHit_max: " + humanize.Comma(int64(CacheHit_max)) +
			"<br>CachHit_avg: " + humanize.Comma(int64(CacheHit_avg)) +
			"<br>Legitimated_min: " + humanize.Comma(int64(Legitimated_min)) +
			"<br>Legitimated_max: " + humanize.Comma(int64(Legitimated_max)) +
			"<br>Legitimated_avg: " + humanize.Comma(int64(Legitimated_avg)) +
			"<br>Serve by origin min: " + humanize.Comma(int64(Upstream_min)) +
			"<br>Serve by origin max: " + humanize.Comma(int64(Upstream_max)) +
			"<br>Serve by origin avg: " + humanize.Comma(int64(Upstream_avg))
	*/
