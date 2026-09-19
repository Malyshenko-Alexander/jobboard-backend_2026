#!/bin/bash
set -e

echo "Applying auth schema..."
psql -v ON_ERROR_STOP=1 --username auth --dbname auth_db -f /schemas/auth/001_init.sql

echo "Applying applicant schema..."
psql -v ON_ERROR_STOP=1 --username applicant --dbname applicant_db -f /schemas/applicant/001_init.sql

echo "Applying employer schema..."
psql -v ON_ERROR_STOP=1 --username employer --dbname employer_db -f /schemas/employer/001_init.sql

echo "Applying vacancy schema..."
psql -v ON_ERROR_STOP=1 --username vacancy --dbname vacancy_db -f /schemas/vacancy/001_init.sql

echo "Done"
