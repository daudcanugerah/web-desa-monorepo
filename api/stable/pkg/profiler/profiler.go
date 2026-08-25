package debug

import (
	profile "github.com/bygui86/multi-profile/v2"
)

type CustomProfiler interface {
	Stop()
}

type emptyProfile struct{}

func (e *emptyProfile) Stop() {}

func SetProfiler(mode []string) []CustomProfiler {
	results := []CustomProfiler{}
	for _, v := range mode {
		switch v {
		case "cpu":
			results = append(results, profile.CPUProfile(&profile.Config{Path: "."}).Start())
		case "ram":
			results = append(results, profile.MemProfile(&profile.Config{Path: "."}).Start())
		case "gorountine":
			results = append(results, profile.GoroutineProfile(&profile.Config{Path: "."}).Start())
		case "trace":
			results = append(results, profile.TraceProfile(&profile.Config{Path: "."}).Start())
		case "mutex":
			results = append(results, profile.MutexProfile(&profile.Config{Path: "."}).Start())
		case "block":
			results = append(results, profile.BlockProfile(&profile.Config{Path: "."}).Start())
		default:
		}
	}
	return results
}
