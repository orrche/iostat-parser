package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jessevdk/go-flags"
)

type iostat struct {
	Device  string
	Rrqm    float64
	Wrqm    float64
	R       float64
	W       float64
	Rsec    float64
	Wsec    float64
	Avgrqsz float64
	Avgqusz float64
	Await   float64
	Rawait  float64
	Wawait  float64
	Svctm   float64
	Util    float64
}

func iostatParser(input io.Reader, reporter chan iostat) {
	reader := bufio.NewReader(input)
	for {
		test, _ := reader.ReadString('\n')
		test = strings.TrimSpace(test)
		oldlen := len(test)
		test = strings.Replace(test, "  ", " ", -1)
		newlen := len(test)

		for oldlen > newlen {
			test = strings.Replace(test, "  ", " ", -1)
			oldlen = newlen
			newlen = len(test)
		}

		dataLine := strings.Split(test, " ")
		if len(dataLine) != 14 {
			continue
		}

		io := iostat{}
		io.Device = dataLine[0]
		io.Rrqm, _ = strconv.ParseFloat(dataLine[1], 64)
		io.Wrqm, _ = strconv.ParseFloat(dataLine[2], 64)
		io.R, _ = strconv.ParseFloat(dataLine[3], 64)
		io.W, _ = strconv.ParseFloat(dataLine[4], 64)
		io.Rsec, _ = strconv.ParseFloat(dataLine[5], 64)
		io.Wsec, _ = strconv.ParseFloat(dataLine[6], 64)
		io.Avgrqsz, _ = strconv.ParseFloat(dataLine[7], 64)
		io.Avgqusz, _ = strconv.ParseFloat(dataLine[8], 64)
		io.Await, _ = strconv.ParseFloat(dataLine[9], 64)
		io.Rawait, _ = strconv.ParseFloat(dataLine[10], 64)
		io.Wawait, _ = strconv.ParseFloat(dataLine[11], 64)
		io.Svctm, _ = strconv.ParseFloat(dataLine[12], 64)
		io.Util, _ = strconv.ParseFloat(dataLine[13], 64)

		reporter <- io
	}
}

func writeSensorData(buf *bytes.Buffer, host string, device string, t string, value float64, time int64) {
	fmt.Fprintf(buf, "iostat,host=%s,device=%s,type=%s value=%f %d\n", host, device, t, value, time)
}

func sendData(io iostat, influxdb string, host string) {
	influxbuf := new(bytes.Buffer)

	t := time.Now().UnixNano()
	writeSensorData(influxbuf, host, io.Device, "rrqm", io.Rrqm, t)
	writeSensorData(influxbuf, host, io.Device, "wrqm", io.Wrqm, t)
	writeSensorData(influxbuf, host, io.Device, "r", io.R, t)
	writeSensorData(influxbuf, host, io.Device, "w", io.W, t)
	writeSensorData(influxbuf, host, io.Device, "rsec", io.Rsec, t)
	writeSensorData(influxbuf, host, io.Device, "wsec", io.Wsec, t)
	writeSensorData(influxbuf, host, io.Device, "avgrqsz", io.Avgrqsz, t)
	writeSensorData(influxbuf, host, io.Device, "avgqusz", io.Avgqusz, t)
	writeSensorData(influxbuf, host, io.Device, "await", io.Await, t)
	writeSensorData(influxbuf, host, io.Device, "rawait", io.Rawait, t)
	writeSensorData(influxbuf, host, io.Device, "wawait", io.Wawait, t)
	writeSensorData(influxbuf, host, io.Device, "svctm", io.Svctm, t)
	writeSensorData(influxbuf, host, io.Device, "util", io.Util, t)

	resppost, err := http.Post(influxdb, "text", influxbuf)
	if err != nil {
		log.Print(err)
		log.Print("Data will be thrown away")
	}
	defer resppost.Body.Close()
}

var opts struct {
	Influxdb string `short:"i" long:"influxdb" description:"Url to influxdb write to database e.g. http://localhost/write?db=iostat" required:"true"`
	Hostname string `short:"H" long:"hostname" description:"Hostname to repport to influxdb" required:"true"`
}

func main() {
	_, err := flags.ParseArgs(&opts, os.Args)
	if err != nil {
		log.Fatal(err)
	}

	reporter := make(chan iostat)
	go iostatParser(os.Stdin, reporter)

	for {
		stat := <-reporter

		sendData(stat, opts.Influxdb, opts.Hostname)
	}
}
