package utils

import "flag"

// var (
// 	Method    = flag.String("m", "GET", "Current method\n")
// 	ShowBody  = flag.Bool("b", false, "show response body\n")
// 	Limit     = flag.Int("limit", 0, "sets limit to body\n")
// 	Rate      = flag.Int("rate", 0, "sends request per N seconds\n")
// 	WriteFile = flag.Bool("f", false, "")
// 	TimeLimit = flag.Int("time-limit", 10, "Use with'--rate' flag to limit overall time of making requests")
// 	BodyData  = flag.String("d", "", "Needs to send request with json body")
// )

type Config struct {
	Method    string
	ShowBody  bool
	Limit     int
	Rate      int
	WriteFile bool
	TimeLimit int
	BodyData  string
}

func ParseFlags() *Config {
	cfg := &Config{}
	flag.StringVar(&cfg.Method, "m", "GET", "HTTP method")
	flag.BoolVar(&cfg.ShowBody, "b", false, "Show response body")
	flag.IntVar(&cfg.Limit, "limit", 0, "Limit body output")
	flag.IntVar(&cfg.Rate, "rate", 0, "Request rate per second")
	flag.BoolVar(&cfg.WriteFile, "f", false, "Write output to file")
	flag.IntVar(&cfg.TimeLimit, "time-limit", 10, "Time limit in seconds")
	flag.StringVar(&cfg.BodyData, "d", "", "Request body")
	flag.Parse()
	return cfg
}
