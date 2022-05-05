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
    resp = requests.get(url) # additional query_params
    print(url)
    print(resp)
    print(resp.content)
    # parse query params
    url_query = parse.parse_qs(parse.urlparse(url).query)
    url_query_params = {k:v[0] if v and len(v) == 1 else v for k,v in url_query.items()}
    # For example, http://example.com/?foo=bar&foo=baz&bar=baz would return:
    # {'foo': ['bar', 'baz'], 'bar': 'baz'}
    print(url_query_params)

    yaml_queries, columns = parseYAML()
    print(yaml_queries)
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

    print(placeholders)
    # map strings from placeholder to url_query dictionary
    qp_list = [] 
    for p in placeholders:
        if p in url_query_params:
            qp_list.append(url_query_params[p])
    print(qp_list)
            
    psql.get_connection_by_config(psql, 'database.ini', 'postgresql_conn_data')            
    cursor = psql.get_cursor(psql)

    #print(sql.SQL(yaml_queries['gender'][0]).format(sql.SQL(', ').join(map(sql.Identifier, qp_list[0]))))
    #print(sql.SQL(yaml_queries['gender'][0]).format(concept_id=qp_list[0]).as_string(None))
    # zipped = list(zip(placeholders, qp_list))
    # print(zipped[0][0])
    for key,val in yaml_queries.items():
        for i in range(len(val)):
            if placeholders[i] == "concept_id":
                result = cursor.execute(sql.SQL(yaml_queries[key][i]).format(concept_id=qp_list[i]))
                print(sql.SQL(yaml_queries[key][i]).format(concept_id=qp_list[i]).as_string(None))
            elif placeholders[i] == "concept_name":
                result = cursor.execute(sql.SQL(yaml_queries[key][i]).format(concept_name=qp_list[i]))
                print(sql.SQL(yaml_queries[key][i]).format(concept_name=qp_list[i]).as_string(None))
            ret = result.fetchone()
            print(ret)
            if ret and len(ret) > 0:
                col_vals = []
                for name in ret:
                    for l in columns:
                        for name in l:
                            col_vals.append(name)
                if col_vals and len(col_vals) > 0:
                    return col_vals
    # for block in resp.json():
    #     for key, values in block.items():
    #         for val in values:
    #             cursor = psql.get_cursor(psql)
    #             # if need to pass directives use %s to execute
    #             for arr in qp_list:
    #                 result = cursor.execute(sql.SQL(yaml_queries[key]).format(
    #                     sql.SQL(', ').join(map(sql.Placeholder, arr))))
    #                 ret = result.fetchone()
    #                 if ret and len(ret) > 0:
    #                     col_vals = []
    #                     for name in ret:
    #                         for l in columns:
    #                             for name in l:
    #                                 col_vals.append(name)
    #                     if col_vals and len(col_vals) > 0:
    #                         return col_vals
    return None

if __name__ == '__main__':
    # first create an instance of PostgresqlManager class.
    postgresql_manager = PostgresqlManager()                        
    postgresql_manager.get_connection_by_config('database.ini', 'postgresql_conn_data')
   
process("http://localhost:8000?id=gender&concept_id=8507&concept_name=blah")
# http://localhost:8000?id=gender&concept_id=123&concept_name=blah