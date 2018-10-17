package main

import (
	"log"
	"time"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Reservation contains a definition of a AWS Instance Reservation
type Reservation struct {
	ID           string    `m:"RI_ID"`
	InstanceType string    `m:"instance_type"`
	Platform     string    `m:"platform"`
	OfferClass   string    `m:"offer_class"`
	OfferType    string    `m:"offer_type"`
	Start        time.Time `m:"start"`
	End          time.Time `m:"duration"`
	Duration     int64     `m:"end"`
	TimeLeft     float64   `m:"left"`
	Count        float64
	Active       bool
}

var (
	reservationFields = []string{
		"RI_ID",
		"instance_type",
		"platform",
		"offer_class",
		"offer_type",
		"start",
		"duration",
		"end",
		"left",
	}
	activeReservations = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "ec2",
		Subsystem: "reserved_instances",
		Name:      "active_count",
		Help:      "Number of active reserved instances.",
	}, reservationFields)
	retiredReservations = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "ec2",
		Subsystem: "reserved_instances",
		Name:      "retired_count",
		Help:      "Number of retired reserved instances.",
	}, reservationFields)
)

func updateReserved() {
	log.Println("Update Reserved Instances data...")

	data, err := getReservedInstances()

	if err != nil {
		log.Fatalf("Could not get Reserved Instances: %v\n", err)
	}
	for _, r := range data {

		parsed, err := ToMap(r, "m")

		if err != nil {
			return
		}

		// fmt.Printf("%#v\n", teste)

		// labels := []string{
		// 	r.ID,
		// 	r.InstanceType,
		// 	r.Platform,
		// 	r.OfferClass,
		// 	r.OfferType,
		// 	fmt.Sprintf("%v", r.Start),
		// 	fmt.Sprintf("%d", r.Duration),
		// 	fmt.Sprintf("%v", r.End),
		// 	fmt.Sprintf("%.2f", r.TimeLeft),
		// }

		if r.Active {
			// activeReservations.WithLabelValues(labels...).Set(r.Count)
			activeReservations.With(parsed).Set(r.Count)
			continue
		}

		retiredReservations.With(parsed).Set(r.Count)
		// retiredReservations.WithLabelValues(labels...).Set(r.Count)
	}
}

func getReservedInstances() (output []Reservation, err error) {

	input := &ec2.DescribeReservedInstancesInput{}

	result, err := svc.DescribeReservedInstances(input)
	if err != nil {
		return
	}

	for i := 0; i < len(result.ReservedInstances); i++ {

		duration := *result.ReservedInstances[i].Duration
		startDate := *result.ReservedInstances[i].Start
		endDate := startDate.Add(time.Second * time.Duration(duration))
		left := endDate.Sub(startDate)

		output = append(output, Reservation{
			ID:           *result.ReservedInstances[i].ReservedInstancesId,
			InstanceType: *result.ReservedInstances[i].InstanceType,
			Platform:     *result.ReservedInstances[i].ProductDescription,
			OfferClass:   *result.ReservedInstances[i].OfferingClass,
			OfferType:    *result.ReservedInstances[i].OfferingType,
			Count:        float64(*result.ReservedInstances[i].InstanceCount),
			Start:        startDate,
			End:          endDate,
			Duration:     duration,
			TimeLeft:     left.Seconds(),
			Active:       *result.ReservedInstances[i].State == "active",
		})
	}
	return
}
