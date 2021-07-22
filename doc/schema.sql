-- 2021-07-22
CREATE TABLE `_lock_store` (
   `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
   `name` varchar(100) NOT NULL DEFAULT '',
   `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
   PRIMARY KEY (`id`),
   UNIQUE KEY `uni_name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;