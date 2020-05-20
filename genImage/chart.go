package main

import (
	"bytes"
	"fmt"
	"math"

	"github.com/pkg/errors"
	"github.com/wcharczuk/go-chart"
	"github.com/wcharczuk/go-chart/drawing"
)

var barWidthConst = [5]int{500, 250, 188, 141, 113}

func createBarChart(games []gameInfo, userColors returnedColors) (*bytes.Buffer, error) {
	var bars []chart.Value
	var height float64
	var fontColor drawing.Color
	if len(games) == 0 {
		return new(bytes.Buffer), nil
	}
	height = games[0].Hours
	for i := range games {
		roundedHours := math.Round(games[i].Hours)
		R, G, B := games[i].Colors.Main.RGB()
		valueToAdd := chart.Value{
			Value: games[i].Hours,
			Label: fmt.Sprintf("%g", roundedHours),
			Style: chart.Style{
				FillColor:   drawing.Color{R: uint8(R), G: uint8(G), B: uint8(B), A: 255},
				StrokeColor: drawing.Color{R: uint8(R), G: uint8(G), B: uint8(B), A: 255},
			},
		}
		bars = append(bars, valueToAdd)
	}
	R, G, B := userColors.Secondary.RGB()
	fontColor = drawing.Color{R: uint8(R), G: uint8(G), B: uint8(B), A: 255}

	barChart := chart.BarChart{
		Height:   500,
		Width:    640,
		BarWidth: barWidthConst[len(bars)-1],
		Background: chart.Style{
			FillColor: drawing.Color{R: 1, G: 1, B: 1, A: 0},
		},
		XAxis: chart.Style{
			FontSize:  25,
			FontColor: fontColor,
			Show:      true,
		},
		YAxis: chart.YAxis{
			Range: &chart.ContinuousRange{
				Min: 0,
				Max: height,
			},
		},
		Canvas: chart.Style{
			FillColor: drawing.Color{R: 1, G: 1, B: 1, A: 0},
		},
		Bars: bars,
	}

	var graphImgByteArr []byte
	graphImg := bytes.NewBuffer(graphImgByteArr)
	err := barChart.Render(chart.PNG, graphImg)
	if err != nil {
		return nil, errors.Wrap(err, "Generating Bar Graph")
	}
	return graphImg, nil
}

func createPieChart(games []gameInfo) (*bytes.Buffer, error) {
	if len(games) == 0 {
		return new(bytes.Buffer), nil
	}

	var valuesToAdd []chart.Value

	for i := range games {
		R, G, B := games[i].Colors.Main.RGB()
		valueToAdd := chart.Value{
			Label: games[i].Name,
			Value: games[i].Hours,
			Style: chart.Style{
				FillColor: drawing.Color{R: uint8(R), G: uint8(G), B: uint8(B), A: 255},
				FontSize:  20,
			},
		}
		valuesToAdd = append(valuesToAdd, valueToAdd)
	}

	pie := chart.PieChart{
		Height: 600,
		Width:  600,
		Background: chart.Style{
			FillColor: drawing.Color{R: 1, G: 1, B: 1, A: 0},
		},
		Canvas: chart.Style{
			FillColor: drawing.Color{R: 1, G: 1, B: 1, A: 0},
		},
		Values: valuesToAdd,
	}

	pieChart := new(bytes.Buffer)
	err := pie.Render(chart.PNG, pieChart)
	if err != nil {
		return nil, errors.Wrap(err, "Rendering Pie Chart")
	}
	return pieChart, nil
}
