package main

import (
	"flag"
	"fmt"
	"log"
	"net/smtp"
	"os"
	"path"
	"sync"
	"time"

	"github.com/knadh/smtppool"
	"github.com/schollz/progressbar/v3"
)

type Args struct {
	SMTPServer            string
	Username              string
	Password              string
	From                  string
	To                    string
	ConcurrentConnections int
	DurationSeconds       int
	Port                  int
	TimeoutSeconds        int
	EmailCount            int
	AttachmentPath        string
}

func main() {
	args := parseArgs()

	pool, err := createSMTPPool(args)
	if err != nil {
		log.Fatalf("Failed to create SMTP pool: %v", err)
	}
	defer pool.Close()

	email, err := buildEmail(args)
	if err != nil {
		log.Fatalf("Failed to build email: %v", err)
	}

	if args.EmailCount > 0 {
		fmt.Printf("Sending %d test emails...\n", args.EmailCount)
		sendMultipleEmails(pool, email, args)
	} else {
		runBenchmark(pool, email, args)
	}
}

func parseArgs() Args {
	args := Args{}
	flag.StringVar(&args.SMTPServer, "smtp-server", "", "SMTP server address")
	flag.StringVar(&args.Username, "username", "", "SMTP username")
	flag.StringVar(&args.Password, "password", "", "SMTP password")
	flag.StringVar(&args.From, "from", "", "Sender email address")
	flag.StringVar(&args.To, "to", "", "Recipient email address")
	flag.IntVar(&args.ConcurrentConnections, "concurrent-connections", 10, "Number of concurrent connections")
	flag.IntVar(&args.DurationSeconds, "duration-seconds", 60, "Duration of the benchmark in seconds")
	flag.IntVar(&args.Port, "port", 25, "SMTP server port")
	flag.IntVar(&args.TimeoutSeconds, "timeout-seconds", 10, "Connection timeout in seconds")
	flag.IntVar(&args.EmailCount, "email-count", 0, "Number of emails to send (overrides duration)")
	flag.StringVar(&args.AttachmentPath, "attachment", "", "Path to file to attach to the email")
	flag.Parse()
	return args
}

func createSMTPPool(args Args) (*smtppool.Pool, error) {
	auth := smtp.PlainAuth("", args.Username, args.Password, args.SMTPServer)
	config := smtppool.Opt{
		Host:        args.SMTPServer,
		Port:        args.Port,
		MaxConns:    args.ConcurrentConnections,
		IdleTimeout: time.Duration(args.TimeoutSeconds) * time.Second,
		Auth:        auth,
	}

	return smtppool.New(config)
}

func addAttachment(email *smtppool.Email, filepath string) error {
	if filepath == "" {
		return nil
	}

	file, err := os.Open(filepath)
	if err != nil {
		return fmt.Errorf("failed to open attachment file: %v", err)
	}
	defer file.Close()

	filename := path.Base(filepath)
	email.Attach(file, filename, "")

	return nil
}

func buildEmail(args Args) (smtppool.Email, error) {
	email := smtppool.Email{
		From:    args.From,
		To:      []string{args.To},
		Subject: "SMTP Benchmark",
		Text:    []byte("This is a test email for SMTP benchmarking."),
	}

	if err := addAttachment(&email, args.AttachmentPath); err != nil {
		return email, err
	}

	return email, nil
}

func sendEmail(pool *smtppool.Pool, email smtppool.Email) error {
	return pool.Send(email)
}

func runBenchmark(pool *smtppool.Pool, email smtppool.Email, args Args) {
	startTime := time.Now()
	endTime := startTime.Add(time.Duration(args.DurationSeconds) * time.Second)

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, args.ConcurrentConnections)
	totalSent := 0
	totalLatency := time.Duration(0)

	bar := progressbar.NewOptions(-1,
		progressbar.OptionSetDescription("Sending emails"),
		progressbar.OptionSetWidth(50),
		progressbar.OptionSetRenderBlankState(true),
		progressbar.OptionShowCount(),
		progressbar.OptionShowIts(),
		progressbar.OptionSetTheme(progressbar.Theme{Saucer: "=", SaucerHead: ">", SaucerPadding: " ", BarStart: "[", BarEnd: "]"}),
	)

	for time.Now().Before(endTime) {
		wg.Add(1)
		semaphore <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-semaphore }()

			start := time.Now()
			err := sendEmail(pool, email)
			latency := time.Since(start)

			if err != nil {
				log.Printf("Error sending email: %v", err)
			} else {
				totalSent++
				totalLatency += latency
				bar.Add(1)
			}
		}()
	}

	wg.Wait()
	bar.Finish()
	totalDuration := time.Since(startTime)
	printResults(totalSent, totalLatency, totalDuration)
}

func sendMultipleEmails(pool *smtppool.Pool, email smtppool.Email, args Args) {
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, args.ConcurrentConnections)
	totalSent := 0
	totalLatency := time.Duration(0)
	startTime := time.Now()

	bar := progressbar.NewOptions(args.EmailCount,
		progressbar.OptionSetDescription("Sending emails"),
		progressbar.OptionSetWidth(50),
		progressbar.OptionShowCount(),
		progressbar.OptionShowIts(),
		progressbar.OptionSetTheme(progressbar.Theme{Saucer: "=", SaucerHead: ">", SaucerPadding: " ", BarStart: "[", BarEnd: "]"}),
	)

	for i := 0; i < args.EmailCount; i++ {
		wg.Add(1)
		semaphore <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-semaphore }()

			start := time.Now()
			err := sendEmail(pool, email)
			latency := time.Since(start)

			if err != nil {
				log.Printf("Error sending email: %v", err)
			} else {
				totalSent++
				totalLatency += latency
				bar.Add(1)
			}
		}()
	}

	wg.Wait()
	bar.Finish()
	totalDuration := time.Since(startTime)
	printResults(totalSent, totalLatency, totalDuration)
}

func printResults(totalSent int, totalLatency, totalDuration time.Duration) {
	throughput := float64(totalSent) / totalDuration.Seconds()
	avgLatency := totalLatency / time.Duration(totalSent)

	fmt.Printf("Total emails sent: %d\n", totalSent)
	fmt.Printf("Throughput: %.2f emails/second\n", throughput)
	fmt.Printf("Average latency: %v\n", avgLatency)
	fmt.Printf("Total duration: %v\n", totalDuration)
}
