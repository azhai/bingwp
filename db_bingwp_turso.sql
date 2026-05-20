/*
 Navicat Premium Dump SQL

 Source Server         : turso
 Source Server Type    : SQLite
 Source Server Version : 3045000 (3.45.0)
 Source Schema         : main

 Target Server Type    : SQLite
 Target Server Version : 3045000 (3.45.0)
 File Encoding         : 65001

 Date: 25/09/2025 11:57:00
*/

PRAGMA foreign_keys = false;

-- ----------------------------
-- Table structure for t_wall_daily
-- ----------------------------
DROP TABLE IF EXISTS "t_wall_daily";
CREATE TABLE "t_wall_daily" (
  "id" integer(64) NOT NULL,
  "guid" text(10) NOT NULL,
  "bing_date" text,
  "bing_sku" text(100) NOT NULL,
  "title" text(255) NOT NULL,
  "headline" text(255) NOT NULL,
  "color" text(15) NOT NULL,
  "max_dpi" text(15) NOT NULL,
  CONSTRAINT "t_wall_daily_pkey" PRIMARY KEY ("id")
);

-- ----------------------------
-- Table structure for t_wall_image
-- ----------------------------
DROP TABLE IF EXISTS "t_wall_image";
CREATE TABLE "t_wall_image" (
  "id" integer(64) NOT NULL,
  "daily_id" integer(64) NOT NULL,
  "file_name" text(100) NOT NULL,
  "img_md5" text(32) NOT NULL,
  "img_size" integer(64) NOT NULL,
  "img_offset" integer(64) NOT NULL,
  "img_width" integer(32) NOT NULL,
  "img_height" integer(32) NOT NULL,
  CONSTRAINT "t_wall_image_pkey" PRIMARY KEY ("id")
);

-- ----------------------------
-- Table structure for t_wall_note
-- ----------------------------
DROP TABLE IF EXISTS "t_wall_note";
CREATE TABLE "t_wall_note" (
  "id" integer(64) NOT NULL,
  "daily_id" integer(64) NOT NULL,
  "note_type" text(50) NOT NULL,
  "note_chinese" text,
  "note_english" text,
  CONSTRAINT "t_wall_note_pkey" PRIMARY KEY ("id")
);

-- ----------------------------
-- Indexes structure for table t_wall_daily
-- ----------------------------
CREATE INDEX "main"."IDX_t_wall_daily_bing_sku"
ON "t_wall_daily" (
  "bing_sku" ASC
);
CREATE INDEX "main"."IDX_t_wall_daily_title"
ON "t_wall_daily" (
  "title" ASC
);

-- ----------------------------
-- Indexes structure for table t_wall_image
-- ----------------------------
CREATE INDEX "main"."IDX_t_wall_image_daily_id"
ON "t_wall_image" (
  "daily_id" ASC
);
CREATE INDEX "main"."IDX_t_wall_image_img_md5"
ON "t_wall_image" (
  "img_md5" ASC
);
CREATE INDEX "main"."IDX_t_wall_image_img_size"
ON "t_wall_image" (
  "img_size" ASC
);

-- ----------------------------
-- Indexes structure for table t_wall_note
-- ----------------------------
CREATE INDEX "main"."t_wall_note_daily_id_idx"
ON "t_wall_note" (
  "daily_id" ASC
);

PRAGMA foreign_keys = true;
