package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"gitlab.com/neonmei/wesense_exporter/model"
	"gitlab.com/neonmei/wesense_exporter/o11y"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/metric/instrument"
)

var Config = model.LoadConfig("wesense_exporter")
var Tracer = otel.Tracer("wesense_exporter")
var Meter = global.MeterProvider().Meter("wesense_exporter")

// Web Portal has update interval of 1 min, so even if not the safest option if might be worth a shot
func sampleFromWebPortal(ctx context.Context, weSenseEndpoint string) (*model.WeSenseAirMetric, error) {
	_, span := Tracer.Start(ctx, "sampleFromWebPortal")
	defer span.End()

	response, err := http.Get(weSenseEndpoint)

	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, errors.New("Error connecting to endpoint")
	}

	defer response.Body.Close()

	if response.StatusCode != 200 {
		err := errors.New("Request returned not ok status")
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(response.Body)

	dioxideData := strings.TrimSpace(doc.Find("html body div div p").Get(1).FirstChild.Data)
	temperatureData := strings.TrimSpace(doc.Find("html body h1").Get(1).FirstChild.Data)
	humidityData := doc.Find("html body h1").Get(1).FirstChild.NextSibling.NextSibling.Data

	if !strings.Contains(humidityData, "|") {
		err := errors.New("Unexpected format at humidity data returned by html selector")
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	parsedTemp, parsedTempError := strconv.ParseFloat(temperatureData, 32)
	parsedHumidity, parsedHumidityError := strconv.ParseUint(strings.TrimSpace(strings.Split(humidityData, "|")[1]), 10, 8)
	parsedDioxide, parsedDioxideError := strconv.ParseUint(dioxideData, 10, 16)

	if parsedTempError != nil || parsedHumidityError != nil || parsedDioxideError != nil {
		err := errors.New("HTML Selector matched, but data was not parseable")
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	span.SetStatus(codes.Ok, "")

	return &model.WeSenseAirMetric{
		Temperature: float32(parsedTemp),
		Humidity:    uint8(parsedHumidity),
		Dioxide:     uint16(parsedDioxide),
	}, nil
}

// Kickstart Meters and setup callbacks that probe weSense status when OTel requires
func initializeObserver(ctx context.Context) {
	metricsTemp, _ := Meter.AsyncFloat64().Gauge(
		"wesense.air.temperature",
		instrument.WithUnit("1"),
		instrument.WithDescription("Temperature measured in celsius"),
	)

	metricsHumidity, _ := Meter.AsyncInt64().Gauge(
		"wesense.air.humidity",
		instrument.WithUnit("1"),
		instrument.WithDescription("Relative Humidity"),
	)

	metricsDioxide, _ := Meter.AsyncInt64().Gauge(
		"wesense.air.dioxide",
		instrument.WithUnit("1"),
		instrument.WithDescription("Carbon Dioxide in parts per million (ppm)"),
	)

	if err := Meter.RegisterCallback(
		[]instrument.Asynchronous{
			metricsTemp, metricsHumidity, metricsDioxide,
		},
		func(ctx context.Context) {
			// TODO: add option to fetch from CSV (interval ~5m)
			reading, err := sampleFromWebPortal(ctx, Config.Endpoint)

			if err == nil {
				metricsTemp.Observe(ctx, float64(reading.Temperature))
				metricsHumidity.Observe(ctx, int64(reading.Humidity))
				metricsDioxide.Observe(ctx, int64(reading.Dioxide))
			} else {
				fmt.Fprintf(os.Stderr, "Error getting data")
			}

		},
	); err != nil {
		panic(err)
	}
}

func main() {
	o11y.InitTracerOTLP(&Config)
	o11y.InitTracerMetrics(&Config)
	initializeObserver(context.Background())

	for {
		time.Sleep(time.Second * 30)
	}
}
