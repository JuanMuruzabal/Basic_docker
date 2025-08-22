-- Create database if not exists
CREATE DATABASE IF NOT EXISTS appdb_qa;
CREATE DATABASE IF NOT EXISTS appdb_prod;

-- Create application user with proper authentication (with existence check)
DROP USER IF EXISTS 'appuser'@'%';
CREATE USER IF NOT EXISTS 'appuser'@'%' IDENTIFIED WITH mysql_native_password BY 'apppass';

GRANT ALL PRIVILEGES ON appdb_qa.* TO 'appuser'@'%';
GRANT ALL PRIVILEGES ON appdb_prod.* TO 'appuser'@'%';
FLUSH PRIVILEGES;
