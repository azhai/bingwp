/*
 Navicat Premium Dump SQL

 Source Server         : postgres
 Source Server Type    : PostgreSQL
 Source Server Version : 140018 (140018)
 Source Host           : 127.0.0.1:5432
 Source Catalog        : db_bingwp
 Source Schema         : public

 Target Server Type    : PostgreSQL
 Target Server Version : 140018 (140018)
 File Encoding         : 65001

 Date: 06/06/2025 01:44:39
*/


-- ----------------------------
-- Sequence structure for t_wall_note_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."t_wall_note_id_seq";
CREATE SEQUENCE "public"."t_wall_note_id_seq"
INCREMENT 1
MINVALUE  1
MAXVALUE 2147483647
START 1
CACHE 1;
ALTER SEQUENCE "public"."t_wall_note_id_seq" OWNER TO "dba";

-- ----------------------------
-- Table structure for t_wall_daily
-- ----------------------------
DROP TABLE IF EXISTS "public"."t_wall_daily";
CREATE TABLE "public"."t_wall_daily" (
  "id" int8 NOT NULL,
  "guid" char(10) COLLATE "pg_catalog"."default" NOT NULL,
  "bing_date" date,
  "bing_sku" varchar(100) COLLATE "pg_catalog"."default" NOT NULL,
  "title" varchar(255) COLLATE "pg_catalog"."default" NOT NULL,
  "headline" varchar(255) COLLATE "pg_catalog"."default" NOT NULL,
  "color" varchar(15) COLLATE "pg_catalog"."default" NOT NULL,
  "max_dpi" varchar(15) COLLATE "pg_catalog"."default" NOT NULL
)
;
ALTER TABLE "public"."t_wall_daily" OWNER TO "dba";
COMMENT ON COLUMN "public"."t_wall_daily"."guid" IS 'bing.wilii.cn原始ID';
COMMENT ON COLUMN "public"."t_wall_daily"."bing_date" IS '必应的发布日期';
COMMENT ON COLUMN "public"."t_wall_daily"."bing_sku" IS '必应图片编号';
COMMENT ON COLUMN "public"."t_wall_daily"."title" IS '标题';
COMMENT ON COLUMN "public"."t_wall_daily"."headline" IS '简介';
COMMENT ON COLUMN "public"."t_wall_daily"."color" IS '主色调';
COMMENT ON COLUMN "public"."t_wall_daily"."max_dpi" IS '图片最大分辨率';
COMMENT ON TABLE "public"."t_wall_daily" IS '每日壁纸';

-- ----------------------------
-- Table structure for t_wall_image
-- ----------------------------
DROP TABLE IF EXISTS "public"."t_wall_image";
CREATE TABLE "public"."t_wall_image" (
  "id" int8 NOT NULL,
  "daily_id" int8 NOT NULL,
  "file_name" varchar(100) COLLATE "pg_catalog"."default" NOT NULL,
  "img_md5" char(32) COLLATE "pg_catalog"."default" NOT NULL,
  "img_size" int8 NOT NULL,
  "img_offset" int8 NOT NULL,
  "img_width" int4 NOT NULL,
  "img_height" int4 NOT NULL
)
;
ALTER TABLE "public"."t_wall_image" OWNER TO "dba";
COMMENT ON COLUMN "public"."t_wall_image"."daily_id" IS '墙纸ID';
COMMENT ON COLUMN "public"."t_wall_image"."file_name" IS '文件路径';
COMMENT ON COLUMN "public"."t_wall_image"."img_md5" IS '图片MD5哈希';
COMMENT ON COLUMN "public"."t_wall_image"."img_size" IS '图片大小，单位：字节';
COMMENT ON COLUMN "public"."t_wall_image"."img_offset" IS '图片在文件中偏移';
COMMENT ON COLUMN "public"."t_wall_image"."img_width" IS '图片宽度';
COMMENT ON COLUMN "public"."t_wall_image"."img_height" IS '图片高度';
COMMENT ON TABLE "public"."t_wall_image" IS '壁纸图片';

-- ----------------------------
-- Table structure for t_wall_note
-- ----------------------------
DROP TABLE IF EXISTS "public"."t_wall_note";
CREATE TABLE "public"."t_wall_note" (
  "id" int8 NOT NULL DEFAULT nextval('t_wall_note_id_seq'::regclass),
  "daily_id" int8 NOT NULL,
  "note_type" varchar(50) COLLATE "pg_catalog"."default" NOT NULL,
  "note_chinese" text COLLATE "pg_catalog"."default",
  "note_english" text COLLATE "pg_catalog"."default"
)
;
ALTER TABLE "public"."t_wall_note" OWNER TO "dba";
COMMENT ON COLUMN "public"."t_wall_note"."daily_id" IS '墙纸ID';
COMMENT ON COLUMN "public"."t_wall_note"."note_type" IS '小知识类型';
COMMENT ON COLUMN "public"."t_wall_note"."note_chinese" IS '中文描述';
COMMENT ON COLUMN "public"."t_wall_note"."note_english" IS '英文描述';
COMMENT ON TABLE "public"."t_wall_note" IS '墙纸小知识';

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "public"."t_wall_note_id_seq"
OWNED BY "public"."t_wall_note"."id";
SELECT setval('"public"."t_wall_note_id_seq"', 11113, true);

-- ----------------------------
-- Indexes structure for table t_wall_daily
-- ----------------------------
CREATE INDEX "IDX_t_wall_daily_bing_sku" ON "public"."t_wall_daily" USING btree (
  "bing_sku" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST
);
CREATE INDEX "IDX_t_wall_daily_title" ON "public"."t_wall_daily" USING btree (
  "title" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST
);

-- ----------------------------
-- Uniques structure for table t_wall_daily
-- ----------------------------
ALTER TABLE "public"."t_wall_daily" ADD CONSTRAINT "t_wall_daily_bing_date_key" UNIQUE ("bing_date");

-- ----------------------------
-- Primary Key structure for table t_wall_daily
-- ----------------------------
ALTER TABLE "public"."t_wall_daily" ADD CONSTRAINT "t_wall_daily_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Indexes structure for table t_wall_image
-- ----------------------------
CREATE INDEX "IDX_t_wall_image_daily_id" ON "public"."t_wall_image" USING btree (
  "daily_id" "pg_catalog"."int8_ops" ASC NULLS LAST
);
CREATE INDEX "IDX_t_wall_image_img_md5" ON "public"."t_wall_image" USING btree (
  "img_md5" COLLATE "pg_catalog"."default" "pg_catalog"."bpchar_ops" ASC NULLS LAST
);
CREATE INDEX "IDX_t_wall_image_img_size" ON "public"."t_wall_image" USING btree (
  "img_size" "pg_catalog"."int8_ops" ASC NULLS LAST
);

-- ----------------------------
-- Primary Key structure for table t_wall_image
-- ----------------------------
ALTER TABLE "public"."t_wall_image" ADD CONSTRAINT "t_wall_image_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Indexes structure for table t_wall_note
-- ----------------------------
CREATE INDEX "t_wall_note_daily_id_idx" ON "public"."t_wall_note" USING btree (
  "daily_id" "pg_catalog"."int8_ops" ASC NULLS LAST
);

-- ----------------------------
-- Primary Key structure for table t_wall_note
-- ----------------------------
ALTER TABLE "public"."t_wall_note" ADD CONSTRAINT "t_wall_note_pkey" PRIMARY KEY ("id");
