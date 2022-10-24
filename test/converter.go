package main

import (
	"math"
	"strconv"
	"strings"

	"gopkg.in/inf.v0"
)

func ptrint32(p int32) *int32 {
	return &p
}

func convertDec(cpu string) inf.Dec {
	base := inf.NewDec(1, 0)
	if strings.HasSuffix(cpu, "m") {
		base = inf.NewDec(1000, 0)
		cpu = strings.TrimRight(cpu, "m")
	}
	convertedStrUint64, _ := strconv.ParseUint(cpu, 10, 64)
	num_inf := inf.NewDec(int64(convertedStrUint64), 0)
	return *new(inf.Dec).QuoExact(num_inf, base)
}

func Int64ToInt(i int64) int {
	if i < math.MinInt32 || i > math.MaxInt32 {
		return 0
	} else {
		return int(i)
	}
}
