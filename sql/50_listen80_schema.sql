use isucon_listen80

CREATE TABLE `user` (
  `account` VARCHAR(191) NOT NULL,
  `display_name` VARCHAR(191) NOT NULL,
  `password_hash` VARCHAR(191) NOT NULL,
  `is_ban` TINYINT(2) NOT NULL,
  `created_at` TIMESTAMP(3) NOT NULL,
  `last_logined_at` TIMESTAMP(3) NOT NULL,
  PRIMARY KEY (`account`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `song` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `ulid` VARCHAR(191) NOT NULL,
  `title` VARCHAR(191) NOT NULL,
  `artist_id` BIGINT NOT NULL,
  `album`  VARCHAR(191) NOT NULL,
  `track_number` INT NOT NULL,
  `is_public` TINYINT(2) NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `artist` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `ulid` VARCHAR(191) NOT NULL,
  `name` VARCHAR(191) NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `playlist` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `ulid` VARCHAR(191) NOT NULL,
  `name` VARCHAR(191) NOT NULL,
  `user_account` VARCHAR(191) NOT NULL,
  `is_public` TINYINT(2) NOT NULL,
  `created_at` TIMESTAMP(3) NOT NULL,
  `updated_at` TIMESTAMP(3) NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `playlist_song` (
  `playlist_id` BIGINT NOT NULL,
  `sort_order` INT NOT NULL,
  `song_id` BIGINT NOT NULL,
  PRIMARY KEY (`playlist_id`, `sort_order`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `playlist_favorite` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `playlist_id` BIGINT NOT NULL,
  `favorite_user_account` VARCHAR(191) NOT NULL,
  `created_at` TIMESTAMP(3) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE `uniq_playlist_id_favorite_user_account` (`playlist_id`, `favorite_user_account`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `sessions` (
  `session_id` varchar(128) COLLATE utf8mb4_bin NOT NULL,
  `expires` int(11) unsigned NOT NULL,
  `data` mediumtext COLLATE utf8mb4_bin,
  PRIMARY KEY (`session_id`)
) ENGINE=InnoDB
