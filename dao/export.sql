-- `container-dev`.`role` definition

CREATE TABLE `role` (
                        `id` smallint(6) NOT NULL,
                        `desc` varchar(127) NOT NULL DEFAULT '',
                        `permission_value` int(11) NOT NULL,
                        PRIMARY KEY (`id`),
                        UNIQUE KEY `desc` (`desc`),
                        UNIQUE KEY `permission_value` (`permission_value`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='角色权限表';


-- `container-dev`.`user` definition

CREATE TABLE `user` (
                        `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '自增主键',
                        `username` varchar(255) NOT NULL DEFAULT '' COMMENT '用户名',
                        `hashed_password` varchar(512) NOT NULL DEFAULT '' COMMENT '加盐后密码',
                        `created_at` datetime NOT NULL DEFAULT '1971-01-01 00:00:00' COMMENT '新增时间',
                        `updated_at` datetime NOT NULL DEFAULT '1971-01-01 00:00:00' COMMENT '更新时间',
                        `is_delete` tinyint(4) NOT NULL DEFAULT 0 COMMENT '是否删除',
                        `email` varchar(255) NOT NULL DEFAULT '' COMMENT '用户邮箱',
                        `user_role` smallint(6) NOT NULL COMMENT '用户角色',
                        PRIMARY KEY (`id`),
                        UNIQUE KEY `username` (`username`),
                        KEY `role_user_fk` (`user_role`),
                        CONSTRAINT `role_user_fk` FOREIGN KEY (`user_role`) REFERENCES `role` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=12 DEFAULT CHARSET=utf8mb4 COMMENT='用户表';

-- `container-dev`.`algorithm` definition

CREATE TABLE `algorithm` (
                             `path` varchar(1024) NOT NULL COMMENT ' 攻击算法的路径',
                             `entry_point` varchar(255) NOT NULL,
                             `author_id` bigint(20) DEFAULT NULL,
                             `exec_binary` varchar(16) DEFAULT NULL COMMENT 'sh or py2 or py3 or ...',
                             `name` varchar(100) NOT NULL,
                             `desc` varchar(512) DEFAULT NULL,
                             `uuid` varchar(256) NOT NULL,
                             `files` text NOT NULL COMMENT '文件列表（; seperated）',
                             `created_at` datetime DEFAULT NULL,
                             `default_image_name` varchar(256) NOT NULL,
                             PRIMARY KEY (`uuid`),
                             KEY `algorithm_FK` (`author_id`),
                             CONSTRAINT `algorithm_FK` FOREIGN KEY (`author_id`) REFERENCES `user` (`id`) ON DELETE SET NULL ON UPDATE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='攻击算法';


-- `container-dev`.container_user definition

CREATE TABLE `container_user` (
                                  `user_id` bigint(20) NOT NULL,
                                  `container_id` varchar(255) NOT NULL,
                                  KEY `container_user_FK` (`user_id`),
                                  CONSTRAINT `container_user_FK` FOREIGN KEY (`user_id`) REFERENCES `user` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户-容器表';


-- `container-dev`.task definition

CREATE TABLE `task` (
                        `uuid` varchar(256) NOT NULL,
                        `task_name` varchar(100) NOT NULL,
                        `task_desc` varchar(512) DEFAULT NULL,
                        `algorithm_uuid` varchar(256) NOT NULL,
                        `user_id` bigint(20) NOT NULL,
                        `uploaded_files` varchar(100) DEFAULT NULL,
                        `status` varchar(64) NOT NULL,
                        `created_at` datetime DEFAULT NULL,
                        `image_name` varchar(128) NOT NULL,
                        `container_id` varchar(128) DEFAULT NULL,
                        PRIMARY KEY (`uuid`),
                        KEY `task_FK_1` (`user_id`),
                        KEY `task_FK` (`algorithm_uuid`),
                        CONSTRAINT `task_FK` FOREIGN KEY (`algorithm_uuid`) REFERENCES `algorithm` (`uuid`),
                        CONSTRAINT `task_FK_1` FOREIGN KEY (`user_id`) REFERENCES `user` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='机器学习任务';