gender_query = 
        "SELECT {field} FROM {table} WHERE {pkey} = %s").format(
        field=sql.Identifier('my_name'),
        table=sql.Identifier('concept'),
        pkey=sql.Identifier('id'))