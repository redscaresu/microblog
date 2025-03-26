-- Creation of blog table
CREATE TABLE IF NOT EXISTS blog (
  blog_id uuid NOT NULL,
  blog_title TEXT NOT NULL,
  blog_post TEXT NOT NULL,
  blog_name character varying(255) NOT NULL
  PRIMARY KEY (blog_id)
);