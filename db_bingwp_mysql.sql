/*
 Navicat Premium Dump SQL

 Source Server         : local
 Source Server Type    : MariaDB
 Source Server Version : 110602 (11.6.2-MariaDB)
 Source Host           : localhost:3306
 Source Schema         : db_bingwp

 Target Server Type    : MariaDB
 Target Server Version : 110602 (11.6.2-MariaDB)
 File Encoding         : 65001

 Date: 06/12/2024 11:23:30
*/

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- Table structure for t_wall_daily
-- ----------------------------
DROP TABLE IF EXISTS `t_wall_daily`;
CREATE TABLE `t_wall_daily` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `orig_id` int(10) unsigned NOT NULL DEFAULT 0 COMMENT 'bing.wilii.cn原始ID',
  `bing_date` date DEFAULT NULL COMMENT '必应的发布日期',
  `bing_sku` varchar(100) NOT NULL DEFAULT '' COMMENT '必应图片编号',
  `title` varchar(255) NOT NULL DEFAULT '' COMMENT '简介',
  `max_dpi` varchar(15) NOT NULL DEFAULT '' COMMENT '图片最大分辨率',
  PRIMARY KEY (`id`),
  UNIQUE KEY `UQE_t_wall_daily_bing_date` (`bing_date`),
  KEY `IDX_t_wall_daily_bing_sku` (`bing_sku`),
  KEY `IDX_t_wall_daily_title` (`title`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- ----------------------------
-- Table structure for t_wall_image
-- ----------------------------
DROP TABLE IF EXISTS `t_wall_image`;
CREATE TABLE `t_wall_image` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `daily_id` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '墙纸ID',
  `file_name` varchar(100) NOT NULL DEFAULT '' COMMENT '文件路径',
  `img_md5` char(32) NOT NULL DEFAULT '' COMMENT '图片MD5哈希',
  `img_size` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '图片大小（单位：字节）',
  `img_offset` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '图片在文件中偏移',
  `img_width` mediumint(6) unsigned NOT NULL DEFAULT 0 COMMENT '图片宽度',
  `img_height` mediumint(6) unsigned NOT NULL DEFAULT 0 COMMENT '图片高度',
  PRIMARY KEY (`id`),
  KEY `IDX_t_wall_image_img_size` (`img_size`),
  KEY `IDX_t_wall_image_daily_id` (`daily_id`),
  KEY `IDX_t_wall_image_img_md5` (`img_md5`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- ----------------------------
-- Table structure for t_wall_note
-- ----------------------------
DROP TABLE IF EXISTS `t_wall_note`;
CREATE TABLE `t_wall_note` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `daily_id` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '墙纸ID',
  `note_type` varchar(50) NOT NULL DEFAULT '' COMMENT '小知识类型',
  `note_chinese` mediumtext DEFAULT NULL COMMENT '中文描述',
  `note_english` mediumtext DEFAULT NULL COMMENT '英文描述',
  PRIMARY KEY (`id`),
  KEY `IDX_t_wall_note_daily_id` (`daily_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

SET FOREIGN_KEY_CHECKS = 1;
