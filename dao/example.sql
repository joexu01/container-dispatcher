CREATE TABLE `role`
(
    `id`               smallint            NOT NULL PRIMARY KEY,
    `desc`             varchar(127) UNIQUE NOT NULL DEFAULT '',
    `permission_value` int UNIQUE          NOT NULL
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4 COMMENT ='角色权限表';


CREATE TABLE `user`
(
    `id`              bigint(20)          NOT NULL PRIMARY KEY AUTO_INCREMENT COMMENT '自增主键',
    `username`        varchar(255) UNIQUE NOT NULL DEFAULT '' COMMENT '用户名',
    `hashed_password` varchar(512)        NOT NULL DEFAULT '' COMMENT '加盐后密码',
    `created_at`      datetime            NOT NULL DEFAULT '1971-01-01 00:00:00' COMMENT '新增时间',
    `updated_at`      datetime            NOT NULL DEFAULT '1971-01-01 00:00:00' COMMENT '更新时间',
    `is_delete`       tinyint(4)          NOT NULL DEFAULT '0' COMMENT '是否删除',
    `email`           varchar(255)        NOT NULL DEFAULT '' COMMENT '用户邮箱',
    `user_role`       smallint            NOT NULL COMMENT '用户角色',
    CONSTRAINT `role_user_fk` FOREIGN KEY (`user_role`) REFERENCES `role`(`id`) ON UPDATE CASCADE ON DELETE SET NULL
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4 COMMENT ='用户表';

