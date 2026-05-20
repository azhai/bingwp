
-- ----------------------------
-- Table structure for t_wall_daily
-- ----------------------------
CREATE TABLE IF NOT EXISTS t_wall_daily (
  id integer NOT NULL,
  guid varchar[10] NOT NULL,
  bing_date timestamp,
  bing_sku varchar[100] NOT NULL,
  title varchar[255] NOT NULL,
  headline varchar[255] NOT NULL,
  color varchar[15] NOT NULL,
  max_dpi varchar[15] NOT NULL,
  PRIMARY KEY (id)
)
;
CREATE UNIQUE INDEX IF NOT EXISTS ON t_wall_daily (bing_date);
CREATE INDEX IF NOT EXISTS ON t_wall_daily (title);
-- COMMENT ON COLUMN t_wall_daily.guid IS 'bing.wilii.cn原始ID';
-- COMMENT ON COLUMN t_wall_daily.bing_date IS '必应的发布日期';
-- COMMENT ON COLUMN t_wall_daily.bing_sku IS '必应图片编号';
-- COMMENT ON COLUMN t_wall_daily.title IS '标题';
-- COMMENT ON COLUMN t_wall_daily.headline IS '简介';
-- COMMENT ON COLUMN t_wall_daily.color IS '主色调';
-- COMMENT ON COLUMN t_wall_daily.max_dpi IS '图片最大分辨率';
-- COMMENT ON TABLE t_wall_daily IS '每日壁纸';

-- ----------------------------
-- Table structure for t_wall_image
-- ----------------------------
CREATE TABLE IF NOT EXISTS t_wall_image (
  id integer NOT NULL,
  daily_id integer NOT NULL,
  file_name varchar[100] NOT NULL,
  img_md5 varchar[32] NOT NULL,
  img_size integer NOT NULL,
  img_offset integer NOT NULL,
  img_width integer NOT NULL,
  img_height integer NOT NULL,
  PRIMARY KEY (id)
)
;
CREATE INDEX IF NOT EXISTS ON t_wall_image (daily_id);
CREATE INDEX IF NOT EXISTS ON t_wall_image (img_md5);
CREATE INDEX IF NOT EXISTS ON t_wall_image (img_size);
-- COMMENT ON COLUMN t_wall_image.daily_id IS '墙纸ID';
-- COMMENT ON COLUMN t_wall_image.file_name IS '文件路径';
-- COMMENT ON COLUMN t_wall_image.img_md5 IS '图片MD5哈希';
-- COMMENT ON COLUMN t_wall_image.img_size IS '图片大小，单位：字节';
-- COMMENT ON COLUMN t_wall_image.img_offset IS '图片在文件中偏移';
-- COMMENT ON COLUMN t_wall_image.img_width IS '图片宽度';
-- COMMENT ON COLUMN t_wall_image.img_height IS '图片高度';
-- COMMENT ON TABLE t_wall_image IS '壁纸图片';

-- ----------------------------
-- Table structure for t_wall_note
-- ----------------------------
CREATE TABLE IF NOT EXISTS t_wall_note (
  id integer NOT NULL AUTO_INCREMENT,
  daily_id integer NOT NULL,
  note_type varchar[50] NOT NULL,
  note_chinese varchar,
  note_english varchar,
  PRIMARY KEY (id)
)
;
CREATE INDEX IF NOT EXISTS ON t_wall_note (daily_id);
-- COMMENT ON COLUMN t_wall_note.daily_id IS '墙纸ID';
-- COMMENT ON COLUMN t_wall_note.note_type IS '小知识类型';
-- COMMENT ON COLUMN t_wall_note.note_chinese IS '中文描述';
-- COMMENT ON COLUMN t_wall_note.note_english IS '英文描述';
-- COMMENT ON TABLE t_wall_note IS '墙纸小知识';
