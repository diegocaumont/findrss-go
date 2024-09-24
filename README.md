# FindRSS

FindRSS is a command-line tool that searches for the Atom or RSS feed of a given website by trying various common and uncommon paths.

## Features

- Searches for RSS/Atom feeds by generating and checking multiple possible feed URLs
- Supports concurrent requests to improve performance
- Reads input from a JSON file containing a list of websites
- Updates the JSON file with the discovered RSS/Atom feed URLs
- Handles cases where no feed is found for a website

## Installation

1. Make sure you have Go installed on your system.
2. Clone this repository:
   ```
   git clone https://github.com/yourusername/findrss-github.git
   ```
3. Navigate to the project directory:
   ```
   cd findrss-github
   ```
4. Build the executable:
   ```
   go build -o findrss
   ```

## Usage

1. Prepare a JSON file containing a list of websites to search for RSS/Atom feeds. The JSON file should have the following format:
   ```json
   [
     {
       "url": "https://example1.com"
     },
     {
       "url": "https://example2.com"
     }
   ]
   ```
2. Run the `findrss` tool, providing the path to the JSON file as an argument:
   ```
   ./findrss input.json
   ```
3. The tool will process each website concurrently, searching for RSS/Atom feeds.
4. The discovered feed URLs will be added to the JSON file, and the updated file will be saved.
5. If no feed is found for a website, it will be marked with `"NO_RSS_FEED"` in the JSON file.

## Contributing

Contributions are welcome! If you find any issues or have suggestions for improvements, please open an issue or submit a pull request.

## License

This project is licensed under the [MIT License](LICENSE).

(Created in collaboration with [j3s](https://j3s.sh).)

