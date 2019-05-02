DROP DATABASE IF EXISTS minipro;
CREATE DATABASE minipro;

use minipro;

DROP TABLE IF EXISTS user;
DROP TABLE IF EXISTS pair;
DROP TABLE IF EXISTS msg;

CREATE TABLE user
(
    openId varchar(40) NOT NULL,
    answer BIGINT,
    PRIMARY KEY (openId)
)ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE pair
(
    user varchar(40) NOT NULL,
    userMatchTarget varchar(40),
    PRIMARY KEY (user),
    FOREIGN KEY (user) REFERENCES user(openId),
    FOREIGN KEY (userMatchTarget) REFERENCES user(openId)
)ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE msg
(
    id INT AUTO_INCREMENT,
    _from CHAR NOT NULL,
    _to CHAR NOT NULL,
    -- createTime timestamp,
    msgType CHAR NOT NULL,
    msgContent CHAR NOT NULL,
    deleted TINYINT NOT NULL DEFAULT 0,
    PRIMARY KEY (id),
    FOREIGN KEY (_from) REFERENCES user(openId),
    FOREIGN KEY (_to) REFERENCES user(openId)
)ENGINE=InnoDB DEFAULT CHARSET=utf8;
