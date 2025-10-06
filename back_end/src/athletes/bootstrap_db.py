import os
from pathlib import Path
import sqlite3
import psycopg2

def boostrap_db(db_type: str = "sqlite"):

    with open(Path(__file__).parent / "schema.sql") as f:
        schema_sql = f.read()

    if db_type == "sqlite":
        db_url = os.getenv("DB_URL")
        if db_url is None:
            raise TypeError("Environmental variable 'db_url' is required and is missing.")
        
        connection = sqlite3.connect(db_url)
    
    elif db_type == "postgres":
        connection = psycopg2.connect(
        dbname=os.getenv("POSTGRES_DB"),
        user=os.getenv("POSTGRES_USER"),
        password=os.getenv("POSTGRES_PASSWORD"),
        host=os.getenv("DB_HOST"), 
        port=os.getenv("DB_PORT", "5432") 
        )
    
    else:
        raise ValueError("Unsupported db type.")
    
    cursor = connection.cursor()
    sql_cmds = schema_sql.split(";")
    for sql_cmd in sql_cmds[:-1]:
        cursor.execute(sql_cmd)
        connection.commit()

    cursor.close()


if __name__ == "__main__":
    boostrap_db("postgres")