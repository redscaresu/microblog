-- Creation of blog table
CREATE TABLE IF NOT EXISTS blog (
  blog_id uuid NOT NULL,
  blog_title varchar(10000) NOT NULL,
  blog_post varchar(10000) NOT NULL,
  PRIMARY KEY (blog_id)
);