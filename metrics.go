package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	failedLogins = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "doorman_total_failed_logins",
			Help: "Total number of failed logins",
		},
	)
	loginRate = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "doorman_login_rate_seconds",
			Help:    "Total logins in a given time period",
			Buckets: []float64{0.1, 0.5, 1, 5},
		},
		[]string{"total_logins"},
	)
)
