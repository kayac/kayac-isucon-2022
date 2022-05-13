package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Songmu/timeout"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/fujiwara/ridge"
	"github.com/natureglobal/realip"
	"golang.org/x/net/context"
)

var middleware func(http.Handler) http.Handler

var scoreRegexp = regexp.MustCompile(`SCORE: (\d+)`)

func init() {
	var ipnets []*net.IPNet
	for _, n := range []string{"10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16", "127.0.0.1/32"} {
		_, ipnet, err := net.ParseCIDR(n)
		if err != nil {
			panic(err)
		}
		ipnets = append(ipnets, ipnet)
	}
	middleware = realip.MustMiddleware(&realip.Config{
		RealIPFrom:      ipnets,
		RealIPHeader:    realip.HeaderXForwardedFor,
		RealIPRecursive: true,
	})
}

func main() {
	var mux = http.NewServeMux()
	mux.HandleFunc("/bench", handleBench)
	mux.HandleFunc("/", handleRoot)
	ridge.Run(":8080", "/", middleware(mux))
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	host := r.Header.Get(realip.HeaderXRealIP)
	fmt.Fprintln(w, "Hello "+host)
	fmt.Fprintln(w, "TAG "+os.Getenv("TAG"))
}

func handleBench(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	host := r.Header.Get(realip.HeaderXRealIP)
	target := fmt.Sprintf("http://%s", host)
	teamID := r.FormValue("team_id")
	if teamID == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "team_id is required")
		return
	}

	args := []string{"-target-url", target}
	log.Println("start bench", args)
	tio := &timeout.Timeout{
		Cmd:       exec.Command("./bench", args...),
		Duration:  120 * time.Second,
		KillAfter: 5 * time.Second,
	}
	exitStatus, stdout, stderr, err := tio.Run()

	w.Header().Set("Content-Type", "text/plain")
	for _, o := range []io.Writer{w, os.Stdout} {
		fmt.Fprintln(o, "bench", strings.Join(args, " "))
		fmt.Fprintln(o, "exit status:", exitStatus)
		if err != nil {
			fmt.Fprintln(o, "error:", err)
		}
		fmt.Fprintln(o, stdout)
	}
	fmt.Fprintln(os.Stderr, stderr)

	score, err := parseScore(stdout)
	if err != nil {
		log.Println("failed to parse score:", err)
		fmt.Fprintln(w, err)
		return
	}
	if err := postScore(teamID, score); err != nil {
		log.Println("failed to post score:", err)
		fmt.Fprintln(w, err)
	}
}

func parseScore(s string) (int64, error) {
	m := scoreRegexp.FindStringSubmatch(s)
	if m == nil {
		return -1, fmt.Errorf("score is not found in stdout")
	}
	return strconv.ParseInt(m[1], 10, 64)
}

func postScore(teamID string, score int64) error {
	sess := session.Must(session.NewSession())
	svc := cloudwatch.New(sess)
	in := &cloudwatch.PutMetricDataInput{
		Namespace: aws.String("isucon"),
		MetricData: []*cloudwatch.MetricDatum{{
			MetricName: aws.String("score"),
			Timestamp:  aws.Time(time.Now()),
			Value:      aws.Float64(float64(score)),
			Dimensions: []*cloudwatch.Dimension{{
				Name:  aws.String("team_id"),
				Value: aws.String(teamID),
			}},
		}},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err := svc.PutMetricDataWithContext(ctx, in)

	return err
}
