package apksm

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"time"
)

func calculateServiceUptime(statusAtTime []*statusAtTime) string {
	if len(statusAtTime) == 0 {
		return "unknown"
	}

	var sum float64

	for _, val := range statusAtTime {
		var i float64
		if val.Status {
			i = 1
		} else {
			i = 0
		}
		sum += i
	}

	return fmt.Sprintf("%.2f", sum/float64(len(statusAtTime))*100)
}

func lastStatus(statusAtTime []*statusAtTime) string {
	if len(statusAtTime) == 0 {
		return "Not yet checked"
	}
	lastChecked := statusAtTime[len(statusAtTime)-1]
	difference := time.Since(lastChecked.Time)
	status := "OK"
	if !lastChecked.Status {
		status = "ERR"
	}
	return fmt.Sprintf("%s, %.0f seconds ago", status, difference.Seconds())
}

func RunHttp(address string, monitor *Monitor) {
	funcMap := template.FuncMap{
		"calculateServiceUptime": calculateServiceUptime,
		"lastStatus":             lastStatus,
	}

	t := template.Must(template.New("main").Funcs(funcMap).Parse(`<!DOCTYPE html>
<html lang="en">
  <head>
    <!-- Required meta tags -->
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no">
	<title>APK Service Monitor - Dashboard</title>
	
    <!-- Bootstrap CSS -->
    <link rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/bootstrap/4.0.0-beta/css/bootstrap.min.css" integrity="sha384-/Y6pD6FV/Vv2HJnA6t+vslU6fwYXjCFtcEpHbNJ0lyAFsXTsjBbfaDjzALeQsN6M" crossorigin="anonymous">
  </head>
  <body>
	<div class="container">
		<br>
		<center><h1>APK Service Monitor Dashboard</h1></center>
		<hr>
		<div class="row">
			{{ range $service, $statusAtTime := .}}
			<div class="col-md-4">
				<div class="card" style="margin-top: 5px;">
					<div class="card-body">
						<h4 class="card-title">Service: {{ $service.Name }}</h4>
						<p class="card-text">{{ $service }}<br>tested {{ len $statusAtTime }} times<br>{{ $statusAtTime | lastStatus }}</p>
						<p class="card-text"><b>UpTime:</b> {{ $statusAtTime | calculateServiceUptime }}%</p>

					</div>
				</div>
			</div>
			{{ end }}
		</div>
	</div>
  </body>
</html>`))

	http.HandleFunc("/", func(rw http.ResponseWriter, req *http.Request) {
		_ = t.Execute(rw, monitor.serviceStatusData.GetServiceStatus())
	})

	http.HandleFunc("/api", func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("Content-Type", "application/json")

		jsonBytes, err := json.Marshal(monitor.serviceStatusData.GetServiceStatus())
		if err != nil {
			jsonError, _ := json.Marshal(struct {
				Message string `json:"message"`
			}{Message: "Unable to format JSON."})

			_, _ = rw.Write(jsonError)
			return
		}

		_, _ = rw.Write(jsonBytes)
	})

	err := http.ListenAndServe(address, nil)
	if err != nil {
		panic("error in running http server")
		return
	}
}
