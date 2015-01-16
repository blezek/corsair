Examples:

corsair --remote http://google.com

This example serve the current working directory on port 8080, forwarding any unknown requests to http://google.com.

corsair --dir path/to/webstuff  --port 8900 --remote https://itunes.apple.com

Serve path/to/webstuff on port 8900 (http://localhost:8900), any REST calls will be forwarded to https://itunes.apple.com
