package env

import "os"

var FromAgent = os.Getenv("TDL_AGENT") == "1"
