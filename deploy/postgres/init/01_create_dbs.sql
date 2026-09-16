-- Create one database (and role) per microservice.
CREATE USER auth WITH PASSWORD 'auth';
CREATE USER applicant WITH PASSWORD 'applicant';
CREATE USER employer WITH PASSWORD 'employer';
CREATE USER vacancy WITH PASSWORD 'vacancy';

CREATE DATABASE auth_db OWNER auth;
CREATE DATABASE applicant_db OWNER applicant;
CREATE DATABASE employer_db OWNER employer;
CREATE DATABASE vacancy_db OWNER vacancy;

GRANT ALL PRIVILEGES ON DATABASE auth_db TO auth;
GRANT ALL PRIVILEGES ON DATABASE applicant_db TO applicant;
GRANT ALL PRIVILEGES ON DATABASE employer_db TO employer;
GRANT ALL PRIVILEGES ON DATABASE vacancy_db TO vacancy;
