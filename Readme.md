Corsair
=======

*Overcoming [CORS](http://en.wikipedia.org/wiki/Cross-origin_resource_sharing) problems on the high seas.*


***corsair*** noun \ˈkȯr-ˌser;'\
  1.  A pirate; one who cruises or scours the ocean, with an armed vessel, without a commission from any prince or state
  2. A Go application to proxy REST calls and avoid CORS violations



## Why Corsair Exists

Frequently, I develop [single page JavaScript applications (SPA)](http://en.wikipedia.org/wiki/Single-page_application).  Usually, the REST API is completed long before the JavaScript application is completed.  REST also tends to be slow changing, while the SPA needs constant updating.  We would typically jump through hoops to serve the SPA through the REST server.  Corsair eliminates these hoops by giving the browser a single server, so static pages can be served from a local file, and REST calls are proxied to the remote server.  This is useful for developing [Notion](https://github.com/dblezek/Notion), and [Freeboard dashboards](https://github.com/Freeboard/freeboard).

Once the SPA is completed, it can be integrated into the build system for final production deployment.


## How Corsair Works
Corsair's purpose to proxy a [REST server](http://en.wikipedia.org/wiki/Representational_state_transfer) and serve static files.  Each HTTP request that Corsair receives is examined for a match in the static files directory.  If matched, the file is served.  Any unknown URLs are forwarded ([proxied](http://en.wikipedia.org/wiki/Proxy_server)) to the REST server.

Suppose you are serving `/path/index.html` and `/path/image.png` on port 8080, with the proxied remote site `https://remote.com:1234`.  From your trusty browser you open http://localhost:8080.  If we listen in on the conversation:

**Browser**: `> GET / HTTP/1.1`  (Hey, give me `/`)

**Corsair**: Hmm, I'll treat `/` as a request for `index.html`.  Looks like I have it right here `/path/index.html`.  Hey browser, here you go!

**Browser**: `> GET /image.png HTTP/1.1` (after parsing `index.html`)

**Corsair**: Yup, got that right here (sends `/path/image.png`)

**Browser**: `jquery.getJSON('/rest/v1/version')` (makes a REST request using jquery)

**Corsair**: Hmm, I don't seem to have a file called `/path/rest/v1/version`, I'll punt to `http://remote.com:1234/rest/v1/version` and see what he has.  (returns the result of forwarding the request)


## Usage

```bash
Usage: corsair [options]
```

Important Options:

* **`--dir`**, **`-d`** Path to static files, default is current directory
* **`--port`**, **`-p`** Port to serve on, default is http://localhost:8080
* **`--remote`**, **`-r`** Remote server to proxy, default is http://localhost:80
* **`--livereload`**, **`-l`** Use Livereload to monitor the static files path and reload pages on changes (**planned**)

## Example Command Lines

`corsair --remote http://google.com`

This example serve the current working directory on port 8080, forwarding any unknown requests to http://google.com.

`corsair --dir path/to/webstuff  --port 8900 --remote https://itunes.apple.com`

Serve path/to/webstuff on port 8900 (http://localhost:8900), any REST calls will be forwarded to https://itunes.apple.com
