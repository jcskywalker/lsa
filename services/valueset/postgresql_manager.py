import psycopg
from configparser import ConfigParser

'''
This method will use the connection data saved in configuration file to get postgresql database server connection.

config_file_path : Is the configuration file saved path, the configuration file is database.ini in this example, and it is saved in the same path of PostgresqlManager.py file.

section_name : This is the section name in above configuration file. The options in this section record the postgresql database server connection info.

'''

class PostgresqlManager:
    def get_connection_by_config(self, config_file_path, section_name):
        if(len(config_file_path) > 0 and len(section_name) > 0):
            # Create an instance of ConfigParser class.
            config_parser = ConfigParser()
            # read the configuration file.
            config_parser.read(config_file_path)
            # if the configuration file contains the provided section name.
            if(config_parser.has_section(section_name)):
                # read the options of the section. the config_params is a list object.
                config_params = config_parser.items(section_name)
                # so we need below code to convert the list object to a python dictionary object.
                # define an empty dictionary.
                db_conn_dict = {}
                # loop in the list.
                for config_param in config_params:
                    # get options key and value.
                    key = config_param[0]
                    value = config_param[1]
                    # add the key value pair in the dictionary object.
                    db_conn_dict[key] = value
                # get connection object use above dictionary object.
                # conn = psycopg.connect(**db_conn_dict)
                result = " ".join(str(key + "=") + str(value) for key, value in db_conn_dict.items())
                print(result)
                conn = psycopg.connect(result)
                self._conn = conn
                print("******* get postgresql database connection with configuration file ********", "\n")

    def close_connection(self):
        if self._cursor is not None:
            self._cursor.close()
        if self._conn is not None:
            self._conn.close()
        print("******* close postgresql database connection ********", "\n")
    # get db cursor object.
    def get_cursor(self):
        if self._conn is not None:
            if not hasattr(self, '_cursor') or self._cursor is None or self._cursor.closed:
               # save the db cursor object in private instance variable.
               self._cursor = self._conn.cursor()
        return self._cursor
    # execute select sql command.
    def execute_sql(self, sql):
        self.get_cursor()
        self._cursor.execute(sql)
        # get the sql execution result.
        result = self._cursor.fetchone()
        print("Record is : ", result, "\n")
        return result 

    
