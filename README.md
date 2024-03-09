# Scrappy

Scrappy is a web scraping tool written in Go. It is designed to download and parse web pages concurrently, based on the provided configuration.

## Description

Scrappy starts from a root URL and recursively downloads and parses all linked pages that match the specified prefix. The tool includes features such as concurrent downloading and parsing, depth limitation, and URL filtering.

## Installation

To install Scrappy, you need to have Go installed on your machine. Once Go is installed, you can clone the repository and build the project:

```bash
git clone https://github.com/PatrikValkovic/scrappy.git
cd scrappy
go build
```

## Usage

You can run Scrappy using the following command:

```bash
./scrappy
```

Scrappy uses a configuration file named `env.yaml` located in the project root directory. You can also provide configuration via command-line flags.

### Configuration Options

- `parse-root`: The starting URL for parsing.
- `output-dir`: The directory where downloaded files will be stored.
- `max-depth`: The maximum depth of the crawling.
- `required-prefix`: The prefix that all the links must have. By default match parse root option.
- `environment`: The environment setting ("development" or "production"). Change log outputs.
- `download-concurrency`: The maximum number of files to download in parallel.
- `parse-concurrency`: The maximum number of files to parse in parallel.
- `ignore-pattern`: List of regular expressions to ignore during parsing. This can be specified multiple times.

For example, to run Scrappy with a maximum depth of 10 and 4 concurrent downloads, you would use:

```bash
./scrappy --max-depth=10 --download-concurrency=4
```

## Contributing

Contributions are welcome. Please open an issue or submit a pull request on GitHub.

## License

Scrappy is licensed under the GPLv3 License. See the `LICENSE` file for more information.
