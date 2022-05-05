from ast import Tuple
from collections import defaultdict
import yaml
import requests
from psycopg import sql
from postgresql_manager import PostgresqlManager
from urllib import parse
import re

psql = PostgresqlManager

def parseYAML() -> Tuple(defaultdict, list):
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
    return (queries, columns)

def process(url):
    # url as param -> received from Go?
    # use CLI cmd python3 -m http.server to start listening on port 8000
    # input ([terminology:loinc, value: male] table: gender) --> output: (value:8503)
    # resp = requests.get(url) # additional query_params
    # parse query params
    url_query = parse.parse_qs(parse.urlparse(url).query)
    url_query_params = {k:v[0] if v and len(v) == 1 else v for k,v in url_query.items()}
    # For example, http://example.com/?foo=bar&foo=baz&bar=baz would return:
    # {'foo': ['bar', 'baz'], 'bar': 'baz'}

    yaml_queries, columns = parseYAML()
    # defaultdict(<class 'list'>, {'gender': ["select concept_id,concept_name from concepts where vocabulary_id='gender' 
    # and concept_id=$concept_id", "select concept_id,concept_name from concepts where vocabulary_id='gender' 
    # and concept_name=$concept_name"]})
    
    # Extract from yaml_queries -> what is a placeholder $, substring search from start of placeholder
    placeholders = []
    for key,val in yaml_queries.items():
        for s in val:
            result = [_.start() for _ in re.finditer("{", s)] 
            for idx_start in result:
                x = idx_start
                for c in s[idx_start:]:
                    if c == "}":
                        # +1, exclude {
                        placeholders.append(s[idx_start+1:x])
                        break
                    x += 1

    # map strings from placeholder to url_query dictionary
    qp_list = [] 
    for p in placeholders:
        if p in url_query_params:
            qp_list.append(url_query_params[p])
            
    psql.get_connection_by_config(psql, 'database.ini', 'postgresql_conn_data')            
    cursor = psql.get_cursor(psql)

    for key,val in yaml_queries.items():
        for i in range(len(val)):
            if placeholders[i] == "concept_id":
                result = cursor.execute(sql.SQL(yaml_queries[key][i]).format(concept_id=qp_list[i]))
            elif placeholders[i] == "concept_name":
                result = cursor.execute(sql.SQL(yaml_queries[key][i]).format(concept_name=qp_list[i]))
            ret = result.fetchone()
            if not ret:
                continue
            print(ret)
            return ret
            
    return None

if __name__ == '__main__':
    # first create an instance of PostgresqlManager class.
    postgresql_manager = PostgresqlManager()                        
   
process("http://localhost:8000?id=gender&concept_id=8507&concept_name=blah")
# http://localhost:8000?id=gender&concept_id=123&concept_name=blah