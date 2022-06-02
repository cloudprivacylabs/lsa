hostName = "localhost"
serverPort = 8011

from collections import defaultdict
import yaml

from postgresql_manager import PostgresqlManager

from http.server import BaseHTTPRequestHandler, HTTPServer
from urllib import parse
import json

db = PostgresqlManager
db.get_connection_by_config(db, 'database.ini', 'postgresql_conn_data') 

def parseYAML() -> defaultdict:
    queries = defaultdict(list)
    columns = []
    with open("queries.yaml") as yaml_file:
        vs_list = yaml.full_load(yaml_file)
        for item, doc in vs_list.items():
            for key,val in doc[0].items():
                if key == "tableId":
                    id = val
                elif key == "queries":
                    for i in range(len(val)):
                        for k, v in val[i].items():
                            if k == "query":
                                queries[id].append(v)
                            elif k == "columns":
                                columns.append(v)
    return queries

config = parseYAML()


def process(url):
    # parse query params
    url_query = parse.parse_qs(parse.urlparse(url).query)
    url_query_params = {k:v[0] if v and len(v) == 1 else v for k,v in url_query.items()}
    # For example, http://example.com/?foo=bar&foo=baz&bar=baz would return:
    # {'foo': ['bar', 'baz'], 'bar': 'baz'}

    # iterate through yaml dictionary and for each query, execute statement, binding url query parameters
    for key,val in config.items():
        for i in range(len(val)):
            result = db.get_results(db,config[key][i],url_query_params)
            if not result or result == None:
                continue
            return result
            
    return None

class VS_Server(BaseHTTPRequestHandler):
    def do_GET(self):
        query_params = parse.urlparse(self.path).query
        full_url = "http://" + ''.join([hostName, ':', str(serverPort), '?',query_params])
        result=process(full_url)
        self.send_response(200)
        self.send_header("Content-type", "application/json; charset=UTF-8")
        self.end_headers()
        self.wfile.write(json.dumps(result).encode("utf-8"))

if __name__ == '__main__':
    webServer = HTTPServer((hostName, serverPort), VS_Server)
    print("Server started http://%s:%s" % (hostName, serverPort))
    try:
        webServer.serve_forever()
    except KeyboardInterrupt:
        pass
    webServer.server_close()
    
