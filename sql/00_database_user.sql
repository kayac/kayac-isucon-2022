CREATE DATABASE IF NOT EXISTS `isucon_listen80`;
CREATE USER isucon IDENTIFIED BY 'isucon';
GRANT ALL PRIVILEGES ON isucon_listen80.* TO 'isucon'@'%';

SET PERSIST local_infile=1;
