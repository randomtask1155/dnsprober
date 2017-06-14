package main 

import (
	"net/http"
	"os"
	"os/exec"
	"time"
	"strconv"
	"fmt"
	"strings"
)

var (
	dig = "dig"
	interval int64
	hostname = "www.google.com"
	dnsServers  []string
	StatsCounter Stats
)

type DnsStats struct {
	Passed	int `json:"passed"`
	Failed int `json:"failed"`
	ErrOutput []DnsError `json:"errors"`
}

type DnsError struct {
	TimeStamp string
	Output []byte `json:"output"` 

}

type Stats map[string]*DnsStats


func rootHandler(w http.ResponseWriter, r *http.Request) {
	html := `<html>
<table border=1>
	<tr>
		<th>DNS Server</th>
		<th>Passed</th>
		<th>Failed</th>
		<th>10 Recent Errors</th>
	</tr>
`
	for _, server := range dnsServers {
		erroutput := ""
		for _, errs := range StatsCounter[server].ErrOutput {
			erroutput += fmt.Sprintf("<p>%s: %s</p>", errs.TimeStamp, errs.Output)
		}
		html +=  fmt.Sprintf(`
	<tr>
		<td>%s</td>
		<td>%d</td>
		<td>%d</td>
		<td>%s</td>
	</tr>
`, server, StatsCounter[server].Passed, StatsCounter[server].Failed, erroutput)
	}
	html += "</table>"
	w.Write([]byte(html))
}


// only keep 10 errors in memory 
func (stats *DnsStats) AppendError(b []byte) {
	t := fmt.Sprintf("%s", time.Now())
	if len(stats.ErrOutput) >= 9 {
		
		stats.ErrOutput = stats.ErrOutput[1:] 
		stats.ErrOutput = append(stats.ErrOutput, DnsError{t,b})
	} else {
		stats.ErrOutput = append(stats.ErrOutput, DnsError{t, b})
	}
}

func dnsProber() {
	for {
		for _, server := range dnsServers {
			out, err := exec.Command(dig, "@" + server, hostname).CombinedOutput()
			if err != nil {
				StatsCounter[server].Failed += 1
				fmt.Printf("DNS FAILED:%s @%s %s:\n%s\n%s\n", dig, server, hostname, out, err)
				StatsCounter[server].AppendError(out)
			} else {
				StatsCounter[server].Passed += 1
				fmt.Printf("DNS PASSED:%s @%s %s\n", dig, server, hostname)
			}
			time.Sleep(500 * time.Millisecond) // buffer between sends
		}
		time.Sleep(time.Duration(interval))
	}
}

func main() {
	var err error
	dig, err = exec.LookPath(dig)
	if err !=nil {
		fmt.Printf("Failed to find dig command in $PATH: %s\n", err)
		os.Exit(1)
	}

	var defaultInterval int64
	defaultInterval = int64(15 * time.Second)
	if os.Getenv("INTERVAL") == "" {
		interval = defaultInterval
		fmt.Printf("setting default interval to %d microseconds\n", interval)
	} else {
		i, err := strconv.Atoi(os.Getenv("INTERVAL"))
		if err != nil {
			
			fmt.Printf("WARNING: could not convert $INTERVAL=%s string to integer so using default of %d microseconds\n", os.Getenv("INTERVAL"), defaultInterval)
			interval = defaultInterval
		} else {
			interval = int64(i) * int64(time.Second)
		}
	}

	if os.Getenv("PINGHOST") == "" {
		fmt.Printf("PINGHOST not set so using default %s\n", hostname)
	} else {
		hostname = os.Getenv("PINGHOST")
		fmt.Printf("PINGHOST set to %s\n", hostname)
	}


	if os.Getenv("DNS_SERVERS") == "" {
		fmt.Println("DNS_SERVERS must be set using comma delimited format \"1.1.1.1,2.2.2.2\"")
		os.Exit(2)
	}
	dnsServers = strings.Split(os.Getenv("DNS_SERVERS"), ",")

	StatsCounter = make(Stats)
        for _, server := range dnsServers {
                StatsCounter[server] = &DnsStats{}
        }

	go dnsProber()
	http.HandleFunc("/", rootHandler)
	http.ListenAndServe(":" + os.Getenv("PORT"), nil)
}
