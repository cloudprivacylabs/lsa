from collections import defaultdict
import yaml
from postgresql_manager import PostgresqlManager
from urllib import parse

psql = PostgresqlManager
psql.get_connection_by_config(psql, 'database.ini', 'postgresql_conn_data') 

def parseYAML() -> defaultdict:
    queries = defaultdict(list)
    columns = []
    with open("example_queries.yaml") as yaml_file:
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

def process(url):
    # parse query params
    url_query = parse.parse_qs(parse.urlparse(url).query)
    url_query_params = {k:v[0] if v and len(v) == 1 else v for k,v in url_query.items()}
    # For example, http://example.com/?foo=bar&foo=baz&bar=baz would return:
    # {'foo': ['bar', 'baz'], 'bar': 'baz'}

    yaml_queries = parseYAML()
    # defaultdict(<class 'list'>, {'gender': ["select concept_id,concept_name from concepts where vocabulary_id='gender' 
    # and concept_id=$concept_id", "select concept_id,concept_name from concepts where vocabulary_id='gender' 
    # and concept_name=$concept_name"]})
    
    # grab cursor needed to executing SQL calls           
    cursor = psql.get_cursor(psql)

    # iterate through yaml dictionary and for each query, execute statement, binding url query parameters
    for key,val in yaml_queries.items():
        for i in range(len(val)):
            result = cursor.execute(yaml_queries[key][i], url_query_params)
            row = result.fetchone()
            print(row)
            if not row or row == None:
                continue
            return row
            
    return None

if __name__ == '__main__':
    # first create an instance of PostgresqlManager class.
    postgresql_manager = PostgresqlManager()                        
   
# process("http://localhost:8000?id=gender&concept_id=8507&concept_name=rf")
# http://localhost:8000?id=gender&concept_id=123&concept_name=blah