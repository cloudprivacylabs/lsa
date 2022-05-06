hostName = "localhost"
serverPort = 8000
from http.server import BaseHTTPRequestHandler, HTTPServer
from urllib import parse

from main import process

class VS_Server(BaseHTTPRequestHandler):
    def do_GET(self):
        self.send_response(200)
        self.send_header("Content-type", "text/html")
        self.end_headers()

        query_params = parse.urlparse(self.path).query
        full_url = "http://" + ''.join([hostName, ':', str(serverPort), '?',query_params])
        process(full_url)

        self.wfile.write(bytes("<html><head><title>https://pythonbasics.org</title></head>", "utf-8"))
        self.wfile.write(bytes("<p>Request: %s</p>" % self.path, "utf-8"))
        self.wfile.write(bytes("<body>", "utf-8"))
        self.wfile.write(bytes("<p>This is an example web server.</p>", "utf-8"))
        self.wfile.write(bytes("</body></html>", "utf-8"))

if __name__ == '__main__':
    webServer = HTTPServer((hostName, serverPort), VS_Server)
    print("Server started http://%s:%s" % (hostName, serverPort))
    try:
        webServer.serve_forever()
    except KeyboardInterrupt:
        pass
    webServer.server_close()
    