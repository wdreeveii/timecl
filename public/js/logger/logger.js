/* 
	data graphs
	(C) 2014 Whitham Reeve
	thetawaves@gmail.com
*/

"use strict";

var chartData = {};

$.ajax({
	url: "/logging/data",
	context: document.body,
	dataType: "json"
}).done(function(data) {
	chartData = data;
	$(function () {
		$(".loading-animation").hide();
		createStockChart();
	});
});


function createStockChart() {
	var chart = new AmCharts.AmStockChart();
	chart.pathToImages = "/public/js/amcharts/images/";
	chart.categoryAxesSettings.minPeriod = "ss";
	// DATASETS //////////////////////////////////////////
	// create data sets first
	var datasets = []
	Object.keys(chartData).forEach(function (x) {
		for (var i = 0; i < chartData[x].length; i++) {
			chartData[x][i].Timestamp = new Date(chartData[x][i].Timestamp * 1000);
		}
		var dset = new AmCharts.DataSet();
		dset.title = x;
		dset.fieldMappings = [{
			fromField: "Max",
			toField: "Max"
		}, {
			fromField: "Min",
			toField: "Min"
		}, {
			fromField: "Avg",
			toField: "Avg"
		}];
		dset.dataProvider = chartData[x];
		dset.categoryField = "Timestamp";
		datasets.push(dset);
	});

	// set data sets to the chart
	chart.dataSets = datasets;

	// PANELS ///////////////////////////////////////////
	// first stock panel
	var stockPanel1 = new AmCharts.StockPanel();
	stockPanel1.title = "Max";
	stockPanel1.percentHeight = 30;

	// graph of first stock panel
	var graph1 = new AmCharts.StockGraph();
	graph1.valueField = "Max";
	graph1.comparable = true;
	graph1.compareField = "Max";
	graph1.periodValue = "High"
	graph1.bullet = "round";
	graph1.bulletBorderColor = "#FFFFFF";
	graph1.bulletBorderAlpha = 1;
	graph1.balloonText = "[[title]]:<b>[[value]]</b>";
	graph1.compareGraphBalloonText = "[[title]]:<b>[[value]]</b>";
	graph1.compareGraphBullet = "round";
	graph1.compareGraphBulletBorderColor = "#FFFFFF";
	graph1.compareGraphBulletBorderAlpha = 1;
	stockPanel1.addStockGraph(graph1);

	// create stock legend
	var stockLegend1 = new AmCharts.StockLegend();
	stockPanel1.stockLegend = stockLegend1;


	// second stock panel
	var stockPanel2 = new AmCharts.StockPanel();
	stockPanel2.title = "Min";
	stockPanel2.percentHeight = 30;

	var graph2 = new AmCharts.StockGraph();
	graph2.valueField = "Min";
	graph2.comparable = true;
	graph2.compareField = "Min";
	graph2.periodValue = "Low"
	graph2.bullet = "round";
	graph2.bulletBorderColor = "#FFFFFF";
	graph2.bulletBorderAlpha = 1;
	graph2.balloonText = "[[title]]:<b>[[value]]</b>";
	graph2.compareGraphBalloonText = "[[title]]:<b>[[value]]</b>";
	graph2.compareGraphBullet = "round";
	graph2.compareGraphBulletBorderColor = "#FFFFFF";
	graph2.compareGraphBulletBorderAlpha = 1;
	stockPanel2.addStockGraph(graph2);

	var stockLegend2 = new AmCharts.StockLegend();
	stockPanel2.stockLegend = stockLegend2;

	var stockPanel3 = new AmCharts.StockPanel();
	stockPanel3.title = "Avg";
	stockPanel3.percentHeight = 30;

	// graph of first stock panel
	var graph3 = new AmCharts.StockGraph();
	graph3.valueField = "Avg";
	graph3.comparable = true;
	graph3.compareField = "Avg";
	graph3.periodValue = "Average"
	graph3.bullet = "round";
	graph3.bulletBorderColor = "#FFFFFF";
	graph3.bulletBorderAlpha = 1;
	graph3.balloonText = "[[title]]:<b>[[value]]</b>";
	graph3.compareGraphBalloonText = "[[title]]:<b>[[value]]</b>";
	graph3.compareGraphBullet = "round";
	graph3.compareGraphBulletBorderColor = "#FFFFFF";
	graph3.compareGraphBulletBorderAlpha = 1;
	stockPanel3.addStockGraph(graph3);

	// create stock legend
	var stockLegend3 = new AmCharts.StockLegend();
	stockPanel3.stockLegend = stockLegend3;

	// set panels to the chart
	chart.panels = [stockPanel1, stockPanel2, stockPanel3];


	// OTHER SETTINGS ////////////////////////////////////
	var sbsettings = new AmCharts.ChartScrollbarSettings();
	sbsettings.graph = graph3;
	sbsettings.usePeriod = "10mm";
	chart.chartScrollbarSettings = sbsettings;

	// CURSOR
	var cursorSettings = new AmCharts.ChartCursorSettings();
	cursorSettings.valueBalloonsEnabled = true;
	chart.chartCursorSettings = cursorSettings;


	// PERIOD SELECTOR ///////////////////////////////////
	var periodSelector = new AmCharts.PeriodSelector();
	periodSelector.dateFormat = "YYYY-MM-DD JJ:NN:SS";
	periodSelector.position = "left";
	periodSelector.periods = [{
		period: "ss",
		count: 60,
		label: "1 Minute"
	}, {
		period: "mm",
		count: 5,
		label: "5 Minutes"
	}, {
		period: "mm",
		count: 60,
		label: "1 Hour"
	}, {
		period: "mm",
		count: 360,
		label: "6 Hours"
	}, {
		period: "mm",
		count: 720,
		label: "12 Hours"
	}, {
		selected: true,
		period: "mm",
		count: 1440,
		label: "1 Day"
	}, {
		period: "DD",
		count: 5,
		label: "5 Days"
	}, {
		period: "YYYY",
		count: 1,
		label: "1 year"
	}, {
		period: "YTD",
		label: "YTD"
	}, {
		period: "MAX",
		label: "MAX"
	}];
	chart.periodSelector = periodSelector;


	// DATA SET SELECTOR
	var dataSetSelector = new AmCharts.DataSetSelector();
	dataSetSelector.position = "left";
	chart.dataSetSelector = dataSetSelector;

	chart.write('chartdiv');
}