DATABASE_NAME="test"

echo "*** CREATING DATABASE ***"

# check if db exists, if yes do nothing, otherwise import db

db_exists=$( psql -U postgres -t -c "SELECT datname FROM pg_catalog.pg_database WHERE datname = '$DATABASE_NAME';" )
if [ -z "$db_exists" ]
then
      echo "\$var is empty"
      createdb $DATABASE_NAME -U postgres
      psql -U postgres -d $DATABASE_NAME -c "GRANT ALL PRIVILEGES ON DATABASE "$DATABASE_NAME" TO postgres;"
else
      echo "DB $db_exists exists, do nothing"
fi


echo "*** DATABASE CREATED! ***"