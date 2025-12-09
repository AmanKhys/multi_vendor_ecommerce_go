package chartGen

import (
	"bytes"
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/jung-kurt/gofpdf"
	"github.com/wcharczuk/go-chart/v2"
	"github.com/wcharczuk/go-chart/v2/drawing"
)

type PaymentStat struct {
	Count  int
	Amount float64
}

func GenerateChartsForSeller(orderStatusCounts map[string]int, paymentStatusCounts map[string]PaymentStat, platformFees, netProfit float64) (string, string, error) {
	pieChart := chart.PieChart{
		Width:  512,
		Height: 512,
		Values: []chart.Value{},
	}
	for status, count := range orderStatusCounts {
		if count > 0 {
			pieChart.Values = append(pieChart.Values, chart.Value{
				Value: float64(count),
				Label: status,
			})
		}
	}

	pieChartPath := fmt.Sprintf("/tmp/pie_%s.png", uuid.New().String())
	pieFile, err := os.Create(pieChartPath)
	if err != nil {
		return "", "", err
	}
	defer pieFile.Close()
	if err := pieChart.Render(chart.PNG, pieFile); err != nil {
		return "", "", err
	}

	barChart := chart.BarChart{
		Width:    512,
		Height:   512,
		BarWidth: 60,
		Bars:     []chart.Value{},
	}
	colors := map[string]drawing.Color{
		"Pending":   chart.ColorBlue,
		"Waiting":   chart.ColorYellow,
		"Failed":    chart.ColorRed,
		"Received":  chart.ColorGreen,
		"Cancelled": chart.ColorLightGray,
	}

	for status, stat := range paymentStatusCounts {
		if stat.Amount > 0 {
			barChart.Bars = append(barChart.Bars, chart.Value{
				Value: stat.Amount,
				Label: fmt.Sprintf("%s ($%.2f)", status, stat.Amount),
				Style: chart.Style{
					FillColor:   colors[status],
					StrokeColor: chart.ColorBlack,
				},
			})
		}
	}

	barChart.Bars = append(barChart.Bars,
		chart.Value{
			Value: netProfit,
			Label: fmt.Sprintf("Profit ($%.2f)", netProfit),
			Style: chart.Style{FillColor: chart.ColorGreen, StrokeColor: chart.ColorBlack},
		},
		chart.Value{
			Value: platformFees,
			Label: fmt.Sprintf("Fees ($%.2f)", platformFees),
			Style: chart.Style{FillColor: chart.ColorRed, StrokeColor: chart.ColorBlack},
		},
	)

	barChartPath := fmt.Sprintf("/tmp/bar_%s.png", uuid.New().String())
	barFile, err := os.Create(barChartPath)
	if err != nil {
		return "", "", err
	}
	defer barFile.Close()
	if err := barChart.Render(chart.PNG, barFile); err != nil {
		return "", "", err
	}

	return pieChartPath, barChartPath, nil
}

func GenerateOrderStatusPieChartForAdmin(pending, processing, shipped, delivered, cancelled, returned int) ([]byte, error) {
	values := []chart.Value{
		{Value: float64(pending), Label: "Pending"},
		{Value: float64(processing), Label: "Processing"},
		{Value: float64(shipped), Label: "Shipped"},
		{Value: float64(delivered), Label: "Delivered"},
		{Value: float64(cancelled), Label: "Cancelled"},
		{Value: float64(returned), Label: "Returned"},
	}

	pie := chart.PieChart{
		Width:  500,
		Height: 400,
		Values: values,
	}

	var buf bytes.Buffer
	err := pie.Render(chart.PNG, &buf)
	return buf.Bytes(), err
}

func GenerateProfitLossBarChartForAdmin(profit, loss float64) ([]byte, error) {
	graph := chart.BarChart{
		Title:      "Profit vs Loss",
		TitleStyle: chart.StyleTextDefaults(),
		Height:     400,
		BarWidth:   60,
		Bars: []chart.Value{
			{Value: profit, Label: "Profit"},
			{Value: loss, Label: "Loss"},
		},
	}

	var buf bytes.Buffer
	err := graph.Render(chart.PNG, &buf)
	return buf.Bytes(), err
}

func AddChartToPDFForAdmin(pdf *gofpdf.Fpdf, imgBytes []byte, name string, x, y, width float64) error {
	opt := gofpdf.ImageOptions{ImageType: "PNG", ReadDpi: false}
	pdf.RegisterImageOptionsReader(name, opt, bytes.NewReader(imgBytes))
	pdf.ImageOptions(name, x, y, width, 0, false, opt, 0, "")
	return nil
}
