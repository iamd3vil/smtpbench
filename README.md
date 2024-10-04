# SMTP Benchmark Tool

This Go-based tool allows you to benchmark SMTP servers by sending test emails and measuring throughput and latency.

## Features

- Configurable number of concurrent connections
- Adjustable benchmark duration or fixed email count
- SMTP authentication support
- Progress bar for visual feedback
- Detailed performance metrics output

## Installation

1. Ensure you have Go installed on your system.
2. Clone this repository:
   ```
   git clone https://git.simhadri.rocks/sarat/smtpbench.git
   ```
3. Navigate to the project directory:
   ```
   cd smtpbench
   ```
4. Install dependencies:
   ```
   go mod tidy
   ```
5. Build the binary:
  ```
  go build
  ```

## Usage

Run the tool with the following command:

```
./smtpbench [flags]
```

### Flags

- `-smtp-server`: SMTP server address (required)
- `-username`: SMTP username for authentication
- `-password`: SMTP password for authentication
- `-from`: Sender email address (required)
- `-to`: Recipient email address (required)
- `-concurrent-connections`: Number of concurrent connections (default: 10)
- `-duration-seconds`: Duration of the benchmark in seconds (default: 60)
- `-port`: SMTP server port (default: 25)
- `-timeout-seconds`: Connection timeout in seconds (default: 10)
- `-email-count`: Number of emails to send (overrides duration if set)

## Examples

Run a 60-second benchmark with 10 concurrent connections:
```
./smtpbench -smtp-server smtp.example.com -from sender@example.com -to recipient@example.com -username myuser -password mypass
```

Send 1000 emails with 20 concurrent connections:
```
./smtpbench -smtp-server smtp.example.com -from sender@example.com -to recipient@example.com -username myuser -password mypass -concurrent-connections 20 -email-count 1000
```

## Output

The tool will display a progress bar during execution and print the following metrics upon completion:

- Total emails sent
- Throughput (emails/second)
- Average latency
- Total duration

## Note

This tool is for benchmarking purposes only. Be sure to comply with the SMTP server's usage policies and obtain necessary permissions before running extensive tests.

## License

[MIT License](LICENSE)
