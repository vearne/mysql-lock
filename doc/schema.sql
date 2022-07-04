-- 2021-07-22
CREATE TABLE `_lock_store` (
   `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
   `name` varchar(100) NOT NULL DEFAULT '',
   `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
   PRIMARY KEY (`id`),
   UNIQUE KEY `uni_name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;


CREATE TABLE `_lock_counter` (
   `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
   `name` varchar(100) NOT NULL DEFAULT '',
   `counter` int(10) unsigned  NOT NULL DEFAULT 0,
   `owner` varchar(100) NOT NULL DEFAULT '',
   `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
   `modified_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
   `expired_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
   PRIMARY KEY (`id`),
   UNIQUE KEY `uni_name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;