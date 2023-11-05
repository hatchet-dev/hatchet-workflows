package dispatcher

import (
	"crypto/rand"
	"fmt"
	"math/big"

	"github.com/hatchet-dev/hatchet-workflows/pkg/workflows/types"
)

func parseScheduleInput(input string) (parsedSchedule string, skipUpdateSchedule bool) {
	switch input {
	case string(types.Random15Min):
		return get15MinRandomSchedule(), true
	case string(types.RandomHourly):
		return getHourlyRandomSchedule(), true
	case string(types.RandomDaily):
		return getDailyRandomSchedule(), true
	default:
		return input, false
	}
}

func getDailyRandomSchedule() string {
	hour := getRandomInt(24)

	return fmt.Sprintf("0 %d * * *", hour)
}

func getHourlyRandomSchedule() string {
	minute := getRandomInt(60)

	return fmt.Sprintf("%d * * * *", minute)
}

func get15MinRandomSchedule() string {
	firstQuarter := getRandomInt(15)

	return fmt.Sprintf(
		"%d,%d,%d,%d * * * *",
		firstQuarter,
		firstQuarter+15,
		firstQuarter+30,
		firstQuarter+45,
	)
}

// getRandomInt generates random int in the range [0, max)
func getRandomInt(max int64) int64 {
	nBig, err := rand.Int(rand.Reader, big.NewInt(max))

	if err != nil {
		panic(err)
	}

	return nBig.Int64()
}
