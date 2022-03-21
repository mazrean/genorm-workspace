CREATE TABLE `messages` (
  `id`  CHAR(36) PRIMARY KEY NOT NULL,
  `user_id` CHAR(36) NOT NULL,
  `content` TEXT NOT NULL,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY `idx_user` (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE = InnoDB;
